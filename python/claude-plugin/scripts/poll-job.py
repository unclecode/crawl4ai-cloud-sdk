#!/usr/bin/env python3
"""Poll a Crawl4AI async job until complete."""

import argparse
import json
import os
import sys
import time
from pathlib import Path
from urllib.request import Request, urlopen
from urllib.error import HTTPError


def get_config():
    api_key = os.environ.get("CRAWL4AI_API_KEY", "")
    api_url = os.environ.get("CRAWL4AI_API_URL", "")
    if not api_key or not api_url:
        config_path = Path.home() / ".crawl4ai" / "claude_config.json"
        if config_path.exists():
            cfg = json.load(open(config_path))
            api_key = api_key or cfg.get("api_key", "")
            api_url = api_url or cfg.get("api_base_url", "")
    return api_key, (api_url or "https://api.crawl4ai.com").rstrip("/")


def get_job(job_id, job_type, api_key, api_url):
    if job_type == "deep" or job_id.startswith("scan_"):
        path = f"/v1/crawl/deep/jobs/{job_id}"
    else:
        path = f"/v1/{job_type}/jobs/{job_id}"

    url = f"{api_url}{path}"
    req = Request(url, headers={"X-API-Key": api_key})
    try:
        resp = urlopen(req, timeout=30)
        return json.loads(resp.read().decode())
    except HTTPError as e:
        return {"error": e.read().decode()[:200], "status_code": e.code}


def main():
    parser = argparse.ArgumentParser(description="Poll Crawl4AI async job")
    parser.add_argument("job_id", help="Job ID (job_xxx or scan_xxx)")
    parser.add_argument("--type", default="markdown", choices=["markdown", "screenshot", "extract", "deep"])
    parser.add_argument("--interval", type=float, default=3.0)
    parser.add_argument("--timeout", type=float, default=300.0)
    args = parser.parse_args()

    api_key, api_url = get_config()
    if not api_key:
        print(json.dumps({"error": "No API key configured"}))
        sys.exit(1)

    start = time.time()
    while True:
        data = get_job(args.job_id, args.type, api_key, api_url)
        status = data.get("status", "unknown")
        progress = data.get("progress", data.get("pages_crawled", "?"))
        elapsed = int(time.time() - start)
        print(f"[{elapsed}s] status={status} progress={progress}", file=sys.stderr)

        if status in ("completed", "partial", "failed", "cancelled"):
            print(json.dumps(data, indent=2))
            sys.exit(0 if status in ("completed", "partial") else 1)

        if data.get("error") or data.get("status_code"):
            print(json.dumps(data, indent=2))
            sys.exit(1)

        if time.time() - start > args.timeout:
            print(json.dumps({"error": f"Timeout after {args.timeout}s", "last_status": data}, indent=2))
            sys.exit(1)

        time.sleep(args.interval)


if __name__ == "__main__":
    main()
