#!/usr/bin/env npx ts-node
/**
 * Error Handling - SDK Example
 *
 * This script demonstrates how to properly handle exceptions when using the Crawl4AI SDK.
 * Covers rate limits, quota errors, authentication errors, and other common exceptions.
 *
 * Usage:
 *     npx ts-node 03_error_handling_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';
import {
  RateLimitError,
  QuotaExceededError,
  AuthenticationError,
  ValidationError,
  NotFoundError,
  ServerError,
  TimeoutError,
  Crawl4AIError,
} from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function crawlWithErrorHandling(url: string): Promise<any> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    console.log(`Crawling ${url}...`);
    const result = await crawler.run(url);
    console.log(`Success! HTML size: ${result.html?.length || 0} bytes`);
    return result;
  } catch (e) {
    if (e instanceof AuthenticationError) {
      console.log(`Authentication failed: ${e.message}`);
      console.log('Check your API key and make sure it is valid');
    } else if (e instanceof RateLimitError) {
      console.log(`Rate limit exceeded: ${e.message}`);
      console.log(`Limit: ${e.limit} requests per minute`);
      console.log(`Remaining: ${e.remaining}`);
      console.log(`Retry after: ${e.retryAfter} seconds`);
    } else if (e instanceof QuotaExceededError) {
      console.log(`Quota exceeded: ${e.message}`);
      console.log(`Quota type: ${e.quotaType}`); // 'daily', 'concurrent', or 'storage'

      if (e.quotaType === 'storage') {
        console.log('Your storage is full. Delete old jobs to free up space.');
      } else if (e.quotaType === 'daily') {
        console.log('Daily crawl limit reached. Wait until tomorrow or upgrade plan.');
      } else if (e.quotaType === 'concurrent') {
        console.log('Too many concurrent requests. Wait for some to complete.');
      }
    } else if (e instanceof ValidationError) {
      console.log(`Invalid request: ${e.message}`);
      console.log('Check your URL and configuration parameters');
    } else if (e instanceof NotFoundError) {
      console.log(`Resource not found: ${e.message}`);
      console.log('The job or session ID may be invalid or expired');
    } else if (e instanceof TimeoutError) {
      console.log(`Request timed out: ${e.message}`);
      console.log('The page took too long to load. Try increasing timeout.');
    } else if (e instanceof ServerError) {
      console.log(`Server error: ${e.message}`);
      console.log('The API is experiencing issues. Try again later.');
    } else if (e instanceof Crawl4AIError) {
      // Catch-all for other SDK errors
      console.log(`Crawl4AI error: ${e.message}`);
      console.log(`Status code: ${e.statusCode}`);
    } else {
      // Catch-all for unexpected errors
      console.log(`Unexpected error: ${e}`);
    }

    return null;
  } finally {
    await crawler.close();
  }
}

async function crawlWithRetryLogic(url: string, maxRetries = 3): Promise<any> {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

    try {
      console.log(`Attempt ${attempt + 1}/${maxRetries}: Crawling ${url}...`);
      const result = await crawler.run(url);
      console.log(`Success! HTML size: ${result.html?.length || 0} bytes`);
      return result;
    } catch (e) {
      if (e instanceof RateLimitError) {
        if (e.retryAfter > 0) {
          console.log(`Rate limited. Waiting ${e.retryAfter}s...`);
          await sleep(e.retryAfter * 1000);
          continue;
        } else {
          break;
        }
      } else if (e instanceof ServerError || e instanceof TimeoutError) {
        if (attempt < maxRetries - 1) {
          const waitTime = Math.pow(2, attempt); // Exponential backoff
          console.log(`Transient error: ${e}. Retrying in ${waitTime}s...`);
          await sleep(waitTime * 1000);
          continue;
        } else {
          console.log(`Max retries reached. Last error: ${e}`);
          break;
        }
      } else if (
        e instanceof AuthenticationError ||
        e instanceof QuotaExceededError ||
        e instanceof ValidationError
      ) {
        // Don't retry on these errors
        console.log(`Non-retryable error: ${e}`);
        break;
      }
    } finally {
      await crawler.close();
    }
  }

  return null;
}

async function main(): Promise<void> {
  // Example 1: Basic error handling
  console.log('=== Example 1: Basic Error Handling ===');
  await crawlWithErrorHandling('https://www.example.com');

  // Example 2: Retry logic
  console.log('\n=== Example 2: With Retry Logic ===');
  await crawlWithRetryLogic('https://www.example.com');
}

main().catch(console.error);
