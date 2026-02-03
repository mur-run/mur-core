# Change Spec: 15 — Web Dashboard

## Summary
Add a local web dashboard to visualize stats and patterns via `mur serve`.

## Technical Stack

### Backend
- **Go HTTP server** — using `net/http` (stdlib, no external deps)
- **API format** — JSON REST
- **Port** — default `:8383` (configurable via `--port`)

### Frontend
- **HTML + inline CSS/JS** — embedded in Go binary via `embed`
- **No build step** — simple, single-file approach
- **Minimal JS** — vanilla JavaScript for fetching data, no framework
- **CSS** — system font stack, CSS variables for theming

### Deployment
- **Local only** — `mur serve` for development/personal use
- **No cloud deployment** for v1

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/stats` | Returns summary statistics |
| GET | `/api/stats/daily` | Returns daily usage breakdown |
| GET | `/api/patterns` | Returns all patterns (optional `?domain=`, `?category=` filters) |
| GET | `/api/patterns/:name` | Returns single pattern by name |
| POST | `/api/sync` | Triggers sync to AI CLI tools |
| GET | `/api/config` | Returns current config |
| GET | `/api/health` | Health check for all tools |

### Response Format
```json
{
  "success": true,
  "data": { ... },
  "error": null
}
```

## Dashboard Pages

### Overview (`/`)
- Total runs, estimated cost, estimated saved
- 7-day usage trend chart (simple bar chart)
- Tool distribution pie/bar
- Auto-routing stats (free vs paid ratio)

### Patterns (`/patterns`)
- List all patterns with domain/category badges
- Search/filter by name, domain, category
- Click to expand pattern content
- Quick edit/delete actions (future)

### Tools (`/tools`)
- Tool health status (installed? working?)
- Tool config overview
- Quick links to tool docs

### Settings (`/settings`)
- View current config (read-only for v1)
- Links to config file location
- Sync status

## File Structure

```
cmd/mur/cmd/
  serve.go           # mur serve command

internal/server/
  server.go          # HTTP server setup
  handlers.go        # API handlers
  static.go          # Embedded HTML/CSS/JS

internal/server/static/
  index.html         # Main dashboard page
  style.css          # Styles (optional, could be inline)
```

## Implementation Notes

1. **Embedded assets**: Use `//go:embed` for static files
2. **CORS**: Not needed (same-origin)
3. **Auth**: None for v1 (local only)
4. **Hot reload**: None for v1 (restart server)

## Future Enhancements (Out of Scope)
- Pattern editing from dashboard
- Config editing from dashboard
- Real-time updates via WebSocket
- Dark mode toggle
- Export to PDF/CSV
- Mobile-responsive design improvements

## Dependencies
- None new (stdlib only)

## Commands
```bash
mur serve              # Start on :8383
mur serve --port 9000  # Custom port
```

## Status
- [x] Spec created
- [ ] API handlers implemented
- [ ] Static HTML embedded
- [ ] serve command added
- [ ] Build tested
