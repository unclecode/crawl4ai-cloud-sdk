# Error Reference

All errors return `{"detail": "message"}`. Wrapper endpoints also include `error_message` in the response body.

---

| Code | Name | Causes | Fix |
|------|------|--------|-----|
| 400 | Bad Request | Invalid URL format, missing required field, malformed JSON body | Check URL has scheme (`https://`). Validate JSON payload. |
| 401 | Unauthorized | Missing `X-API-Key` header, invalid/expired key | Set `CRAWL4AI_API_KEY` env var. Run `test-key.sh` to verify. |
| 403 | Forbidden | Plan doesn't allow this feature, not resource owner | Upgrade plan. Check you own the job/profile you're accessing. |
| 404 | Not Found | Job ID doesn't exist, endpoint typo | Verify `job_id`. Check endpoint path (e.g. `/v1/markdown/jobs/` not `/v1/jobs/`). |
| 422 | Validation Error | Pydantic model rejection: wrong types, both `url` and `urls` set, `auto` method on batch extract | Read `detail` for specific field. For batch extract use `method: "llm"` or `"schema"`. |
| 429 | Rate/Quota Exceeded | Too many requests per minute, daily crawl limit hit, credit quota exhausted | Wait and retry. Check `X-RateLimit-Reset` header. Free: 10 req/min, 100/day. Pro: 100 req/min, 10k/day. |
| 500 | Server Error | Internal bug, worker crash | Retry once. If persistent, report with the `request_id` from response headers. |
| 503 | Service Unavailable | No workers available, all browsers busy | Retry after 5-10s. System auto-scales. Peak hours may be congested. |
| 504 | Gateway Timeout | Page took too long to load, extraction exceeded time limit | Increase `page_timeout` in crawler_config. Use `http` strategy if page is static. Simplify extraction. |

---

## Rate Limit Headers

Every response includes:

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Requests allowed per minute |
| `X-RateLimit-Remaining` | Requests left this window |
| `X-RateLimit-Reset` | Seconds until limit resets |
| `X-Storage-Used-MB` | Storage consumed |
| `X-Storage-Remaining-MB` | Storage left |

## Retry Strategy

1. **429**: Wait `X-RateLimit-Reset` seconds, then retry.
2. **503**: Retry after 5s, max 3 attempts.
3. **504**: Switch to `http` strategy, or increase `page_timeout`.
4. **5xx**: Retry once after 2s. If still failing, something is wrong server-side.
