# Session Management Examples

These examples demonstrate how to create and use browser sessions with Crawl4AI Cloud.

## What are Sessions?

Sessions provide persistent browser instances that you can connect to using:
- **Crawl4AI** - Use local Crawl4AI features with cloud browsers
- **Puppeteer** - Advanced browser automation from Node.js
- **Playwright** - Cross-browser testing and automation

## Examples

### Basic Session Management

1. **01_basic_session_sdk.py** - Create and release sessions using the Python SDK
2. **01_basic_session_http.py** - Create and release sessions using HTTP API

Both examples show how to:
- Create a browser session
- Get the WebSocket URL
- Check session status
- Release the session when done

### Using Sessions with Crawl4AI

3. **02_session_with_crawl4ai_sdk.py** - Use SDK to create session + local Crawl4AI
4. **02_session_with_crawl4ai_http.py** - Use HTTP API to create session + local Crawl4AI

These examples demonstrate:
- Creating a cloud browser session
- Connecting to it with local Crawl4AI library
- Running crawls with full Crawl4AI features
- Properly releasing the session

### Advanced Integration

5. **03_session_with_puppeteer.py** - Connect to sessions with Puppeteer/Playwright

Shows how to:
- Create a session from Python
- Connect with Puppeteer (Node.js)
- Use advanced browser automation
- Includes Playwright example

## Quick Start

1. Install dependencies:
```bash
# For SDK examples
pip install crawl4ai-cloud

# For HTTP examples
pip install httpx

# For Crawl4AI integration
pip install crawl4ai-cloud crawl4ai
```

2. Set your API key in the example files:
```python
API_KEY = "your_api_key_here"
```

3. Run an example:
```bash
python 01_basic_session_sdk.py
```

## Session Lifecycle

1. **Create** - Request a browser session with timeout
2. **Connect** - Use the WebSocket URL to connect
3. **Use** - Interact with the browser (crawl, automate, etc.)
4. **Release** - Free the session when done

## Important Notes

- Sessions have a timeout (default: 600 seconds / 10 minutes)
- Always release sessions when done to avoid charges
- One session = one browser instance
- WebSocket URLs are single-use and session-specific
- Don't call `browser.close()` in Puppeteer/Playwright - use `disconnect()` instead
