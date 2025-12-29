package crawl4ai

import (
	"fmt"
	"strings"
)

// CrawlerRunConfig represents configuration for crawl requests.
type CrawlerRunConfig struct {
	// Content processing
	WordCountThreshold     int      `json:"word_count_threshold,omitempty"`
	ExcludeExternalLinks   bool     `json:"exclude_external_links,omitempty"`
	ExcludeSocialMediaLinks bool    `json:"exclude_social_media_links,omitempty"`
	ExcludeExternalImages  bool     `json:"exclude_external_images,omitempty"`
	ExcludeDomains         []string `json:"exclude_domains,omitempty"`

	// HTML processing
	ProcessIframes     bool `json:"process_iframes,omitempty"`
	RemoveForms        bool `json:"remove_forms,omitempty"`
	KeepDataAttributes bool `json:"keep_data_attributes,omitempty"`

	// Output options
	OnlyText  bool `json:"only_text,omitempty"`
	Prettiify bool `json:"prettiify,omitempty"`

	// Screenshot/PDF
	Screenshot        bool   `json:"screenshot,omitempty"`
	ScreenshotWaitFor string `json:"screenshot_wait_for,omitempty"`
	PDF               bool   `json:"pdf,omitempty"`

	// Wait conditions
	WaitFor              string  `json:"wait_for,omitempty"`
	DelayBeforeReturnHTML float64 `json:"delay_before_return_html,omitempty"`

	// Page interaction
	JsCode               string  `json:"js_code,omitempty"`
	JsOnly               bool    `json:"js_only,omitempty"`
	IgnoreBodyVisibility bool    `json:"ignore_body_visibility,omitempty"`
	ScanFullPage         bool    `json:"scan_full_page,omitempty"`
	ScrollDelay          float64 `json:"scroll_delay,omitempty"`

	// Network
	WaitForImages          bool `json:"wait_for_images,omitempty"`
	AdjustViewportToContent bool `json:"adjust_viewport_to_content,omitempty"`
	PageTimeout            int  `json:"page_timeout,omitempty"`

	// Magic mode
	Magic bool `json:"magic,omitempty"`

	// Simulate user
	SimulateUser      bool `json:"simulate_user,omitempty"`
	OverrideNavigator bool `json:"override_navigator,omitempty"`

	// Cache (cloud-controlled, will be stripped)
	CacheMode    string `json:"cache_mode,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	BypassCache  bool   `json:"bypass_cache,omitempty"`
	NoCacheRead  bool   `json:"no_cache_read,omitempty"`
	NoCacheWrite bool   `json:"no_cache_write,omitempty"`
	DisableCache bool   `json:"disable_cache,omitempty"`
}

// BrowserConfig represents browser configuration for crawl requests.
type BrowserConfig struct {
	// Browser settings
	Headless    bool   `json:"headless,omitempty"`
	BrowserType string `json:"browser_type,omitempty"`
	Verbose     bool   `json:"verbose,omitempty"`

	// Viewport
	ViewportWidth  int `json:"viewport_width,omitempty"`
	ViewportHeight int `json:"viewport_height,omitempty"`

	// User agent
	UserAgent     string `json:"user_agent,omitempty"`
	UserAgentMode string `json:"user_agent_mode,omitempty"`

	// Headers & cookies
	Headers map[string]string      `json:"headers,omitempty"`
	Cookies []map[string]interface{} `json:"cookies,omitempty"`

	// HTTPS errors
	IgnoreHTTPSErrors  bool `json:"ignore_https_errors,omitempty"`
	JavaScriptEnabled  bool `json:"java_script_enabled,omitempty"`

	// Text mode
	TextMode  bool `json:"text_mode,omitempty"`
	LightMode bool `json:"light_mode,omitempty"`

	// Cloud-controlled fields (will be stripped)
	CdpURL            string `json:"cdp_url,omitempty"`
	UseManagedBrowser bool   `json:"use_managed_browser,omitempty"`
	BrowserMode       string `json:"browser_mode,omitempty"`
	UserDataDir       string `json:"user_data_dir,omitempty"`
	ChromeChannel     string `json:"chrome_channel,omitempty"`
}

// crawlerConfigSanitizeFields are fields to remove from CrawlerRunConfig.
var crawlerConfigSanitizeFields = []string{
	"cache_mode",
	"session_id",
	"bypass_cache",
	"no_cache_read",
	"no_cache_write",
	"disable_cache",
}

// browserConfigSanitizeFields are fields to remove from BrowserConfig.
var browserConfigSanitizeFields = []string{
	"cdp_url",
	"use_managed_browser",
	"browser_mode",
	"user_data_dir",
	"chrome_channel",
}

// SanitizeCrawlerConfig removes cloud-controlled fields from config.
func SanitizeCrawlerConfig(config *CrawlerRunConfig) map[string]interface{} {
	if config == nil {
		return nil
	}

	result := make(map[string]interface{})

	// Add non-zero values
	if config.WordCountThreshold > 0 {
		result["word_count_threshold"] = config.WordCountThreshold
	}
	if config.ExcludeExternalLinks {
		result["exclude_external_links"] = true
	}
	if config.ExcludeSocialMediaLinks {
		result["exclude_social_media_links"] = true
	}
	if config.ExcludeExternalImages {
		result["exclude_external_images"] = true
	}
	if len(config.ExcludeDomains) > 0 {
		result["exclude_domains"] = config.ExcludeDomains
	}
	if config.ProcessIframes {
		result["process_iframes"] = true
	}
	if config.RemoveForms {
		result["remove_forms"] = true
	}
	if config.KeepDataAttributes {
		result["keep_data_attributes"] = true
	}
	if config.OnlyText {
		result["only_text"] = true
	}
	if config.Prettiify {
		result["prettiify"] = true
	}
	if config.Screenshot {
		result["screenshot"] = true
	}
	if config.ScreenshotWaitFor != "" {
		result["screenshot_wait_for"] = config.ScreenshotWaitFor
	}
	if config.PDF {
		result["pdf"] = true
	}
	if config.WaitFor != "" {
		result["wait_for"] = config.WaitFor
	}
	if config.DelayBeforeReturnHTML > 0 {
		result["delay_before_return_html"] = config.DelayBeforeReturnHTML
	}
	if config.JsCode != "" {
		result["js_code"] = config.JsCode
	}
	if config.JsOnly {
		result["js_only"] = true
	}
	if config.IgnoreBodyVisibility {
		result["ignore_body_visibility"] = true
	}
	if config.ScanFullPage {
		result["scan_full_page"] = true
	}
	if config.ScrollDelay > 0 {
		result["scroll_delay"] = config.ScrollDelay
	}
	if config.WaitForImages {
		result["wait_for_images"] = true
	}
	if config.AdjustViewportToContent {
		result["adjust_viewport_to_content"] = true
	}
	if config.PageTimeout > 0 {
		result["page_timeout"] = config.PageTimeout
	}
	if config.Magic {
		result["magic"] = true
	}
	if config.SimulateUser {
		result["simulate_user"] = true
	}
	if config.OverrideNavigator {
		result["override_navigator"] = true
	}

	// Note: cache fields are NOT added (sanitized)

	if len(result) == 0 {
		return nil
	}
	return result
}

// SanitizeBrowserConfig removes cloud-controlled fields from config.
func SanitizeBrowserConfig(config *BrowserConfig, strategy string) map[string]interface{} {
	if config == nil {
		return nil
	}

	// Warn if browser config with HTTP strategy
	if strategy == "http" {
		return nil
	}

	result := make(map[string]interface{})

	if config.Headless {
		result["headless"] = true
	}
	if config.BrowserType != "" {
		result["browser_type"] = config.BrowserType
	}
	if config.Verbose {
		result["verbose"] = true
	}
	if config.ViewportWidth > 0 {
		result["viewport_width"] = config.ViewportWidth
	}
	if config.ViewportHeight > 0 {
		result["viewport_height"] = config.ViewportHeight
	}
	if config.UserAgent != "" {
		result["user_agent"] = config.UserAgent
	}
	if config.UserAgentMode != "" {
		result["user_agent_mode"] = config.UserAgentMode
	}
	if len(config.Headers) > 0 {
		result["headers"] = config.Headers
	}
	if len(config.Cookies) > 0 {
		result["cookies"] = config.Cookies
	}
	if config.IgnoreHTTPSErrors {
		result["ignore_https_errors"] = true
	}
	if config.JavaScriptEnabled {
		result["java_script_enabled"] = true
	}
	if config.TextMode {
		result["text_mode"] = true
	}
	if config.LightMode {
		result["light_mode"] = true
	}

	// Note: CDP fields are NOT added (sanitized)

	if len(result) == 0 {
		return nil
	}
	return result
}

// NormalizeProxy converts proxy input to map format.
func NormalizeProxy(proxy interface{}) (map[string]interface{}, error) {
	if proxy == nil {
		return nil, nil
	}

	switch p := proxy.(type) {
	case string:
		return map[string]interface{}{"mode": p}, nil
	case *ProxyConfig:
		result := map[string]interface{}{"mode": p.Mode}
		if p.Country != "" {
			result["country"] = p.Country
		}
		if p.StickySession {
			result["sticky_session"] = true
		}
		return result, nil
	case ProxyConfig:
		result := map[string]interface{}{"mode": p.Mode}
		if p.Country != "" {
			result["country"] = p.Country
		}
		if p.StickySession {
			result["sticky_session"] = true
		}
		return result, nil
	case map[string]interface{}:
		return p, nil
	default:
		return nil, fmt.Errorf("invalid proxy type: %T", proxy)
	}
}

// BuildCrawlRequest builds a crawl request body for the API.
func BuildCrawlRequest(options map[string]interface{}) map[string]interface{} {
	body := make(map[string]interface{})

	// Set strategy
	if strategy, ok := options["strategy"].(string); ok {
		body["strategy"] = strategy
	} else {
		body["strategy"] = "browser"
	}

	// URL(s)
	if url, ok := options["url"].(string); ok {
		body["url"] = url
	}
	if urls, ok := options["urls"].([]string); ok {
		body["urls"] = urls
	}

	// Config
	if config, ok := options["config"].(*CrawlerRunConfig); ok {
		if sanitized := SanitizeCrawlerConfig(config); sanitized != nil {
			body["crawler_config"] = sanitized
		}
	}

	// Browser config
	strategy := "browser"
	if s, ok := options["strategy"].(string); ok {
		strategy = s
	}
	if browserConfig, ok := options["browserConfig"].(*BrowserConfig); ok {
		if sanitized := SanitizeBrowserConfig(browserConfig, strategy); sanitized != nil {
			body["browser_config"] = sanitized
		}
	}

	// Proxy
	if proxy := options["proxy"]; proxy != nil {
		if proxyMap, err := NormalizeProxy(proxy); err == nil && proxyMap != nil {
			body["proxy"] = proxyMap
		}
	}

	// Bypass cache
	if bypassCache, ok := options["bypassCache"].(bool); ok && bypassCache {
		body["bypass_cache"] = true
	}

	// Priority
	if priority, ok := options["priority"].(int); ok {
		body["priority"] = priority
	}

	// Webhook URL
	if webhookURL, ok := options["webhookUrl"].(string); ok && webhookURL != "" {
		body["webhook_url"] = webhookURL
	}

	return body
}

// toSnakeCase converts a camelCase string to snake_case.
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32) // lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
