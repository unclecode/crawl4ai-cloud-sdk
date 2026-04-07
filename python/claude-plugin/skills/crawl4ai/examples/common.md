# Common Patterns

`$S` is shorthand for `$CLAUDE_PLUGIN_ROOT/scripts`.

---

### 1. Get markdown from a news article (fast, no JS needed)

```bash
python3 $S/crawl.py markdown "https://techcrunch.com/2024/01/15/some-article" --strategy http
```

### 2. Screenshot a pricing page (full page)

```bash
python3 $S/crawl.py screenshot "https://example.com/pricing" --full-page true
```

### 3. Extract products from e-commerce (auto mode)

```bash
python3 $S/crawl.py extract "https://books.toscrape.com" --query "get all books with title and price"
```

### 4. Find all blog posts on a domain (map with query)

```bash
python3 $S/crawl.py map "https://example.com" --query "blog" --score-threshold 0.3 --max-urls 50
```

### 5. Crawl a docs site (site crawl, http strategy, 10 pages)

```bash
python3 $S/crawl.py site "https://docs.crawl4ai.com" --max-pages 10 --strategy http --wait
```

### 6. Target a specific page section (css_selector)

```bash
python3 $S/crawl.py markdown "https://blog.example.com/post-123" \
  --crawler-config '{"css_selector":"article.post-body","excluded_tags":["nav","footer"]}'
```

### 7. Extract with LLM instructions (summarize)

```bash
python3 $S/crawl.py extract "https://en.wikipedia.org/wiki/Web_scraping" \
  --method llm --query "summarize in 5 bullet points with key facts"
```

### 8. Wait for lazy-loaded content (wait_for selector)

```bash
python3 $S/crawl.py markdown "https://spa-app.example.com/dashboard" \
  --strategy browser --crawler-config '{"wait_for":".data-table tbody tr","delay_before_return_html":2}'
```

### 9. Batch markdown for 5 URLs (async with wait)

```bash
python3 $S/crawl.py async markdown \
  "https://a.com/page1,https://b.com/page2,https://c.com/page3,https://d.com/page4,https://e.com/page5" \
  --strategy http --wait
```

### 10. Generate a PDF of a receipt page

```bash
python3 $S/crawl.py screenshot "https://app.example.com/receipt/12345" --pdf --wait-for ".receipt-total"
```
