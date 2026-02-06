---
name: mysql9-deprecated-innodb-log-file-size
confidence: HIGH
score: 0.75
category: ops
domain: backend
project: BitL
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# MySQL 9.x Removes innodb_log_file_size Configuration

## Problem / Trigger
MySQL 9.x fails to start or throws warnings when config contains `innodb_log_file_size` parameter

## Solution
Remove `innodb_log_file_size = 48M` (or any value) from MySQL configuration files when targeting MySQL 9.x+

## Verification
MySQL 9.x starts without deprecation warnings or errors related to InnoDB configuration

## Source
Session: 2026-02-05
