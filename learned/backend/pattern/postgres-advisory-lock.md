---
name: postgres-advisory-lock
confidence: HIGH
score: 0.85
category: pattern
domain: backend
first_seen: 2025-01-20
last_seen: 2025-02-01
times_seen: 3
---

# PostgreSQL Advisory Lock 防止重複執行

## Problem / Trigger
多個 worker 同時處理同一個 job，造成重複執行或 race condition。

## Solution
用 `pg_try_advisory_lock(key)` 而不是 SELECT FOR UPDATE — 不需要 transaction，更輕量。
```sql
SELECT pg_try_advisory_lock(hashtext('job:' || job_id));
-- returns true if lock acquired, false if already locked
-- release with pg_advisory_unlock() or auto-release on disconnect
```
適合 cron job、queue worker 等場景。

## Verification
兩個 worker 同時跑，只有一個拿到 lock，另一個 gracefully skip。

## Source
Queue processing optimization, 2025-01-20
