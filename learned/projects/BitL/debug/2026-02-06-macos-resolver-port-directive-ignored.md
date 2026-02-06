---
name: macos-resolver-port-directive-ignored
confidence: HIGH
score: 0.8
category: debug
domain: _global
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# macOS /etc/resolver Files Ignore Custom Port on 127.0.0.1

## Problem / Trigger
macOS DNS resolver files in /etc/resolver/ with 'port 15353' directive pointing to 127.0.0.1 are silently ignored — custom port never gets used for resolution

## Solution
Run dnsmasq on standard port 53 instead of a custom port, and omit the port directive from the resolver file. Port 53 requires sudo — use osascript 'with administrator privileges' for privilege escalation

## Verification
Create a .test domain, confirm dig and browser resolution work without specifying port

## Source
Session: 2026-02-01
