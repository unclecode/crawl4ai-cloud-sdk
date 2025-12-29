"""
Example: List jobs using HTTP API

This example demonstrates:
- GET /v1/crawl/jobs with query parameters
- Pagination and filtering
- Raw JSON response handling
"""

import requests
import json

API_KEY = "YOUR_API_KEY"
BASE_URL = "https://api.crawl4ai.com"

headers = {
    "Authorization": f"Bearer {API_KEY}",
    "Content-Type": "application/json"
}

# List all jobs
print("=== All Jobs (First 20) ===")
response = requests.get(
    f"{BASE_URL}/v1/crawl/jobs",
    headers=headers,
    params={"limit": 20}
)
response.raise_for_status()
data = response.json()

print(f"Total jobs: {data['total']}")
print(f"Showing: {len(data['jobs'])}")

for job in data['jobs']:
    print(f"  {job['job_id']}: {job['status']} | {len(job['urls'])} URLs")

# Filter by status
print("\n=== Completed Jobs ===")
response = requests.get(
    f"{BASE_URL}/v1/crawl/jobs",
    headers=headers,
    params={"status": "completed", "limit": 10}
)
completed = response.json()
for job in completed['jobs']:
    print(f"  {job['job_id']}: {job['urls'][0] if job['urls'] else 'N/A'}")

# Pagination
print("\n=== Pagination (Next 20) ===")
response = requests.get(
    f"{BASE_URL}/v1/crawl/jobs",
    headers=headers,
    params={"limit": 20, "offset": 20}
)
page2 = response.json()
print(f"Page 2: {len(page2['jobs'])} jobs")

# Failed jobs
print("\n=== Failed Jobs ===")
response = requests.get(
    f"{BASE_URL}/v1/crawl/jobs",
    headers=headers,
    params={"status": "failed", "limit": 5}
)
failed = response.json()
for job in failed['jobs']:
    print(f"  {job['job_id']}: {job.get('error', 'Unknown error')}")
