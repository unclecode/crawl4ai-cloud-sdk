// Enrich v2 E2E tests — runs against stage.crawl4ai.com.
//
// Covers the seven-method surface:
//
//	Enrich(...)         POST /v1/enrich/async
//	GetEnrichJob        GET  /v1/enrich/jobs/{id}
//	WaitEnrichJob       poll loop, optional Until phase
//	ResumeEnrichJob     POST /v1/enrich/jobs/{id}/continue
//	StreamEnrichJob     GET  /v1/enrich/jobs/{id}/stream  (SSE)
//	CancelEnrichJob     DELETE /v1/enrich/jobs/{id}
//	ListEnrichJobs      GET  /v1/enrich/jobs
//
// Run:
//
//	go test ./pkg/crawl4ai/ -run TestEnrich -v -timeout 10m
package crawl4ai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

var enrichTestKey = getEnrichTestKey()

func getEnrichTestKey() string {
	if key := os.Getenv("CRAWL4AI_API_KEY"); key != "" {
		return key
	}
	return "sk_live_V89kxHtmkxw0jJORu_sWzyuvGw6TKHaJhoNGK8gGdqU"
}

func getEnrichBaseURL() string {
	if u := os.Getenv("CRAWL4AI_BASE_URL"); u != "" {
		return u
	}
	return "https://stage.crawl4ai.com"
}

