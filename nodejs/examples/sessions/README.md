# Sessions Examples

This directory contains TypeScript/JavaScript examples for browser session management with Crawl4AI Cloud.

## What are Sessions?

Browser sessions provide a persistent browser instance in the cloud. You get a WebSocket URL
that you can connect to using:

- **Crawl4AI** (local library with cloud browser)
- **Puppeteer** (puppeteer-core)
- **Playwright** (playwright-core)

## Files

- **01_basic_session_sdk.ts** - Create, use, and release sessions with SDK
- **01_basic_session_http.ts** - Session management via HTTP API
- **02_session_with_puppeteer.ts** - Connect to cloud session with Puppeteer
- **03_session_with_playwright.ts** - Connect to cloud session with Playwright

## Usage

```bash
# Basic session management
npx ts-node 01_basic_session_sdk.ts

# With Puppeteer
npm install puppeteer-core
npx ts-node 02_session_with_puppeteer.ts

# With Playwright
npm install playwright-core
npx ts-node 03_session_with_playwright.ts
```

## Session Lifecycle

1. **Create**: Request a new browser session (returns WebSocket URL)
2. **Connect**: Use the WebSocket URL with your preferred tool
3. **Use**: Navigate, interact, extract data
4. **Disconnect**: Close your connection (don't close the browser)
5. **Release**: Call the release endpoint to free resources

## Important Notes

- Sessions have a timeout (default: 600 seconds)
- Always release sessions when done to avoid charges
- Don't call `browser.close()` - use `browser.disconnect()` instead
- The cloud manages the browser lifecycle
