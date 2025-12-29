package crawl4ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// Version is the SDK version.
	Version = "0.1.0"

	// DefaultBaseURL is the default API base URL.
	DefaultBaseURL = "https://api.crawl4ai.com"

	// DefaultTimeout is the default request timeout.
	DefaultTimeout = 120 * time.Second

	// DefaultMaxRetries is the default max retry attempts.
	DefaultMaxRetries = 3
)

// HTTPClient is the internal HTTP client.
type HTTPClient struct {
	apiKey     string
	baseURL    string
	timeout    time.Duration
	maxRetries int
	client     *http.Client
}

// HTTPClientOptions are options for creating an HTTPClient.
type HTTPClientOptions struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}

// NewHTTPClient creates a new HTTPClient.
func NewHTTPClient(opts HTTPClientOptions) (*HTTPClient, error) {
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("CRAWL4AI_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("API key is required. Provide it as an option or set the CRAWL4AI_API_KEY environment variable")
	}

	if !strings.HasPrefix(apiKey, "sk_live_") && !strings.HasPrefix(apiKey, "sk_test_") {
		return nil, fmt.Errorf("invalid API key format. Expected sk_live_* or sk_test_*")
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	maxRetries := opts.MaxRetries
	if maxRetries == 0 {
		maxRetries = DefaultMaxRetries
	}

	return &HTTPClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		timeout:    timeout,
		maxRetries: maxRetries,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// RequestOptions are options for making a request.
type RequestOptions struct {
	Method  string
	Path    string
	Params  map[string]string
	Body    map[string]interface{}
	Timeout time.Duration
}

// Request makes an HTTP request with retries and error handling.
func (c *HTTPClient) Request(opts RequestOptions) (map[string]interface{}, error) {
	method := opts.Method
	if method == "" {
		method = "GET"
	}

	// Build URL
	reqURL := c.baseURL + opts.Path
	if len(opts.Params) > 0 {
		params := url.Values{}
		for k, v := range opts.Params {
			params.Set(k, v)
		}
		reqURL += "?" + params.Encode()
	}

	// Build body
	var bodyReader io.Reader
	if opts.Body != nil {
		bodyBytes, err := json.Marshal(opts.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Retry loop
	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		// Create request
		req, err := http.NewRequest(method, reqURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Reset body reader for retries
		if opts.Body != nil {
			bodyBytes, _ := json.Marshal(opts.Body)
			bodyReader = bytes.NewReader(bodyBytes)
			req.Body = io.NopCloser(bodyReader)
		}

		// Set headers
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", fmt.Sprintf("crawl4ai-cloud/%s", Version))

		// Use custom timeout if provided
		client := c.client
		if opts.Timeout > 0 && opts.Timeout != c.timeout {
			client = &http.Client{Timeout: opts.Timeout}
		}

		// Make request
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.maxRetries-1 {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
				continue
			}
			return nil, NewTimeoutError(fmt.Sprintf("request failed: %v", err))
		}

		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			if attempt < c.maxRetries-1 {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
				continue
			}
			return nil, NewCloudError(fmt.Sprintf("failed to read response: %v", err), 0, nil, nil)
		}

		// Parse response
		var result map[string]interface{}
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, &result); err != nil {
				// Try to return as string if not JSON
				result = map[string]interface{}{"raw": string(respBody)}
			}
		} else {
			result = make(map[string]interface{})
		}

		// Success
		if resp.StatusCode < 400 {
			return result, nil
		}

		// Extract error detail
		detail := ""
		if d, ok := result["detail"].(string); ok {
			detail = d
		} else {
			detail = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}

		// Parse headers
		headers := make(map[string]string)
		for k, v := range resp.Header {
			if len(v) > 0 {
				headers[strings.ToLower(k)] = v[0]
			}
		}

		// Map status codes to errors
		switch resp.StatusCode {
		case 401:
			return nil, NewAuthenticationError(detail, result, headers)
		case 404:
			return nil, NewNotFoundError(detail, result, headers)
		case 429:
			if strings.Contains(strings.ToLower(detail), "rate limit") {
				return nil, NewRateLimitError(detail, result, headers)
			}
			return nil, NewQuotaExceededError(detail, result, headers)
		case 400:
			return nil, NewValidationError(detail, result, headers)
		case 504:
			return nil, NewTimeoutError(detail)
		default:
			if resp.StatusCode >= 500 {
				lastErr = NewServerError(detail, resp.StatusCode, result, headers)
				if attempt < c.maxRetries-1 {
					time.Sleep(time.Duration(1<<attempt) * time.Second)
					continue
				}
				return nil, lastErr
			}
			return nil, NewCloudError(detail, resp.StatusCode, result, headers)
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, NewCloudError("max retries exceeded", 0, nil, nil)
}

// Get makes a GET request.
func (c *HTTPClient) Get(path string, params map[string]string) (map[string]interface{}, error) {
	return c.Request(RequestOptions{
		Method: "GET",
		Path:   path,
		Params: params,
	})
}

// Post makes a POST request.
func (c *HTTPClient) Post(path string, body map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	return c.Request(RequestOptions{
		Method:  "POST",
		Path:    path,
		Body:    body,
		Timeout: timeout,
	})
}

// Delete makes a DELETE request.
func (c *HTTPClient) Delete(path string) (map[string]interface{}, error) {
	return c.Request(RequestOptions{
		Method: "DELETE",
		Path:   path,
	})
}
