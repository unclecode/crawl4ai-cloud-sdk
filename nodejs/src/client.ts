/**
 * Internal HTTP client for Crawl4AI Cloud SDK.
 */

import fetch, { Response, RequestInit } from 'node-fetch';
import {
  CloudError,
  AuthenticationError,
  RateLimitError,
  QuotaExceededError,
  NotFoundError,
  ValidationError,
  ServerError,
  TimeoutError,
  ErrorHeaders,
} from './errors';

const VERSION = '0.1.0';
const DEFAULT_BASE_URL = 'https://api.crawl4ai.com';
const DEFAULT_TIMEOUT = 120000; // 120 seconds in ms
const DEFAULT_MAX_RETRIES = 3;

interface RequestOptions {
  method?: string;
  params?: Record<string, string | number | boolean>;
  body?: Record<string, unknown>;
  timeout?: number;
}

/**
 * Internal async HTTP client with retries and error mapping.
 */
export class HTTPClient {
  private apiKey: string;
  private baseUrl: string;
  private timeout: number;
  private maxRetries: number;

  constructor(options: {
    apiKey?: string;
    baseUrl?: string;
    timeout?: number;
    maxRetries?: number;
  } = {}) {
    // Get API key from options or environment
    this.apiKey = options.apiKey || process.env.CRAWL4AI_API_KEY || '';

    if (!this.apiKey) {
      throw new Error(
        'API key is required. Provide it as an option or set ' +
          'the CRAWL4AI_API_KEY environment variable.'
      );
    }

    if (!this.apiKey.startsWith('sk_live_') && !this.apiKey.startsWith('sk_test_')) {
      throw new Error('Invalid API key format. Expected sk_live_* or sk_test_*');
    }

    this.baseUrl = (options.baseUrl || DEFAULT_BASE_URL).replace(/\/$/, '');
    this.timeout = options.timeout || DEFAULT_TIMEOUT;
    this.maxRetries = options.maxRetries || DEFAULT_MAX_RETRIES;
  }

  /**
   * Sleep for specified milliseconds.
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  /**
   * Build URL with query parameters.
   */
  private buildUrl(path: string, params?: Record<string, string | number | boolean>): string {
    const url = new URL(path, this.baseUrl);
    if (params) {
      for (const [key, value] of Object.entries(params)) {
        url.searchParams.append(key, String(value));
      }
    }
    return url.toString();
  }

  /**
   * Parse response headers into a simple object.
   */
  private parseHeaders(response: Response): ErrorHeaders {
    const headers: ErrorHeaders = {};
    response.headers.forEach((value, key) => {
      headers[key.toLowerCase()] = value;
    });
    return headers;
  }

  /**
   * Make HTTP request with error handling and retries.
   */
  async request(
    path: string,
    options: RequestOptions = {}
  ): Promise<Record<string, unknown>> {
    const { method = 'GET', params, body, timeout = this.timeout } = options;

    const url = this.buildUrl(path, params);
    const fetchOptions: RequestInit = {
      method,
      headers: {
        'X-API-Key': this.apiKey,
        'Content-Type': 'application/json',
        'User-Agent': `crawl4ai-cloud/${VERSION}`,
      },
      timeout,
    };

    if (body) {
      fetchOptions.body = JSON.stringify(body);
    }

    for (let attempt = 0; attempt < this.maxRetries; attempt++) {
      try {
        const response = await fetch(url, fetchOptions);

        // Success
        if (response.status < 400) {
          const text = await response.text();
          if (text) {
            return JSON.parse(text) as Record<string, unknown>;
          }
          return {};
        }

        // Parse error response
        let detail: string;
        let errorData: Record<string, unknown> = {};
        try {
          errorData = (await response.json()) as Record<string, unknown>;
          detail = (errorData.detail as string) || JSON.stringify(errorData);
        } catch {
          detail = (await response.text()) || `HTTP ${response.status}`;
        }

        const headers = this.parseHeaders(response);

        // Map status codes to exceptions
        switch (response.status) {
          case 401:
            throw new AuthenticationError(detail, 401, errorData, headers);
          case 404:
            throw new NotFoundError(detail, 404, errorData, headers);
          case 429:
            if (detail.toLowerCase().includes('rate limit')) {
              throw new RateLimitError(detail, 429, errorData, headers);
            } else {
              throw new QuotaExceededError(detail, 429, errorData, headers);
            }
          case 400:
            throw new ValidationError(detail, 400, errorData, headers);
          case 504:
            throw new TimeoutError(detail, 504, errorData, headers);
          default:
            if (response.status >= 500) {
              if (attempt < this.maxRetries - 1) {
                await this.sleep(Math.pow(2, attempt) * 1000);
                continue;
              }
              throw new ServerError(detail, response.status, errorData, headers);
            }
            throw new CloudError(detail, response.status, errorData, headers);
        }
      } catch (error) {
        // Re-throw CloudError instances
        if (error instanceof CloudError) {
          throw error;
        }

        // Handle fetch errors (timeout, network, etc.)
        const err = error as Error;
        if (err.name === 'AbortError' || err.message.includes('timeout')) {
          if (attempt < this.maxRetries - 1) {
            await this.sleep(Math.pow(2, attempt) * 1000);
            continue;
          }
          throw new TimeoutError(`Request timed out: ${err.message}`);
        }

        if (attempt < this.maxRetries - 1) {
          await this.sleep(Math.pow(2, attempt) * 1000);
          continue;
        }
        throw new CloudError(`Request failed: ${err.message}`);
      }
    }

    throw new CloudError('Max retries exceeded');
  }

  /**
   * GET request helper.
   */
  async get(
    path: string,
    params?: Record<string, string | number | boolean>,
    timeout?: number
  ): Promise<Record<string, unknown>> {
    return this.request(path, { method: 'GET', params, timeout });
  }

  /**
   * POST request helper.
   */
  async post(
    path: string,
    body?: Record<string, unknown>,
    timeout?: number
  ): Promise<Record<string, unknown>> {
    return this.request(path, { method: 'POST', body, timeout });
  }

  /**
   * DELETE request helper.
   */
  async delete(
    path: string,
    params?: Record<string, string | number | boolean>
  ): Promise<Record<string, unknown>> {
    return this.request(path, { method: 'DELETE', params });
  }
}
