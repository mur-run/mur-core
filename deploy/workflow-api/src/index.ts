// Security model: "URL as secret" (like GitHub secret gists)
// - Session keys are cryptographically random UUIDs
// - Knowing the key is required to access the data
// - Sessions expire after 24 hours
// - Rate limiting prevents abuse (10 req/min per IP)
// - Max upload size: 5MB

interface Env {
	MUR_SESSIONS: R2Bucket;
	CORS_ORIGIN: string;
}

const MAX_BODY_SIZE = 5 * 1024 * 1024; // 5MB
const RATE_LIMIT_WINDOW = 60_000; // 1 minute
const RATE_LIMIT_MAX = 10;
const SESSION_TTL = 24 * 60 * 60; // 24 hours in seconds

// Simple in-memory rate limiter (resets on worker restart, fine for this use case)
// TODO: In-memory rate limiting resets on worker restart.
// For higher traffic, consider Cloudflare Durable Objects or KV-based rate limiting.
const rateLimitMap = new Map<string, { count: number; resetAt: number }>();

function checkRateLimit(ip: string): boolean {
	const now = Date.now();
	const entry = rateLimitMap.get(ip);

	if (!entry || now > entry.resetAt) {
		rateLimitMap.set(ip, { count: 1, resetAt: now + RATE_LIMIT_WINDOW });
		return true;
	}

	if (entry.count >= RATE_LIMIT_MAX) {
		return false;
	}

	entry.count++;
	return true;
}

function corsHeaders(origin: string): Record<string, string> {
	return {
		"Access-Control-Allow-Origin": origin,
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Max-Age": "86400",
	};
}

function jsonResponse(
	data: unknown,
	status: number,
	corsOrigin: string
): Response {
	return new Response(JSON.stringify(data), {
		status,
		headers: {
			"Content-Type": "application/json",
			...corsHeaders(corsOrigin),
		},
	});
}

function generateKey(): string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789";
	const segments = [8, 4, 4];
	return segments
		.map((len) =>
			Array.from(crypto.getRandomValues(new Uint8Array(len)))
				.map((b) => chars[b % chars.length])
				.join("")
		)
		.join("-");
}

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		const url = new URL(request.url);
		const corsOrigin = env.CORS_ORIGIN || "https://workflow.mur.run";

		// Handle CORS preflight
		if (request.method === "OPTIONS") {
			return new Response(null, {
				status: 204,
				headers: corsHeaders(corsOrigin),
			});
		}

		// Rate limit check
		const ip =
			request.headers.get("CF-Connecting-IP") || "unknown";
		if (!checkRateLimit(ip)) {
			return jsonResponse(
				{ error: "rate limit exceeded, try again later" },
				429,
				corsOrigin
			);
		}

		// Route: POST /upload
		if (url.pathname === "/upload" && request.method === "POST") {
			return handleUpload(request, env, corsOrigin);
		}

		// Route: POST /session/:key/update
		const updateMatch = url.pathname.match(
			/^\/session\/([a-z0-9-]+)\/update$/
		);
		if (updateMatch && request.method === "POST") {
			return handleUpdateSession(
				updateMatch[1],
				request,
				env,
				corsOrigin
			);
		}

		// Route: GET /session/:key/export
		const exportMatch = url.pathname.match(
			/^\/session\/([a-z0-9-]+)\/export$/
		);
		if (exportMatch && request.method === "GET") {
			return handleExportSession(exportMatch[1], env, corsOrigin);
		}

		// Route: GET /session/:key
		const sessionMatch = url.pathname.match(/^\/session\/([a-z0-9-]+)$/);
		if (sessionMatch && request.method === "GET") {
			return handleGetSession(sessionMatch[1], env, corsOrigin);
		}

		return jsonResponse({ error: "not found" }, 404, corsOrigin);
	},
};

async function handleUpload(
	request: Request,
	env: Env,
	corsOrigin: string
): Promise<Response> {
	// Check content length
	const contentLength = request.headers.get("Content-Length");
	if (contentLength && parseInt(contentLength) > MAX_BODY_SIZE) {
		return jsonResponse(
			{ error: "body too large, max 5MB" },
			413,
			corsOrigin
		);
	}

	const body = await request.arrayBuffer();
	if (body.byteLength > MAX_BODY_SIZE) {
		return jsonResponse(
			{ error: "body too large, max 5MB" },
			413,
			corsOrigin
		);
	}

	if (body.byteLength === 0) {
		return jsonResponse({ error: "empty body" }, 400, corsOrigin);
	}

	const key = generateKey();
	const expiresAt = new Date(Date.now() + SESSION_TTL * 1000).toISOString();

	await env.MUR_SESSIONS.put(key, body, {
		httpMetadata: {
			contentType: "application/json",
			contentEncoding: "gzip",
		},
		customMetadata: {
			expires_at: expiresAt,
			uploaded_at: new Date().toISOString(),
		},
	});

	return jsonResponse(
		{
			url: `https://workflow.mur.run/?s=${key}`,
			key,
			expires_at: expiresAt,
		},
		201,
		corsOrigin
	);
}

