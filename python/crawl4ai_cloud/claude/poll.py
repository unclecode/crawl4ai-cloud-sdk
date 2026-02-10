"""CLI script for polling deep crawl job status. Run in background via Claude Code."""
from __future__ import annotations

import argparse
import asyncio
import os
import sys
import time


async def poll_job(job_id: str, api_key: str, api_base: str,
                   interval: float = 3.0, timeout: float = 600) -> None:
    from crawl4ai_cloud import AsyncWebCrawler

    async with AsyncWebCrawler(api_key=api_key, base_url=api_base) as crawler:
        start = time.time()
        while True:
            elapsed = time.time() - start
            if elapsed > timeout:
                print(f"TIMEOUT after {timeout:.0f}s. Job {job_id} still running.")
                sys.exit(1)

            try:
                if job_id.startswith("scan_"):
                    result = await crawler.get_deep_crawl_status(job_id)
                    is_complete = hasattr(result, "is_complete") and result.is_complete
                    count = result.discovered_count if hasattr(result, "discovered_count") else 0
                    print(f"[{elapsed:.0f}s] Scan {job_id}: {count} URLs discovered", flush=True)

                    if is_complete:
                        if hasattr(result, "crawl_job_id") and result.crawl_job_id:
                            job_id = result.crawl_job_id
                            print(f"Scan complete. Now tracking crawl job: {job_id}", flush=True)
                            continue
                        print(f"DONE: {count} URLs discovered (scan only)")
                        return
                else:
                    job = await crawler.get_job(job_id)
                    status = job.status if hasattr(job, "status") else "unknown"
                    progress = job.progress if hasattr(job, "progress") and job.progress else {}
                    print(f"[{elapsed:.0f}s] Job {job_id}: {status} {progress}", flush=True)

                    is_complete = (hasattr(job, "is_complete") and job.is_complete) or \
                                  status in ("completed", "done")
                    if is_complete:
                        dl_url = await crawler.download_url(job_id)
                        print(f"DONE: Job {job_id} complete. Download URL:")
                        print(dl_url)
                        return
            except Exception as e:
                print(f"[{elapsed:.0f}s] Poll error: {e}", flush=True)

            await asyncio.sleep(interval)


def main():
    parser = argparse.ArgumentParser(
        description="Poll Crawl4AI deep crawl job status until completion"
    )
    parser.add_argument("--job-id", required=True, help="Job ID to poll (scan_* or job_*)")
    parser.add_argument("--api-key", default=os.environ.get("CRAWL4AI_API_KEY"),
                        help="API key (default: $CRAWL4AI_API_KEY)")
    parser.add_argument("--api-base", default="https://api.crawl4ai.com",
                        help="API base URL")
    parser.add_argument("--interval", type=float, default=3.0,
                        help="Poll interval in seconds (default: 3)")
    parser.add_argument("--timeout", type=float, default=600,
                        help="Max wait time in seconds (default: 600)")
    args = parser.parse_args()

    if not args.api_key:
        print("ERROR: --api-key or CRAWL4AI_API_KEY environment variable required", file=sys.stderr)
        sys.exit(1)

    asyncio.run(poll_job(args.job_id, args.api_key, args.api_base, args.interval, args.timeout))


if __name__ == "__main__":
    main()
