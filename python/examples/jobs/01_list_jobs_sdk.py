#!/usr/bin/env python3
"""
Example: List jobs using SDK

This example demonstrates:
- Listing all jobs with pagination
- Filtering jobs by status
- Accessing job metadata (timestamps, URLs, status)
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def main():
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # List all jobs (default: 20 per page)
        print("=== All Jobs (First 20) ===")
        jobs = await crawler.list_jobs(limit=20)
        print(f"Total jobs: {jobs.total}")
        print(f"Showing: {len(jobs.jobs)}")

        for job in jobs.jobs:
            print(f"  {job.job_id}: {job.status} | {len(job.urls)} URLs | Created: {job.created_at}")

        # Filter by status
        print("\n=== Completed Jobs ===")
        completed = await crawler.list_jobs(status="completed", limit=10)
        for job in completed.jobs:
            print(f"  {job.job_id}: {job.urls[0] if job.urls else 'N/A'}")

        print("\n=== Running Jobs ===")
        running = await crawler.list_jobs(status="running")
        print(f"Found {len(running.jobs)} running jobs")

        # Pagination example
        print("\n=== Pagination (Next 20) ===")
        page2 = await crawler.list_jobs(limit=20, offset=20)
        print(f"Page 2: {len(page2.jobs)} jobs")

        # Available statuses: pending, running, completed, failed, cancelled
        print("\n=== Failed Jobs ===")
        failed = await crawler.list_jobs(status="failed", limit=5)
        for job in failed.jobs:
            print(f"  {job.job_id}: {job.error or 'Unknown error'}")


if __name__ == "__main__":
    asyncio.run(main())
