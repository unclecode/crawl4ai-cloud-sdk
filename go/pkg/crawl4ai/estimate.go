package crawl4ai

import (
	"fmt"
	"time"
)

// EstimateLineItem is one priced line in a cost Estimate. Credit lines carry
// Credits; token-forecast lines carry Tokens (post-pay, never added to credits).
type EstimateLineItem struct {
	Service   string                 `json:"service"`
	Action    string                 `json:"action"`
	Count     string                 `json:"count"`
	Credits   *string                `json:"credits"`
	Tokens    *int                   `json:"tokens"`
	Phase     string                 `json:"phase"`
	Exact     bool                   `json:"exact"`
	Modifiers map[string]interface{} `json:"modifiers"`
	Note      string                 `json:"note"`
}

// Estimate is the cost preview returned by Estimate() — a dry-run of a service
// request. The server validates + estimates with no execution, no job, and no
// charge. Credits are reservable and charged up-front; TokenForecast is an
// informational LLM-token magnitude billed post-pay (never added to Credits).
type Estimate struct {
	Service          string             `json:"service"`
	Credits          string             `json:"credits"`            // reservable credit cost (string decimal)
	CreditsExact     bool               `json:"credits_exact"`      // false = worst-case upper bound
	Breakdown        []EstimateLineItem `json:"breakdown"`          // credit line items
	TokenForecast    []EstimateLineItem `json:"token_forecast"`     // informational LLM tokens (post-pay)
	CoveredByBalance bool               `json:"covered_by_balance"` // spendable credit >= credits
	ChunksNeeded     int                `json:"chunks_needed"`      // ceil(credits / 25)
	SpendableCredit  string             `json:"spendable_credit"`   // live balance, else plan grant
	DryRun           bool               `json:"dry_run"`
}

// estimatePaths maps a service name to its dry-run-capable endpoint.
var estimatePaths = map[string]string{
	"crawl":      "/v1/crawl",
	"scrape":     "/v1/scrape",
	"screenshot": "/v1/screenshot",
	"extract":    "/v1/extract",
	"scan":       "/v1/scan",
	"site":       "/v1/site",
	"enrich":     "/v1/enrich/async",
	"context":    "/v1/context",
	"schema":     "/v1/schema/generate",
	"discovery":  "/v1/discovery/search",
}

// Estimate previews the cost of a request without running it. Pass a service
// name ("crawl", "scrape", "screenshot", "extract", "scan", "site", "enrich",
// "context", "schema", "discovery") or a full "/v1/..." path, plus the same
// params you would send to the real call. The request is validated and a cost
// Estimate is returned; nothing executes, no job is created, and no credits
// are charged.
//
//	est, err := crawler.Estimate("scrape", map[string]interface{}{"url": "https://example.com"})
//	// Non-default discovery vertical → pass the full path:
//	est, err := crawler.Estimate("/v1/discovery/search", map[string]interface{}{"query": "..."})
func (c *AsyncWebCrawler) Estimate(service string, params map[string]interface{}) (*Estimate, error) {
	path := service
	if len(service) == 0 || service[0] != '/' {
		p, ok := estimatePaths[service]
		if !ok {
			return nil, fmt.Errorf(
				"Estimate: unknown service %q — pass a known service name "+
					"(crawl, scrape, screenshot, extract, scan, site, enrich, "+
					"context, schema, discovery) or a full \"/v1/...\" path", service)
		}
		path = p
	}

	body := make(map[string]interface{}, len(params)+1)
	for k, v := range params {
		body[k] = v
	}
	body["dry_run"] = true

	data, err := c.http.Post(path, body, 60*time.Second)
	if err != nil {
		return nil, err
	}
	return unmarshalWrapper[Estimate](data)
}
