#!/usr/bin/env python3
"""
Crawl4AI CLI -- Universal entry point for all API endpoints.
Used by the Claude Code skill to execute crawl operations.

Usage:
    python3 crawl.py markdown URL [options]
    python3 crawl.py screenshot URL [options]
    python3 crawl.py extract URL [options]
    python3 crawl.py map URL [options]
    python3 crawl.py site URL [options]
    python3 crawl.py crawl URL [options]
    python3 crawl.py markdown-async URL1,URL2 [options]
"""

import argparse
import json
import os
import sys
import time
from pathlib import Path
from urllib.request import Request, urlopen
from urllib.error import HTTPError


def get_config():
    """Load API key and base URL from env vars or config file."""
    api_key = os.environ.get("CRAWL4AI_API_KEY", "")
    api_url = os.environ.get("CRAWL4AI_API_URL", "")

    if not api_key or not api_url:
        config_path = Path.home() / ".crawl4ai" / "claude_config.json"
        if config_path.exists():
            with open(config_path) as f:
                cfg = json.load(f)
            if not api_key:
                api_key = cfg.get("api_key", "")
            if not api_url:
                api_url = cfg.get("api_base_url", "")

    if not api_url:
        api_url = "https://api.crawl4ai.com"
    if not api_key:
        print(json.dumps({"success": False, "error": "No API key. Set CRAWL4AI_API_KEY or run /crawl4ai:setup"}))
        sys.exit(1)

    return api_key, api_url.rstrip("/")


def api_call(method, path, body=None, api_key=None, api_url=None, timeout=120):
    """Make an API call, return parsed JSON."""
    url = f"{api_url}{path}"
    headers = {"Content-Type": "application/json", "X-API-Key": api_key}
    data = json.dumps(body).encode() if body else None
    req = Request(url, data=data, headers=headers, method=method)

    try:
        resp = urlopen(req, timeout=timeout)
        return json.loads(resp.read().decode())
    except HTTPError as e:
        raw = e.read().decode()
        try:
            return {"success": False, "error": json.loads(raw).get("detail", raw[:200]), "status_code": e.code}
        except Exception:
            return {"success": False, "error": raw[:200], "status_code": e.code}
    except Exception as e:
        return {"success": False, "error": str(e)}


def poll_until_done(job_id, job_type, api_key, api_url, interval=3, timeout=300):
    """Poll a wrapper job until complete."""
    start = time.time()
    while True:
        if job_type == "deep":
            data = api_call("GET", f"/v1/crawl/deep/jobs/{job_id}", api_key=api_key, api_url=api_url)
        else:
            data = api_call("GET", f"/v1/{job_type}/jobs/{job_id}", api_key=api_key, api_url=api_url)

        status = data.get("status", "unknown")
        progress = data.get("progress", data.get("pages_crawled", "?"))
        print(f"  [{status}] progress={progress}", file=sys.stderr)

        if status in ("completed", "partial", "failed", "cancelled"):
            return data
        if time.time() - start > timeout:
            return {"success": False, "error": f"Timeout after {timeout}s", "last_status": data}
        time.sleep(interval)


def cmd_markdown(args):
    api_key, api_url = get_config()
    body = {"url": args.url, "strategy": args.strategy, "fit": args.fit}
    if args.include:
        body["include"] = args.include.split(",")
    if args.crawler_config:
        body["crawler_config"] = json.loads(args.crawler_config)
    if args.browser_config:
        body["browser_config"] = json.loads(args.browser_config)
    if args.bypass_cache:
        body["bypass_cache"] = True

    result = api_call("POST", "/v1/markdown", body, api_key, api_url)
    print(json.dumps(result, indent=2))


def cmd_screenshot(args):
    api_key, api_url = get_config()
    body = {"url": args.url, "full_page": args.full_page}
    if args.pdf:
        body["pdf"] = True
    if args.wait_for:
        body["wait_for"] = args.wait_for
    if args.crawler_config:
        body["crawler_config"] = json.loads(args.crawler_config)
    if args.browser_config:
        body["browser_config"] = json.loads(args.browser_config)

    result = api_call("POST", "/v1/screenshot", body, api_key, api_url, timeout=120)
    # Truncate base64 for display
    if result.get("screenshot") and len(result["screenshot"]) > 200:
        result["_screenshot_length"] = len(result["screenshot"])
        result["screenshot"] = result["screenshot"][:100] + "...[truncated]"
    if result.get("pdf") and len(result["pdf"]) > 200:
        result["_pdf_length"] = len(result["pdf"])
        result["pdf"] = result["pdf"][:100] + "...[truncated]"
    print(json.dumps(result, indent=2))


