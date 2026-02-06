---
name: tmp-dir-pre-creation-for-sockets
confidence: HIGH
score: 0.65
category: ops
domain: devops
project: bitl
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Pre-create Socket Directories Before Starting Services

## Problem / Trigger
php-fpm and MySQL fail to start because the socket directory (/tmp/bitl/) doesn't exist. Services silently fail or produce cryptic errors about socket binding.

## Solution
Ensure the socket directory exists (e.g., `/tmp/bitl/`) in `ServiceManager.start()` before launching any service that uses Unix sockets. Create the directory with appropriate permissions as a pre-start step.

## Verification
Reboot the machine (clearing /tmp), then start php-fpm and MySQL. Confirm both start successfully with sockets created in /tmp/bitl/.

## Source
Session: 2026-02-05
