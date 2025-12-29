#!/usr/bin/env python3
"""
Example: Job management using SDK

This example demonstrates:
- Getting job details
- Cancelling running jobs
- Deleting job results to free storage
- Getting presigned download URLs
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def main():
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # Create an async job for testing
        print("=== Creating Test Job ===")
        results = await crawler.run_many(
            urls=["https://example.com", "https://example.org"],
            wait=False  # Don't wait, just create the job
        )
        # When wait=False, we get a job object back
        print(f"Created job")

        # Get job details
        print("\n=== Get Job Details ===")
        jobs = await crawler.list_jobs(limit=1)
        if jobs.jobs:
            job = jobs.jobs[0]
            print(f"Job ID: {job.job_id}")
            print(f"Status: {job.status}")
            print(f"URLs: {job.urls}")
            print(f"Created: {job.created_at}")

            # Cancel the job (without deleting results)
            print("\n=== Cancel Job (Keep Results) ===")
            cancelled = await crawler.cancel_job(job.job_id, delete_results=False)
            print(f"Status: {cancelled.status}")

        # Create another job and delete it completely
        print("\n=== Cancel + Delete Results ===")
        results2 = await crawler.run_many(
            urls=["https://example.com"],
            wait=False
        )

        jobs2 = await crawler.list_jobs(limit=1)
        if jobs2.jobs:
            job2 = jobs2.jobs[0]
            print(f"Created job: {job2.job_id}")
            deleted = await crawler.cancel_job(job2.job_id, delete_results=True)
            print(f"Deleted: {deleted.status}")

        # Get download URL for completed job (example)
        print("\n=== Get Download URL ===")
        try:
            completed_jobs = await crawler.list_jobs(status="completed", limit=1)
            if completed_jobs.jobs:
                job_id = completed_jobs.jobs[0].job_id
                download_url = await crawler.get_download_url(job_id, expires_in=3600)
                print(f"Download URL: {download_url[:100]}...")
                print("URL expires in 3600 seconds (1 hour)")
            else:
                print("No completed jobs found")
        except Exception as e:
            print(f"Error: {e}")


if __name__ == "__main__":
    asyncio.run(main())