def cmd_extract(args):
    api_key, api_url = get_config()
    body = {"url": args.url, "method": args.method, "strategy": args.strategy}
    if args.query:
        body["query"] = args.query
    if args.json_example:
        body["json_example"] = json.loads(args.json_example)
    if args.crawler_config:
        body["crawler_config"] = json.loads(args.crawler_config)
    if args.browser_config:
        body["browser_config"] = json.loads(args.browser_config)

    result = api_call("POST", "/v1/extract", body, api_key, api_url, timeout=180)
    print(json.dumps(result, indent=2))


def cmd_map(args):
    api_key, api_url = get_config()
    body = {"url": args.url, "mode": args.mode}
    if args.max_urls:
        body["max_urls"] = args.max_urls
    if args.query:
        body["query"] = args.query
    if args.score_threshold is not None:
        body["score_threshold"] = args.score_threshold
    body["include_subdomains"] = args.include_subdomains
    body["extract_head"] = args.extract_head
    if args.force:
        body["force"] = True

    result = api_call("POST", "/v1/map", body, api_key, api_url, timeout=120)
    print(json.dumps(result, indent=2))


def cmd_site(args):
    api_key, api_url = get_config()
    body = {
        "url": args.url, "max_pages": args.max_pages,
        "discovery": args.discovery, "strategy": args.strategy, "fit": args.fit,
    }
    if args.pattern:
        body["pattern"] = args.pattern
    if args.max_depth:
        body["max_depth"] = args.max_depth
    if args.include:
        body["include"] = args.include.split(",")
    if args.crawler_config:
        body["crawler_config"] = json.loads(args.crawler_config)

    result = api_call("POST", "/v1/crawl/site", body, api_key, api_url)
    print(json.dumps(result, indent=2))

    if args.wait and result.get("job_id"):
        print("Polling...", file=sys.stderr)
        final = poll_until_done(result["job_id"], "deep", api_key, api_url)
        print(json.dumps(final, indent=2))


def cmd_crawl(args):
    """Full power /v1/crawl endpoint."""
    api_key, api_url = get_config()
    body = {"url": args.url, "strategy": args.strategy}
    if args.crawler_config:
        body["crawler_config"] = json.loads(args.crawler_config)
    if args.browser_config:
        body["browser_config"] = json.loads(args.browser_config)
    if args.bypass_cache:
        body["bypass_cache"] = True

    result = api_call("POST", "/v1/crawl", body, api_key, api_url)
    # Truncate large fields
    for field in ("html", "cleaned_html", "screenshot", "pdf"):
        if result.get(field) and len(str(result[field])) > 500:
            result[f"_{field}_length"] = len(str(result[field]))
            result[field] = str(result[field])[:200] + "...[truncated]"
    print(json.dumps(result, indent=2))


def cmd_async(args):
    """Async batch for any wrapper type."""
    api_key, api_url = get_config()
    urls = [u.strip() for u in args.urls.split(",")]
    endpoint_map = {
        "markdown": "/v1/markdown/async",
        "screenshot": "/v1/screenshot/async",
        "extract": "/v1/extract/async",
    }
    endpoint = endpoint_map.get(args.type)
    if not endpoint:
        print(json.dumps({"success": False, "error": f"Unknown async type: {args.type}"}))
        sys.exit(1)

    body = {"urls": urls}
    if args.type == "markdown":
        body["strategy"] = args.strategy or "http"
        body["fit"] = True
    elif args.type == "screenshot":
        body["full_page"] = True
    elif args.type == "extract":
        body["method"] = args.method or "llm"
        if args.query:
            body["query"] = args.query

    result = api_call("POST", endpoint, body, api_key, api_url)
    print(json.dumps(result, indent=2))

    if args.wait and result.get("job_id"):
        print("Polling...", file=sys.stderr)
        final = poll_until_done(result["job_id"], args.type, api_key, api_url)
        print(json.dumps(final, indent=2))


