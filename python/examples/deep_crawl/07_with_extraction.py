#!/usr/bin/env python3
"""
Deep Crawl - With Extraction Strategies

Combine deep crawl with extraction strategies to get structured data
from all discovered pages. Extraction runs during the crawl phase.

Extraction strategies:
- json_css: CSS selectors for structured data
- llm: LLM-based extraction (requires LLM config)
- cosine: Semantic clustering

Usage:
    python 07_with_extraction.py
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler
import json

API_KEY = "YOUR_API_KEY"


def css_extraction():
    """Extract structured data using CSS selectors."""
    print("=== CSS Extraction ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        # Define CSS extraction schema
        schema = {
            "name": "Documentation",
            "baseSelector": "main, article, .content",
            "fields": [
                {
                    "name": "title",
                    "selector": "h1",
                    "type": "text"
                },
                {
                    "name": "description",
                    "selector": "p.description, .intro, meta[name='description']",
                    "type": "text"
                },
                {
                    "name": "headings",
                    "selector": "h2, h3",
                    "type": "list"
                },
                {
                    "name": "code_blocks",
                    "selector": "pre code, .highlight",
                    "type": "list"
                }
            ]
        }

        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=1,
            max_urls=5,
            crawler_config={
                "extraction_strategy": {
                    "type": "json_css",
                    "schema": schema
                }
            },
            wait=True
        )

        print(f"Pages crawled: {job.progress.completed}")

        if job.results:
            print("\nExtracted content:")
            for r in job.results[:3]:
                print(f"\nURL: {r['url']}")
                if r.get('extracted_content'):
                    try:
                        data = json.loads(r['extracted_content'])
                        print(f"  Title: {data.get('title', 'N/A')}")
                        headings = data.get('headings', [])
                        if headings:
                            print(f"  Headings: {len(headings)}")
                            for h in headings[:3]:
                                print(f"    - {h}")
                    except json.JSONDecodeError:
                        print("  (Parse error)")

    finally:
        


def nested_css_extraction():
    """Extract nested/repeated data structures."""
    print("\n=== Nested CSS Extraction ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        # Schema for extracting repeated items (e.g., products, posts)
        schema = {
            "name": "APIReference",
            "baseSelector": ".method, .function, .endpoint",
            "fields": [
                {
                    "name": "name",
                    "selector": "h3, .method-name",
                    "type": "text"
                },
                {
                    "name": "signature",
                    "selector": ".signature, code:first-of-type",
                    "type": "text"
                },
                {
                    "name": "description",
                    "selector": "p, .description",
                    "type": "text"
                },
                {
                    "name": "parameters",
                    "selector": ".param, li",
                    "type": "nested",
                    "fields": [
                        {"name": "name", "selector": ".param-name, code", "type": "text"},
                        {"name": "type", "selector": ".param-type, em", "type": "text"},
                        {"name": "desc", "selector": ".param-desc, span", "type": "text"}
                    ]
                }
            ]
        }

        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=1,
            max_urls=5,
            pattern="*/api/*",
            crawler_config={
                "extraction_strategy": {
                    "type": "json_css",
                    "schema": schema
                }
            },
            wait=True
        )

        print(f"API pages: {job.progress.completed}")

    finally:
        


def extraction_with_attributes():
    """Extract element attributes (href, src, etc.)."""
    print("\n=== Extract Attributes ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        schema = {
            "name": "PageAssets",
            "baseSelector": "body",
            "fields": [
                {
                    "name": "links",
                    "selector": "a[href]",
                    "type": "list",
                    "attribute": "href"  # Get href attribute
                },
                {
                    "name": "images",
                    "selector": "img[src]",
                    "type": "list",
                    "attribute": "src"  # Get src attribute
                },
                {
                    "name": "meta_tags",
                    "selector": "meta[name]",
                    "type": "nested",
                    "fields": [
                        {"name": "name", "selector": "", "type": "attribute", "attribute": "name"},
                        {"name": "content", "selector": "", "type": "attribute", "attribute": "content"}
                    ]
                }
            ]
        }

        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="map",
            max_urls=3,
            crawler_config={
                "extraction_strategy": {
                    "type": "json_css",
                    "schema": schema
                }
            },
            wait=True
        )

        print(f"Pages processed: {job.progress.completed}")

        if job.results:
            for r in job.results[:1]:
                if r.get('extracted_content'):
                    data = json.loads(r['extracted_content'])
                    links = data.get('links', [])
                    images = data.get('images', [])
                    print(f"\nURL: {r['url']}")
                    print(f"  Links found: {len(links)}")
                    print(f"  Images found: {len(images)}")

    finally:
        


def llm_extraction():
    """Extract using LLM (requires LLM provider config)."""
    print("\n=== LLM Extraction ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="map",
            max_urls=3,
            crawler_config={
                "extraction_strategy": {
                    "type": "llm",
                    "provider": "openai",  # or "anthropic", "ollama"
                    "model": "gpt-4o-mini",
                    "schema": {
                        "name": "PageSummary",
                        "fields": [
                            {"name": "title", "type": "string"},
                            {"name": "summary", "type": "string", "description": "2-3 sentence summary"},
                            {"name": "topics", "type": "list", "description": "Main topics covered"},
                            {"name": "code_examples", "type": "boolean", "description": "Has code examples?"}
                        ]
                    },
                    "instruction": "Extract the main content and summarize this documentation page."
                }
            },
            wait=True
        )

        print(f"LLM processed: {job.progress.completed}")

    finally:
        


def extraction_product_catalog():
    """Example: Extract product catalog data."""
    print("\n=== Product Catalog Extraction ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        schema = {
            "name": "Products",
            "baseSelector": ".product, [itemtype*='Product'], .product-card",
            "fields": [
                {"name": "name", "selector": "h1, h2, .product-title", "type": "text"},
                {"name": "price", "selector": ".price, [itemprop='price']", "type": "text"},
                {"name": "currency", "selector": "[itemprop='priceCurrency']", "type": "attribute", "attribute": "content"},
                {"name": "description", "selector": ".description, [itemprop='description']", "type": "text"},
                {"name": "image", "selector": "img.product-image, [itemprop='image']", "type": "attribute", "attribute": "src"},
                {"name": "sku", "selector": ".sku, [itemprop='sku']", "type": "text"},
                {"name": "availability", "selector": ".stock, [itemprop='availability']", "type": "text"},
                {"name": "rating", "selector": ".rating, [itemprop='ratingValue']", "type": "text"}
            ]
        }

        job = await crawler.deep_crawl(
            url="https://example-shop.com",
            strategy="bfs",
            max_depth=2,
            max_urls=50,
            pattern="/product/*",
            crawler_config={
                "extraction_strategy": {
                    "type": "json_css",
                    "schema": schema
                }
            },
            wait=True
        )

        print(f"Products extracted: {job.progress.completed}")

    finally:
        


def extraction_blog_posts():
    """Example: Extract blog post content."""
    print("\n=== Blog Post Extraction ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        schema = {
            "name": "BlogPost",
            "baseSelector": "article, .post, main",
            "fields": [
                {"name": "title", "selector": "h1, .post-title", "type": "text"},
                {"name": "author", "selector": ".author, [rel='author'], .byline", "type": "text"},
                {"name": "date", "selector": "time, .date, .published", "type": "text"},
                {"name": "categories", "selector": ".category, .tag", "type": "list"},
                {"name": "content", "selector": ".content, .post-body, .entry-content", "type": "text"},
                {"name": "reading_time", "selector": ".reading-time", "type": "text"}
            ]
        }

        job = await crawler.deep_crawl(
            url="https://example-blog.com",
            strategy="bfs",
            max_depth=2,
            max_urls=30,
            pattern="/blog/*",
            filter_nonsense_urls=True,
            crawler_config={
                "extraction_strategy": {
                    "type": "json_css",
                    "schema": schema
                }
            },
            wait=True
        )

        print(f"Blog posts extracted: {job.progress.completed}")

    finally:
        


if __name__ == "__main__":
    css_extraction()
    # Uncomment to run other examples:
    # nested_css_extraction()
    # extraction_with_attributes()
    # llm_extraction()  # Requires LLM provider setup
    # extraction_product_catalog()
    # extraction_blog_posts()
