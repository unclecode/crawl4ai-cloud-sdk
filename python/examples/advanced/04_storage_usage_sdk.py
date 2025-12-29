#!/usr/bin/env python3
"""
Storage Usage Monitoring - SDK Example

This script demonstrates how to check and monitor your storage quota.
Storage is used by async job results stored in S3.

Usage:
    python 04_storage_usage_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler
from crawl4ai_cloud.exceptions import QuotaExceededError

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def check_storage_usage():
    """Check current storage usage and quota."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        print("Checking storage usage...")
        usage = await crawler.get_storage_usage()

        print(f"\n=== STORAGE USAGE ===")
        print(f"Used: {usage.used_mb:.2f} MB")
        print(f"Max: {usage.max_mb:.2f} MB")
        print(f"Remaining: {usage.remaining_mb:.2f} MB")
        print(f"Usage: {usage.percentage:.1f}%")

        # Check if storage is getting full
        if usage.percentage > 90:
            print("\nWARNING: Storage is over 90% full!")
            print("Consider deleting old jobs to free up space.")
        elif usage.percentage > 75:
            print("\nNOTE: Storage is over 75% full.")
        else:
            print("\nStorage usage is healthy.")

        return usage


async def monitor_storage_during_crawl(urls: list):
    """Monitor storage usage while running async crawls."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # Check initial storage
        initial = await crawler.get_storage_usage()
        print(f"Initial storage: {initial.used_mb:.2f} MB / {initial.max_mb:.2f} MB")

        try:
            # Create async job
            print(f"\nStarting async crawl for {len(urls)} URLs...")
            results = await crawler.run_many(
                urls=urls,
                wait=True,  # Wait for completion
            )

            print(f"Crawl completed: {len(results)} results")

            # Check storage after job
            after = await crawler.get_storage_usage()
            print(f"\nAfter crawl storage: {after.used_mb:.2f} MB / {after.max_mb:.2f} MB")
            print(f"Storage used by this job: {after.used_mb - initial.used_mb:.2f} MB")

        except QuotaExceededError as e:
            if e.quota_type == "storage":
                print("\nStorage quota exceeded!")
                print("Delete old jobs to free up space.")

                # List jobs to find candidates for deletion
                jobs = await crawler.list_jobs(limit=10, status="completed")
                print(f"\nRecent completed jobs ({len(jobs.jobs)}):")
                for job in jobs.jobs[:5]:
                    print(f"  - {job.job_id} (created: {job.created_at})")
            else:
                print(f"Quota exceeded: {e.quota_type}")


async def cleanup_old_jobs():
    """Delete old jobs to free up storage."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # Check current storage
        usage = await crawler.get_storage_usage()
        print(f"Current storage: {usage.used_mb:.2f} MB / {usage.max_mb:.2f} MB")

        # List completed jobs
        jobs = await crawler.list_jobs(limit=20, status="completed")
        print(f"\nFound {len(jobs.jobs)} completed jobs")

        if not jobs.jobs:
            print("No jobs to delete.")
            return

        # Delete oldest jobs (be careful in production!)
        print("\nDeleting oldest 3 jobs...")
        deleted_count = 0

        for job in jobs.jobs[-3:]:  # Last 3 (oldest with default sorting)
            try:
                await crawler.cancel_job(job.job_id, delete_results=True)
                print(f"  Deleted job {job.job_id}")
                deleted_count += 1
            except Exception as e:
                print(f"  Failed to delete {job.job_id}: {e}")

        # Check storage after cleanup
        if deleted_count > 0:
            after = await crawler.get_storage_usage()
            freed = usage.used_mb - after.used_mb
            print(f"\nFreed {freed:.2f} MB of storage")
            print(f"New usage: {after.used_mb:.2f} MB / {after.max_mb:.2f} MB")


async def main():
    # Example 1: Check storage
    print("=== Example 1: Check Storage Usage ===")
    await check_storage_usage()

    # Example 2: Monitor during crawl (uncomment to use)
    # print("\n=== Example 2: Monitor During Crawl ===")
    # await monitor_storage_during_crawl([
    #     "https://www.example.com",
    #     "https://www.example.com/about",
    # ])

    # Example 3: Cleanup old jobs (uncomment to use)
    # print("\n=== Example 3: Cleanup Old Jobs ===")
    # await cleanup_old_jobs()


if __name__ == "__main__":
    asyncio.run(main())