def main():
    parser = argparse.ArgumentParser(description="Crawl4AI CLI")
    sub = parser.add_subparsers(dest="command", required=True)

    # markdown
    p = sub.add_parser("markdown", help="Get markdown from URL")
    p.add_argument("url")
    p.add_argument("--strategy", default="browser", choices=["browser", "http"])
    p.add_argument("--fit", type=bool, default=True)
    p.add_argument("--include", help="Comma-separated: links,media,metadata,tables")
    p.add_argument("--crawler-config", dest="crawler_config")
    p.add_argument("--browser-config", dest="browser_config")
    p.add_argument("--bypass-cache", action="store_true", dest="bypass_cache")

    # screenshot
    p = sub.add_parser("screenshot", help="Capture screenshot/PDF")
    p.add_argument("url")
    p.add_argument("--full-page", type=bool, default=True, dest="full_page")
    p.add_argument("--pdf", action="store_true")
    p.add_argument("--wait-for", dest="wait_for")
    p.add_argument("--crawler-config", dest="crawler_config")
    p.add_argument("--browser-config", dest="browser_config")

    # extract
    p = sub.add_parser("extract", help="Extract structured data")
    p.add_argument("url")
    p.add_argument("--query")
    p.add_argument("--method", default="auto", choices=["auto", "llm", "schema"])
    p.add_argument("--strategy", default="http", choices=["browser", "http"])
    p.add_argument("--json-example", dest="json_example")
    p.add_argument("--crawler-config", dest="crawler_config")
    p.add_argument("--browser-config", dest="browser_config")

    # map
    p = sub.add_parser("map", help="Discover URLs on domain")
    p.add_argument("url")
    p.add_argument("--mode", default="default", choices=["default", "deep"])
    p.add_argument("--max-urls", type=int, dest="max_urls")
    p.add_argument("--query")
    p.add_argument("--score-threshold", type=float, dest="score_threshold")
    p.add_argument("--include-subdomains", action="store_true", default=False, dest="include_subdomains")
    p.add_argument("--extract-head", type=bool, default=True, dest="extract_head")
    p.add_argument("--force", action="store_true")

    # site
    p = sub.add_parser("site", help="Crawl entire website (async)")
    p.add_argument("url")
    p.add_argument("--max-pages", type=int, default=20, dest="max_pages")
    p.add_argument("--discovery", default="map", choices=["map", "bfs", "dfs", "best_first"])
    p.add_argument("--strategy", default="browser", choices=["browser", "http"])
    p.add_argument("--fit", type=bool, default=True)
    p.add_argument("--pattern")
    p.add_argument("--max-depth", type=int, dest="max_depth")
    p.add_argument("--include", help="Comma-separated: links,media,metadata,tables")
    p.add_argument("--crawler-config", dest="crawler_config")
    p.add_argument("--wait", action="store_true")

    # crawl (full power)
    p = sub.add_parser("crawl", help="Full power /v1/crawl")
    p.add_argument("url")
    p.add_argument("--strategy", default="browser", choices=["browser", "http"])
    p.add_argument("--crawler-config", dest="crawler_config")
    p.add_argument("--browser-config", dest="browser_config")
    p.add_argument("--bypass-cache", action="store_true", dest="bypass_cache")

    # async batch
    p = sub.add_parser("async", help="Async batch job")
    p.add_argument("type", choices=["markdown", "screenshot", "extract"])
    p.add_argument("urls", help="Comma-separated URLs")
    p.add_argument("--strategy", default=None)
    p.add_argument("--method", default=None)
    p.add_argument("--query")
    p.add_argument("--wait", action="store_true")

    args = parser.parse_args()
    dispatch = {
        "markdown": cmd_markdown, "screenshot": cmd_screenshot,
        "extract": cmd_extract, "map": cmd_map, "site": cmd_site,
        "crawl": cmd_crawl, "async": cmd_async,
    }
    dispatch[args.command](args)


if __name__ == "__main__":
    main()