async function handleGetSession(
	key: string,
	env: Env,
	corsOrigin: string
): Promise<Response> {
	const object = await env.MUR_SESSIONS.get(key);
	if (!object) {
		return jsonResponse(
			{ error: "session not found or expired" },
			404,
			corsOrigin
		);
	}

	// Check expiry from metadata
	const expiresAt = object.customMetadata?.expires_at;
	if (expiresAt && new Date(expiresAt) < new Date()) {
		// Clean up expired object
		await env.MUR_SESSIONS.delete(key);
		return jsonResponse(
			{ error: "session expired" },
			410,
			corsOrigin
		);
	}

	const data = await object.arrayBuffer();

	return new Response(data, {
		status: 200,
		headers: {
			"Content-Type": "application/json",
			"Content-Encoding": "gzip",
			"Cache-Control": "public, max-age=3600",
			...corsHeaders(corsOrigin),
		},
	});
}

async function readSessionJson(
	key: string,
	env: Env
): Promise<{
	json: Record<string, unknown>;
	customMetadata: Record<string, string>;
} | null> {
	const object = await env.MUR_SESSIONS.get(key);
	if (!object) return null;

	const expiresAt = object.customMetadata?.expires_at;
	if (expiresAt && new Date(expiresAt) < new Date()) {
		await env.MUR_SESSIONS.delete(key);
		return null;
	}

	const raw = await object.arrayBuffer();
	let text: string;

	// Data may be gzip-compressed — try decompressing, fall back to raw
	try {
		const ds = new DecompressionStream("gzip");
		const stream = new Blob([raw]).stream().pipeThrough(ds);
		const decompressed = await new Response(stream).arrayBuffer();
		text = new TextDecoder().decode(decompressed);
	} catch {
		text = new TextDecoder().decode(raw);
	}

	return {
		json: JSON.parse(text),
		customMetadata: object.customMetadata || {},
	};
}

async function compressJson(data: unknown): Promise<ArrayBuffer> {
	const jsonBytes = new TextEncoder().encode(JSON.stringify(data));
	const cs = new CompressionStream("gzip");
	const stream = new Blob([jsonBytes]).stream().pipeThrough(cs);
	return new Response(stream).arrayBuffer();
}

async function handleUpdateSession(
	key: string,
	request: Request,
	env: Env,
	corsOrigin: string
): Promise<Response> {
	const contentLength = request.headers.get("Content-Length");
	if (contentLength && parseInt(contentLength) > MAX_BODY_SIZE) {
		return jsonResponse(
			{ error: "body too large, max 5MB" },
			413,
			corsOrigin
		);
	}

	const session = await readSessionJson(key, env);
	if (!session) {
		return jsonResponse(
			{ error: "session not found or expired" },
			404,
			corsOrigin
		);
	}

	let body: { patterns?: unknown[] };
	try {
		body = (await request.json()) as { patterns?: unknown[] };
	} catch {
		return jsonResponse(
			{ error: "invalid JSON body" },
			400,
			corsOrigin
		);
	}

	if (!body.patterns || !Array.isArray(body.patterns)) {
		return jsonResponse(
			{ error: "invalid body: patterns array required" },
			400,
			corsOrigin
		);
	}

	// Merge updated patterns into existing session data
	const updated = { ...session.json } as Record<string, unknown>;
	if (
		updated.session &&
		typeof updated.session === "object" &&
		(updated.session as Record<string, unknown>).patterns
	) {
		(updated.session as Record<string, unknown>).patterns =
			body.patterns;
	} else {
		updated.patterns = body.patterns;
	}

	const compressed = await compressJson(updated);

	await env.MUR_SESSIONS.put(key, compressed, {
		httpMetadata: {
			contentType: "application/json",
			contentEncoding: "gzip",
		},
		customMetadata: {
			...session.customMetadata,
			updated_at: new Date().toISOString(),
		},
	});

	return jsonResponse(
		{ ok: true, updated: body.patterns.length },
		200,
		corsOrigin
	);
}

async function handleExportSession(
	key: string,
	env: Env,
	corsOrigin: string
): Promise<Response> {
	const session = await readSessionJson(key, env);
	if (!session) {
		return jsonResponse(
			{ error: "session not found or expired" },
			404,
			corsOrigin
		);
	}

	const data = session.json as Record<string, unknown>;
	const rawPatterns =
		(data.patterns as unknown[]) ||
		((data.session as Record<string, unknown>)
			?.patterns as unknown[]) ||
		[];

	// Return mur-compatible format: clean JSON array
	const exported = rawPatterns.map((p: unknown) => {
		const pat = p as Record<string, unknown>;
		return {
			name: pat.name || "",
			category: pat.category || "uncategorized",
			domain: pat.domain || "",
			description: pat.description || "",
			content: pat.content || pat.description || "",
			confidence: pat.confidence || 0,
			tags: pat.tags || [],
		};
	});

	return jsonResponse(exported, 200, corsOrigin);
}