func setupEnrich(t *testing.T) *AsyncWebCrawler {
	t.Helper()
	c, err := NewAsyncWebCrawler(CrawlerOptions{
		APIKey:  enrichTestKey,
		BaseURL: getEnrichBaseURL(),
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return c
}

// ─── 1. URLs-only mode (fastest, deterministic) ──────────────────────────

func TestEnrichV2_URLsOnly_SingleURL(t *testing.T) {
	c := setupEnrich(t)
	result, err := c.Enrich(&EnrichOptions{
		URLs: []string{"https://kidocode.com"},
		Features: []EnrichFeature{
			{Name: "company_name"},
			{Name: "contact_email", Description: "primary contact email"},
		},
		Strategy: "http",
		Wait:     true,
		Timeout:  3 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if !result.IsComplete() {
		t.Fatalf("job not terminal: %s (%s)", result.Status, result.Error)
	}
	if !result.IsSuccessful() {
		t.Fatalf("job ended in %s: %s", result.Status, result.Error)
	}
	if len(result.PhaseData.Rows) < 1 {
		t.Fatalf("expected at least 1 row, got %d", len(result.PhaseData.Rows))
	}
	row := result.PhaseData.Rows[0]
	if len(row.Fields) == 0 {
		t.Fatalf("no fields extracted (error=%q)", row.Error)
	}
	hasCompany := false
	for k := range row.Fields {
		if strings.HasPrefix(strings.ToLower(k), "company") {
			hasCompany = true
			break
		}
	}
	if !hasCompany {
		t.Fatalf("company_name missing from row.Fields=%v", row.Fields)
	}
	// Usage envelope
	if result.Usage.Crawls < 1 {
		t.Fatalf("expected crawls>=1, got %d", result.Usage.Crawls)
	}
	bucket, ok := result.Usage.LlmTokensByPurpose["extract"]
	if !ok {
		t.Fatalf("missing extract usage bucket; have %v", keys(result.Usage.LlmTokensByPurpose))
	}
	if bucket.Input <= 0 || bucket.Output <= 0 {
		t.Fatalf("extract bucket empty: input=%d output=%d", bucket.Input, bucket.Output)
	}
}

func TestEnrichV2_URLsOnly_StringFeatures(t *testing.T) {
	c := setupEnrich(t)
	result, err := c.Enrich(&EnrichOptions{
		URLs:     []string{"https://example.com"},
		Features: []EnrichFeature{{Name: "title"}, {Name: "description"}},
		Strategy: "http",
		Wait:     true,
		Timeout:  2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if !result.IsComplete() {
		t.Fatalf("not terminal: %s", result.Status)
	}
	if len(result.PhaseData.Rows) < 1 {
		t.Fatalf("expected at least 1 row, got %d", len(result.PhaseData.Rows))
	}
}

// ─── 2. Job lifecycle ────────────────────────────────────────────────────

func TestEnrichV2_Lifecycle_FireAndForget(t *testing.T) {
	c := setupEnrich(t)
	job, err := c.Enrich(&EnrichOptions{
		URLs:     []string{"https://kidocode.com"},
		Features: []EnrichFeature{{Name: "company_name"}},
		Strategy: "http",
		Wait:     false,
	})
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if !strings.HasPrefix(job.JobID, "enr_") {
		t.Fatalf("bad job_id: %q", job.JobID)
	}
	latest, err := c.GetEnrichJob(job.JobID)
	if err != nil {
		t.Fatalf("GetEnrichJob: %v", err)
	}
	if latest.JobID != job.JobID {
		t.Fatalf("job id mismatch: %q vs %q", latest.JobID, job.JobID)
	}
}

func TestEnrichV2_Lifecycle_WaitToTerminal(t *testing.T) {
	c := setupEnrich(t)
	job, err := c.Enrich(&EnrichOptions{
		URLs:     []string{"https://example.com"},
		Features: []EnrichFeature{{Name: "title"}},
		Strategy: "http",
		Wait:     false,
	})
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	terminal, err := c.WaitEnrichJob(job.JobID, WaitEnrichOptions{Timeout: 2 * time.Minute})
	if err != nil {
		t.Fatalf("WaitEnrichJob: %v", err)
	}
	if !terminal.IsComplete() || !terminal.IsSuccessful() {
		t.Fatalf("not successful: %s %s", terminal.Status, terminal.Error)
	}
}

func TestEnrichV2_Lifecycle_List(t *testing.T) {
	c := setupEnrich(t)
	jobs, err := c.ListEnrichJobs(5, 0)
	if err != nil {
		t.Fatalf("ListEnrichJobs: %v", err)
	}
	if len(jobs) < 1 {
		t.Fatalf("expected >=1 job from earlier tests, got %d", len(jobs))
	}
	for _, j := range jobs {
		if !strings.HasPrefix(j.JobID, "enr_") {
			t.Fatalf("bad job_id in list: %q", j.JobID)
		}
	}
}

func TestEnrichV2_Lifecycle_Cancel(t *testing.T) {
	c := setupEnrich(t)
	job, err := c.Enrich(&EnrichOptions{
		Query:         "top BBQ restaurants in Austin Texas with outdoor seating",
		Country:       "us",
		TopKPerEntity: 2,
		Wait:          false,
	})
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if err := c.CancelEnrichJob(job.JobID); err != nil {
		t.Fatalf("CancelEnrichJob: %v", err)
	}
	time.Sleep(2 * time.Second)
	latest, err := c.GetEnrichJob(job.JobID)
	if err != nil {
		t.Fatalf("GetEnrichJob: %v", err)
	}
	if latest.Status != EnrichStatusCancelled {
		t.Fatalf("expected cancelled, got %s", latest.Status)
	}
}

// ─── 3. Review flow: pause + resume with edits ───────────────────────────

func TestEnrichV2_Review_PauseAndResume(t *testing.T) {
	c := setupEnrich(t)
	autoPlanFalse := false
	autoUrlsTrue := true
	job, err := c.Enrich(&EnrichOptions{
		Query:           "best Italian restaurants in Brooklyn New York",
		Country:         "us",
		TopKPerEntity:   1,
		AutoConfirmPlan: &autoPlanFalse,
		AutoConfirmURLs: &autoUrlsTrue,
		Wait:            false,
	})
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	paused, err := c.WaitEnrichJob(job.JobID, WaitEnrichOptions{
		Until:   EnrichStatusPlanReady,
		Timeout: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("WaitEnrichJob until plan_ready: %v", err)
	}
	if paused.Status != EnrichStatusPlanReady {
		t.Fatalf("expected plan_ready, got %s", paused.Status)
	}
	if paused.PhaseData.Plan == nil {
		t.Fatalf("plan should be populated at plan_ready")
	}
	if len(paused.PhaseData.Plan.Entities) < 1 {
		t.Fatalf("expected entities, got 0")
	}
	if _, ok := paused.Usage.LlmTokensByPurpose["plan_intent"]; !ok {
		t.Fatalf("expected plan_intent usage bucket")
	}

	// Trim to one entity but keep 3 features so we're not flaky when a
	// single page legitimately lacks one specific column.
	editedEntities := []EnrichEntity{{Name: paused.PhaseData.Plan.Entities[0].Name}}
	editedFeatures := []EnrichFeature{
		{Name: "name", Description: "restaurant name"},
		{Name: "address"},
		{Name: "phone"},
	}
	resumed, err := c.ResumeEnrichJob(job.JobID, &ResumeEnrichOptions{
		Entities: editedEntities,
		Features: editedFeatures,
	})
	if err != nil {
		t.Fatalf("ResumeEnrichJob: %v", err)
	}
	if resumed.Status == EnrichStatusPlanReady {
		t.Fatalf("expected past plan_ready after resume, still %s", resumed.Status)
	}
	final, err := c.WaitEnrichJob(job.JobID, WaitEnrichOptions{Timeout: 5 * time.Minute})
	if err != nil {
		t.Fatalf("WaitEnrichJob terminal: %v", err)
	}
	if !final.IsComplete() {
		t.Fatalf("expected terminal, got %s", final.Status)
	}
	if len(final.PhaseData.Rows) > 1 {
		t.Fatalf("expected ≤1 row after edit, got %d", len(final.PhaseData.Rows))
	}

	// AT LEAST ONE row should have populated fields — would have caught the
	// worker P0 (features-not-plumbed) where the job completed but every
	// row had Fields={}.
	populated := 0
	for _, r := range final.PhaseData.Rows {
		if len(r.Fields) > 0 {
			populated++
		}
	}
	if populated == 0 && len(final.PhaseData.Rows) > 0 {
		t.Fatalf("all %d rows have empty fields — likely a regression of features-not-plumbed",
			len(final.PhaseData.Rows))
	}

	// ── tier + reason on URL candidates (added in 0.6.1) ──
	// LLM rerank should have populated these on every URL that made it past
	// the resolve phase.
	for entity, urls := range final.PhaseData.URLsPerEntity {
		for _, c := range urls {
			if c.Tier == nil {
				t.Errorf("missing Tier on URL %s/%s", entity, c.URL)
				continue
			}
			if *c.Tier < 0.0 || *c.Tier > 1.0 {
				t.Errorf("invalid Tier %f on %s/%s", *c.Tier, entity, c.URL)
			}
			if c.Reason == nil || *c.Reason == "" {
				t.Errorf("missing Reason on URL %s/%s", entity, c.URL)
			}
		}
	}
}

// ─── 4. SSE stream ───────────────────────────────────────────────────────

func TestEnrichV2_Stream_SnapshotAndComplete(t *testing.T) {
	c := setupEnrich(t)
	job, err := c.Enrich(&EnrichOptions{
		URLs:     []string{"https://example.com"},
		Features: []EnrichFeature{{Name: "title"}},
		Strategy: "http",
		Wait:     false,
	})
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	events, err := c.StreamEnrichJob(ctx, job.JobID)
	if err != nil {
		t.Fatalf("StreamEnrichJob: %v", err)
	}

	seen := map[string]bool{}
	for evt := range events {
		seen[evt.Type] = true
		if evt.Type == "snapshot" && evt.Snapshot == nil {
			t.Errorf("snapshot event missing parsed snapshot")
		}
		if evt.Type == "snapshot" && evt.Snapshot != nil && evt.Snapshot.JobID != job.JobID {
			t.Errorf("snapshot job_id mismatch: %q", evt.Snapshot.JobID)
		}
		if evt.Type == "complete" {
			break
		}
	}
	if !seen["snapshot"] {
		t.Fatalf("missing snapshot event; saw %v", keys(seen))
	}
	if !seen["complete"] {
		t.Fatalf("stream never emitted complete; saw %v", keys(seen))
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
