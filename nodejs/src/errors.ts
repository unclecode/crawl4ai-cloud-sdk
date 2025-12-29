/**
 * Exception classes for Crawl4AI Cloud SDK.
 */

export interface ErrorResponse {
  detail?: string;
  [key: string]: unknown;
}

export interface ErrorHeaders {
  [key: string]: string;
}

/**
 * Base error for Crawl4AI Cloud SDK.
 */
export class CloudError extends Error {
  public statusCode?: number;
  public response: ErrorResponse;
  public headers: ErrorHeaders;

  constructor(
    message: string,
    statusCode?: number,
    response: ErrorResponse = {},
    headers: ErrorHeaders = {}
  ) {
    super(message);
    this.name = 'CloudError';
    this.statusCode = statusCode;
    this.response = response;
    this.headers = headers;
  }

  toString(): string {
    if (this.statusCode) {
      return `[${this.statusCode}] ${this.message}`;
    }
    return this.message;
  }
}

/**
 * 401 - Invalid or missing API key.
 */
export class AuthenticationError extends CloudError {
  constructor(
    message: string,
    statusCode?: number,
    response: ErrorResponse = {},
    headers: ErrorHeaders = {}
  ) {
    super(message, statusCode, response, headers);
    this.name = 'AuthenticationError';
  }
}

/**
 * 429 - Rate limit exceeded.
 */
export class RateLimitError extends CloudError {
  constructor(
    message: string,
    statusCode?: number,
    response: ErrorResponse = {},
    headers: ErrorHeaders = {}
  ) {
    super(message, statusCode, response, headers);
    this.name = 'RateLimitError';
  }

  get retryAfter(): number {
    const value = this.headers['x-ratelimit-reset'];
    return value ? parseInt(value, 10) || 0 : 0;
  }

  get limit(): number {
    const value = this.headers['x-ratelimit-limit'];
    return value ? parseInt(value, 10) || 0 : 0;
  }

  get remaining(): number {
    const value = this.headers['x-ratelimit-remaining'];
    return value ? parseInt(value, 10) || 0 : 0;
  }
}

/**
 * 429 - Daily/concurrent/storage quota exceeded.
 */
export class QuotaExceededError extends CloudError {
  constructor(
    message: string,
    statusCode?: number,
    response: ErrorResponse = {},
    headers: ErrorHeaders = {}
  ) {
    super(message, statusCode, response, headers);
    this.name = 'QuotaExceededError';
  }

  get quotaType(): 'daily' | 'concurrent' | 'storage' {
    const msg = this.message.toLowerCase();
    if (msg.includes('storage')) return 'storage';
    if (msg.includes('concurrent')) return 'concurrent';
    return 'daily';
  }
}

/**
 * 404 - Resource not found.
 */
export class NotFoundError extends CloudError {
  constructor(
    message: string,
    statusCode?: number,
    response: ErrorResponse = {},
    headers: ErrorHeaders = {}
  ) {
    super(message, statusCode, response, headers);
    this.name = 'NotFoundError';
  }
}

/**
 * 400 - Invalid request parameters.
 */
export class ValidationError extends CloudError {
  constructor(
    message: string,
    statusCode?: number,
    response: ErrorResponse = {},
    headers: ErrorHeaders = {}
  ) {
    super(message, statusCode, response, headers);
    this.name = 'ValidationError';
  }
}

/**
 * 504 or client timeout.
 */
export class TimeoutError extends CloudError {
  constructor(
    message: string,
    statusCode?: number,
    response: ErrorResponse = {},
    headers: ErrorHeaders = {}
  ) {
    super(message, statusCode, response, headers);
    this.name = 'TimeoutError';
  }
}

/**
 * 500/503 - Server error.
 */
export class ServerError extends CloudError {
  constructor(
    message: string,
    statusCode?: number,
    response: ErrorResponse = {},
    headers: ErrorHeaders = {}
  ) {
    super(message, statusCode, response, headers);
    this.name = 'ServerError';
  }
}
