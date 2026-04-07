# Advanced Patterns

`$S` is shorthand for `$CLAUDE_PLUGIN_ROOT/scripts`.

---

### 1. Full /v1/crawl with LLM extraction strategy

Extract structured data using the raw crawl endpoint with an embedded LLM extraction strategy. Use when you need both the page markdown and extracted data in one call.

```bash
python3 $S/crawl.py crawl "https://news.ycombinator.com" \
  --crawler-config '{
    "extraction_strategy": {
      "type": "llm",
      "provider": "openai/gpt-4o-mini",
      "instruction": "Extract the top 10 stories with title, url, points, and comment count"
    },
    "content_filter": {"type": "PruningContentFilter"}
  }'
```

Response includes both `markdown` and `extracted_content` (JSON string).

---

### 2. Deep crawl with BFS + URL pattern filter

Crawl only blog posts on a site using breadth-first discovery with a glob pattern.

```bash
python3 $S/crawl.py site "https://example.com" \
  --discovery bfs --max-pages 50 --max-depth 3 \
  --pattern "*/blog/*" --strategy http --wait
```

Poll returns results per page. Each page has `markdown`, `links`, `metadata`.

---

### 3. JavaScript click before extraction (load-more button)

Click a "Load More" button to expand content, wait for new items, then extract.

```bash
python3 $S/crawl.py extract "https://example.com/products" \
  --strategy browser --query "extract all products with name, price, rating" \
  --crawler-config '{
    "js_code": "document.querySelector(\"button.load-more\").click()",
    "delay_before_return_html": 3,
    "wait_for": ".product-card:nth-child(20)"
  }'
```

---

### 4. Custom headers + cookies for an authenticated page

Scrape behind a login using session cookies and custom headers.

```bash
python3 $S/crawl.py markdown "https://app.example.com/dashboard" \
  --strategy browser \
  --browser-config '{
    "headers": {"Authorization": "Bearer eyJ..."},
    "cookies": [
      {"name": "session", "value": "abc123", "domain": ".example.com", "path": "/"},
      {"name": "_csrf", "value": "xyz789", "domain": ".example.com", "path": "/"}
    ],
    "viewport_width": 1440,
    "viewport_height": 900
  }' \
  --crawler-config '{"css_selector": ".dashboard-content", "wait_for": ".data-loaded"}'
```

---

### 5. Extract schema from one page, reuse on batch

Step 1: Run extract on a single page to get the CSS schema.

```bash
python3 $S/crawl.py extract "https://books.toscrape.com/catalogue/page-1.html" \
  --method schema --query "extract books with title, price, rating, stock status"
```

The response includes `schema_used`. Copy that schema object.

Step 2: Reuse the schema on multiple pages via async batch (no LLM cost).

```bash
python3 $S/crawl.py async extract \
  "https://books.toscrape.com/catalogue/page-1.html,https://books.toscrape.com/catalogue/page-2.html,https://books.toscrape.com/catalogue/page-3.html" \
  --method schema --wait
```

CSS schema extraction is deterministic and free (no LLM tokens). Generate once, reuse everywhere.
