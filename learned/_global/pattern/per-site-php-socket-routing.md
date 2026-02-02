---
name: per-site-php-socket-routing
confidence: HIGH
score: 0.85
category: pattern
domain: backend
first_seen: 2026-02-02
last_seen: 2026-02-02
times_seen: 3
---

# Per-Site PHP-FPM Socket Routing

## Problem / Trigger
不同 Laravel 站點需要不同 PHP 版本，但 nginx 都指向同一個 php-fpm socket。

## Solution
每個 PHP 版本啟動獨立的 FPM pool，使用版本化的 socket 路徑：
- PHP 8.2 → `/tmp/bitl-php-fpm-8.2.sock`
- PHP 8.4 → `/tmp/bitl-php-fpm-8.4.sock`

Site model 存 `phpVersion: String?`，nginx config generator 根據 site 的 phpVersion 選對應 socket：
```swift
func resolveSocket(for site: Site, defaultPHPVersion: String) -> String {
    let version = site.phpVersion ?? defaultPHPVersion
    return "\(tmpDir)/php-fpm-\(version).sock"
}
```

## Verification
多個站點同時跑不同 PHP 版本，nginx 正確路由到對應的 FPM pool。
