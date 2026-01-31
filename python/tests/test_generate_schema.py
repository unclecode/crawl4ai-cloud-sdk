"""Real e2e tests for generate_schema with urls + html support."""
import asyncio
import os
import sys

from crawl4ai_cloud import AsyncWebCrawler

API_KEY = os.environ.get("CRAWL4AI_API_KEY", "")
BASE_URL = os.environ.get("CRAWL4AI_BASE_URL", "https://api.crawl4ai.com")

if not API_KEY:
    print("Set CRAWL4AI_API_KEY environment variable to run these tests")
    sys.exit(1)

RESULTS = {"passed": 0, "failed": 0}


def record(name: str, passed: bool, details: str = ""):
    status = "PASS" if passed else "FAIL"
    RESULTS["passed" if passed else "failed"] += 1
    print(f"  [{status}] {name}" + (f" — {details}" if details else ""))


async def main():
    print("=" * 60)
    print("  generate_schema — real API tests (crawl4ai-cloud-sdk)")
    print("=" * 60)
    print()

    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:

        # 1. urls mode (what tutorial.py uses)
        print("Test 1: generate_schema(urls=[...])")
        try:
            result = await crawler.generate_schema(
                urls=["https://books.toscrape.com"],
                query="Extract all book titles, prices, and ratings",
            )
            record("urls mode returns success", result.success)
            record("urls mode has schema", result.schema is not None)
            if result.schema:
                record("schema has baseSelector", "baseSelector" in result.schema)
                record("schema has fields", "fields" in result.schema)
                print(f"    schema name: {result.schema.get('name')}")
                print(f"    fields: {len(result.schema.get('fields', []))}")
        except Exception as e:
            record("urls mode", False, str(e))

        # 2. html string mode
        print("\nTest 2: generate_schema(html='...')")
        try:
            result2 = await crawler.generate_schema(
                html='<div class="product"><h2 class="title">Widget</h2><span class="price">$9.99</span><span class="stock">In stock</span></div>',
                query="Extract product name, price, and stock status",
            )
            record("html string mode returns success", result2.success)
            record("html string mode has schema", result2.schema is not None)
        except Exception as e:
            record("html string mode", False, str(e))

        # 3. html list mode (multi-sample)
        print("\nTest 3: generate_schema(html=[...list...])")
        try:
            result3 = await crawler.generate_schema(
                html=[
                    '<ul><li class="item"><span class="name">Apple</span><span class="price">$1</span></li></ul>',
                    '<ul><li class="item"><span class="name">Banana</span><span class="price">$2</span></li></ul>',
                ],
                query="Extract item names and prices",
            )
            record("html list mode returns success", result3.success)
            record("html list mode has schema", result3.schema is not None)
        except Exception as e:
            record("html list mode", False, str(e))

        # 4. validation: neither html nor urls
        print("\nTest 4: ValueError when neither provided")
        try:
            await crawler.generate_schema(query="should fail")
            record("raises ValueError", False, "no error raised")
        except ValueError:
            record("raises ValueError", True)
        except Exception as e:
            record("raises ValueError", False, f"wrong error: {type(e).__name__}: {e}")

    print()
    total = RESULTS["passed"] + RESULTS["failed"]
    print(f"Results: {RESULTS['passed']}/{total} passed, {RESULTS['failed']} failed")
    return RESULTS["failed"] == 0


if __name__ == "__main__":
    ok = asyncio.run(main())
    sys.exit(0 if ok else 1)
