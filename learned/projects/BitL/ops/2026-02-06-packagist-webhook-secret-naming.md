---
name: packagist-webhook-secret-naming
confidence: MEDIUM
score: 0.68
category: ops
domain: backend
project: bitl
first_seen: 2026-02-06
last_seen: 2026-02-06
times_seen: 1
---

# Packagist Auto-Update API Requires Specific Secret Names

## Problem / Trigger
GitHub Actions workflow for Packagist auto-update succeeds overall but the Packagist API step fails with 'Missing or invalid username/apiToken in request' even when secrets appear configured

## Solution
Ensure GitHub repo secrets are named exactly PACKAGIST_USERNAME and PACKAGIST_API_TOKEN (not variations like PACKAGIST_USER or PACKAGIST_TOKEN), and that the API token is generated from Packagist account settings (not the password)

## Verification
Re-run the GitHub Actions workflow and confirm Packagist step returns success status, then verify package version updates on packagist.org

## Source
Session: 2026-02-06
