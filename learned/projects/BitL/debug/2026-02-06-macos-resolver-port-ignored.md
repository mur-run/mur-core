---
name: macos-resolver-port-ignored
confidence: HIGH
score: 0.9
category: debug
domain: devops
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# macOS /etc/resolver Ignores Custom DNS Ports

## Problem / Trigger
DNS resolution for custom TLD doesn't work when /etc/resolver/test specifies port 15353 with nameserver 127.0.0.1

## Solution
Run dnsmasq on port 53 (standard DNS port) and omit the port directive from resolver file. Port 53 requires root — use osascript with administrator privileges for start/stop.

## Verification
scutil --dns shows resolver active, dig @127.0.0.1 mysite.test resolves

## Source
Session: 2026-02-01
