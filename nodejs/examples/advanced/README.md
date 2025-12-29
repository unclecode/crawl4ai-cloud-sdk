# Advanced Examples

This directory contains TypeScript/JavaScript examples for advanced Crawl4AI Cloud features.

## Files

### Screenshots and PDFs
- **01_screenshots_sdk.ts** - Capture screenshots and generate PDFs with SDK
- **01_screenshots_http.ts** - Screenshots and PDFs via HTTP API

### Custom Browser Configuration
- **02_custom_browser_config_sdk.ts** - Viewport, headers, proxy with SDK
- **02_custom_browser_config_http.ts** - Custom browser config via HTTP API

### Error Handling
- **03_error_handling_sdk.ts** - Handle rate limits, quota errors, retries

### Storage Monitoring
- **04_storage_usage_sdk.ts** - Monitor and manage storage quota

## Usage

```bash
npx ts-node 01_screenshots_sdk.ts
```

## Screenshot Options

```typescript
const result = await crawler.run(url, {
  config: {
    screenshot: true,           // Capture screenshot
    screenshot_wait_for: '.content',  // Wait for selector
    pdf: true,                  // Generate PDF
  }
});
```

## Custom Browser Options

```typescript
const result = await crawler.run(url, {
  strategy: 'browser',
  browserConfig: {
    viewport: { width: 1920, height: 1080 },
    headers: { 'User-Agent': 'CustomBot/1.0' },
  },
  proxy: { mode: 'datacenter' },
});
```

## Error Types

| Error | Retryable | Description |
|-------|-----------|-------------|
| `AuthenticationError` | No | Invalid API key |
| `RateLimitError` | Yes | Too many requests |
| `QuotaExceededError` | No | Daily/storage limit |
| `ValidationError` | No | Invalid parameters |
| `TimeoutError` | Yes | Request timed out |
| `ServerError` | Yes | API server issue |
