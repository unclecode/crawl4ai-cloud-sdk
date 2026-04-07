# BrowserConfig Reference

Pass as `browser_config` JSON object in any endpoint request body.
Only applies when `strategy="browser"`. Ignored for `http` strategy.

---

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `headers` | object | `null` | Custom HTTP headers |
| `cookies` | object[] | `null` | Array of cookie objects |
| `viewport_width` | int | `1920` | Browser viewport width (px) |
| `viewport_height` | int | `1080` | Browser viewport height (px) |
| `user_agent` | string | `null` | Custom User-Agent string |
| `user_agent_mode` | `"random"\|"fixed"` | `"random"` | Rotate or fix UA |
| `profile_id` | string | `null` | Saved browser profile UUID |
| `ignore_https_errors` | bool | `true` | Skip SSL cert errors |
| `java_script_enabled` | bool | `true` | Enable/disable JS |
| `text_mode` | bool | `false` | Disable images/CSS for speed |
| `light_mode` | bool | `false` | Minimal resource loading |

---

## Common Patterns

**Custom headers (language, auth token):**
```json
{"browser_config": {
  "headers": {"Accept-Language": "fr-FR", "Authorization": "Bearer tok_xxx"}
}}
```

**Cookie authentication:**
```json
{"browser_config": {
  "cookies": [
    {"name": "session_id", "value": "abc123", "domain": ".example.com", "path": "/"},
    {"name": "csrf", "value": "xyz789", "domain": ".example.com", "path": "/"}
  ]
}}
```

**Mobile viewport:**
```json
{"browser_config": {
  "viewport_width": 390,
  "viewport_height": 844,
  "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
}}
```

**Fast text-only mode (no images/CSS):**
```json
{"browser_config": {"text_mode": true}}
```

**Saved browser profile (logged-in session):**
```json
{"browser_config": {"profile_id": "uuid-from-profile-api"}}
```
