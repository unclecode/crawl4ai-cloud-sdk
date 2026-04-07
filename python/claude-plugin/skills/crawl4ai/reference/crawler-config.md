# CrawlerRunConfig Reference

Pass as `crawler_config` JSON object in any endpoint request body.
Fields not listed here are ignored. All fields are optional.

---

## Content Selection

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `css_selector` | string | `null` | Extract only matching elements | `"article.post-body"`, `"main"`, `"#content"` |
| `excluded_tags` | string[] | `null` | Strip these HTML tags | `["nav","footer","aside","header"]` |
| `excluded_selector` | string | `null` | CSS selector for elements to remove | `".sidebar, .ads, .cookie-banner"` |
| `target_elements` | string[] | `null` | Only keep these tag types | `["article","section","main"]` |
| `exclude_external_links` | bool | `false` | Remove links to other domains | `true` |
| `exclude_external_images` | bool | `false` | Remove images from other domains | `true` |
| `word_count_threshold` | int | `0` | Min words per text block | `10` |

```json
{"css_selector": "article", "excluded_tags": ["nav", "footer"], "word_count_threshold": 5}
```

## Wait / Timing

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `wait_for` | string | `null` | CSS selector to wait for before extraction | `".content-loaded"`, `"table tbody tr"` |
| `wait_until` | string | `"domcontentloaded"` | Page event | `"networkidle"`, `"load"`, `"commit"` |
| `delay_before_return_html` | float | `0` | Extra seconds after load | `2.0` |
| `page_timeout` | int | `30000` | Max ms for page load | `60000` |
| `wait_for_images` | bool | `false` | Wait for all images to load | `true` |

```json
{"wait_for": ".lazy-content", "delay_before_return_html": 1.5, "page_timeout": 60000}
```

## JavaScript Execution

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `js_code` | string | `null` | JavaScript to run before extraction | `"document.querySelector('.load-more').click()"` |
| `js_only` | bool | `false` | Only execute JS (skip crawl) | `true` |

```json
{"js_code": "window.scrollTo(0, document.body.scrollHeight); await new Promise(r => setTimeout(r, 2000));"}
```

## Scrolling

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `scan_full_page` | bool | `false` | Scroll through entire page | `true` |
| `scroll_delay` | float | `0.5` | Seconds between scroll steps | `1.0` |
| `max_scroll_steps` | int | `null` | Limit scroll iterations | `20` |

```json
{"scan_full_page": true, "scroll_delay": 0.8, "max_scroll_steps": 15}
```

## Screenshots / PDF

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `screenshot` | bool | `false` | Capture PNG screenshot | `true` |
| `screenshot_wait_for` | float | `null` | Seconds to wait before capture | `2.0` |
| `pdf` | bool | `false` | Generate PDF | `true` |

```json
{"screenshot": true, "pdf": true, "screenshot_wait_for": 1.5}
```

## Content Filter

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `content_filter` | object | `null` | Pruning filter config | `{"type": "PruningContentFilter"}` |

The `PruningContentFilter` removes nav, footer, sidebar, boilerplate. This is what `fit=true` uses internally.

```json
{"content_filter": {"type": "PruningContentFilter", "threshold": 0.48}}
```

## Extraction Strategy

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `extraction_strategy` | object | `null` | LLM or CSS extraction config | See below |

**LLM extraction:**
```json
{"extraction_strategy": {
  "type": "llm",
  "provider": "openai/gpt-4o-mini",
  "instruction": "Extract all product names and prices",
  "schema": {"name": "string", "price": "string"}
}}
```

**CSS extraction (reuse a schema):**
```json
{"extraction_strategy": {
  "type": "json_css",
  "schema": {"name": "ProductList", "baseSelector": ".product", "fields": [
    {"name": "title", "selector": "h2", "type": "text"},
    {"name": "price", "selector": ".price", "type": "text"}
  ]}
}}
```

## Output Control

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `only_text` | bool | `false` | Strip all HTML, return text only | `true` |
| `process_iframes` | bool | `false` | Extract content from iframes | `true` |
| `remove_forms` | bool | `false` | Remove form elements | `true` |

## Magic Mode

| Field | Type | Default | Description | Example |
|-------|------|---------|-------------|---------|
| `magic` | bool | `false` | Auto-detect and handle dynamic content | `true` |
| `simulate_user` | bool | `false` | Random mouse/scroll to bypass anti-bot | `true` |
| `override_navigator` | bool | `false` | Mask automation signals | `true` |

```json
{"magic": true, "simulate_user": true, "override_navigator": true}
```

Magic mode combines: `scan_full_page`, `simulate_user`, `override_navigator`, and smart wait detection. Use for anti-bot-heavy sites.
