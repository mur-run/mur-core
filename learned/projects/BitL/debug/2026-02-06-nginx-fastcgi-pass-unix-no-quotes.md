---
name: nginx-fastcgi-pass-unix-no-quotes
confidence: HIGH
score: 0.82
category: debug
domain: devops
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# nginx fastcgi_pass unix: Directive Rejects Quoted Paths

## Problem / Trigger
Paths with spaces in nginx config cause parse errors, but quoting all directives uniformly breaks fastcgi_pass

## Solution
Quote paths in error_log, pid, access_log, root, ssl_*, and include directives. But `fastcgi_pass unix:` does NOT accept quotes. Move socket paths to a space-free directory like `/tmp/bitl/` instead.

## Verification
Run `nginx -t` with the config; fastcgi_pass with quoted path will fail, unquoted space-free path will pass

## Source
Session: 2026-02-01
