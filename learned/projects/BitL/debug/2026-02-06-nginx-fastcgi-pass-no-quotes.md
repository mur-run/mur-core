---
name: nginx-fastcgi-pass-no-quotes
confidence: HIGH
score: 0.8
category: debug
domain: devops
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# nginx fastcgi_pass unix: Directive Cannot Have Quoted Paths

## Problem / Trigger
nginx config with spaces in paths fails to parse, but quoting all paths breaks fastcgi_pass

## Solution
Quote paths for error_log, pid, access_log, root, ssl_*, include directives. Do NOT quote fastcgi_pass unix: paths — move socket files to a path without spaces (e.g., /tmp/bitl/)

## Verification
nginx -t passes, PHP requests proxy correctly

## Source
Session: 2026-02-01
