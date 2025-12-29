"""
Example: Job management using HTTP API

This example demonstrates:
- GET /v1/crawl/jobs/{id} - Get job details
- DELETE /v1/crawl/jobs/{id} - Cancel/delete job
- GET /v1/crawl/jobs/{id}/download - Get download URL
"""

import requests
import json

API_KEY = "YOUR_API_KEY"
BASE_URL = "https://api.crawl4ai.com"

headers = {
    "Authorization": f"Bearer {API_KEY}",
    "Content-Type": "application/json"
}

# Create a test job
print("=== Creating Test Job ===")
response = requests.post(
    f"{BASE_URL}/v1/crawl/async",
    headers=headers,
    json={
        "urls": ["https://example.com", "https://example.org"],
        "priority": 5
    }
)
response.raise_for_status()
job = response.json()
job_id = job['job_id']
print(f"Created job: {job_id}")

# Get job details
print("\n=== Get Job Details ===")
response = requests.get(
    f"{BASE_URL}/v1/crawl/jobs/{job_id}",
    headers=headers
)
job_details = response.json()
print(f"Status: {job_details['status']}")
print(f"URLs: {job_details['urls']}")

# Cancel job (keep results)
print("\n=== Cancel Job (Keep Results) ===")
response = requests.delete(
    f"{BASE_URL}/v1/crawl/jobs/{job_id}",
    headers=headers,
    params={"delete_results": "false"}
)
cancelled = response.json()
print(f"Status: {cancelled['status']}")

# Create and delete completely
print("\n=== Cancel + Delete Results ===")
response = requests.post(
    f"{BASE_URL}/v1/crawl/async",
    headers=headers,
    json={"urls": ["https://example.com"]}
)
job2_id = response.json()['job_id']
print(f"Created job: {job2_id}")

response = requests.delete(
    f"{BASE_URL}/v1/crawl/jobs/{job2_id}",
    headers=headers,
    params={"delete_results": "true"}
)
print(f"Deleted: {response.json()['status']}")

# Get download URL
print("\n=== Get Download URL ===")
try:
    # Find a completed job
    response = requests.get(
        f"{BASE_URL}/v1/crawl/jobs",
        headers=headers,
        params={"status": "completed", "limit": 1}
    )
    jobs = response.json()

    if jobs['jobs']:
        completed_job_id = jobs['jobs'][0]['job_id']
        response = requests.get(
            f"{BASE_URL}/v1/crawl/jobs/{completed_job_id}/download",
            headers=headers,
            params={"expires_in": 3600}
        )
        download_data = response.json()
        print(f"Download URL: {download_data['url'][:100]}...")
        print("URL expires in 3600 seconds (1 hour)")
    else:
        print("No completed jobs found")
except Exception as e:
    print(f"Error: {e}")
