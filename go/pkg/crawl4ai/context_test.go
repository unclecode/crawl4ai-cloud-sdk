// Context v2 SDK tests — unit + live against stage.
//
// Run:
//
//	# All
//	go test ./pkg/crawl4ai/ -run TestContext -v -timeout 10m
//
//	# Unit only
//	CRAWL4AI_SKIP_LIVE=1 go test ./pkg/crawl4ai/ -run TestContext -v
package crawl4ai

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

// ─── shared helpers ─────────────────────────────────────────────────────

func getContextTestKey() string {
	if key := os.Getenv("CRAWL4AI_API_KEY"); key != "" {
		return key
	}
	return "sk_live_V89kxHtmkxw0jJORu_sWzyuvGw6TKHaJhoNGK8gGdqU"
}

func getContextBaseURL() string {
	if u := os.Getenv("CRAWL4AI_BASE_URL"); u != "" {
		return u
	}
	return "https://stage.crawl4ai.com"
}

func setupContext(t *testing.T) *AsyncWebCrawler {
	t.Helper()
	c, err := NewAsyncWebCrawler(CrawlerOptions{
		APIKey:  getContextTestKey(),
		BaseURL: getContextBaseURL(),
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return c
}

func skipIfLiveDisabled(t *testing.T) {
	t.Helper()
	if os.Getenv("CRAWL4AI_SKIP_LIVE") == "1" {
		t.Skip("CRAWL4AI_SKIP_LIVE=1 — skipping live test")
	}
}

func boolPtr(b bool) *bool { return &b }

// ─── Unit — Source builders ─────────────────────────────────────────────

func TestContext_Unit_GoogleWebSourceDefaults(t *testing.T) {
	out := GoogleWebSource(nil)
	if out.Type != "google_web" {
		t.Fatalf("type: %q", out.Type)
	}
	if out.Params["top_k_per_backend"] != 10 {
		t.Fatalf("top_k: %v", out.Params["top_k_per_backend"])
	}
}

func TestContext_Unit_GoogleWebSourceFull(t *testing.T) {
	out := GoogleWebSource(&GoogleWebSourceOptions{
		Backends:       []string{"google", "bing"},
		TopKPerBackend: 8,
		Region:         "us",
	})
	if !reflect.DeepEqual(out.Params["backends"], []string{"google", "bing"}) {
		t.Fatalf("backends: %v", out.Params["backends"])
	}
}

func TestContext_Unit_GoogleDriveSourceDefault(t *testing.T) {
	out, err := GoogleDriveSource(GoogleDriveSourceOptions{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.Type != "google_drive" || out.Params["mode"] != "search" || out.Params["folder_id"] != "" {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestContext_Unit_GoogleDriveSourceFolder(t *testing.T) {
	out, err := GoogleDriveSource(GoogleDriveSourceOptions{Mode: "folder", FolderID: "abc"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.Params["mode"] != "folder" || out.Params["folder_id"] != "abc" {
		t.Fatalf("unexpected: %+v", out.Params)
	}
}

func TestContext_Unit_GoogleDriveSourceFolderRequiresID(t *testing.T) {
	_, err := GoogleDriveSource(GoogleDriveSourceOptions{Mode: "folder"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestContext_Unit_GoogleDriveSourceBadMode(t *testing.T) {
	_, err := GoogleDriveSource(GoogleDriveSourceOptions{Mode: "bogus"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestContext_Unit_GmailSourceDefault(t *testing.T) {
	out, err := GmailSource(GmailSourceOptions{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.Type != "gmail" || out.Params["mode"] != "search" || out.Params["include_spam_trash"] != false {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestContext_Unit_GmailSourceLabel(t *testing.T) {
	out, err := GmailSource(GmailSourceOptions{
		Mode:             "label",
		LabelID:          "Label_42",
		After:            "2026/01/01",
		Before:           "2026/05/01",
		IncludeSpamTrash: true,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out.Params["label_id"] != "Label_42" || out.Params["after"] != "2026/01/01" {
		t.Fatalf("unexpected: %+v", out.Params)
	}
}

func TestContext_Unit_GmailSourceLabelRequiresID(t *testing.T) {
	_, err := GmailSource(GmailSourceOptions{Mode: "label"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestContext_Unit_CrawlSource(t *testing.T) {
	out := CrawlSource(CrawlSourceOptions{
		Domain: "https://example.com", MaxURLs: 30, MaxDepth: 2,
		ScoreThreshold: 0.5, ProfileID: "my-profile",
	})
	if out.Params["domain"] != "https://example.com" || out.Params["max_urls"] != 30 ||
		out.Params["score_threshold"] != 0.5 || out.Params["profile_id"] != "my-profile" {
		t.Fatalf("unexpected: %+v", out.Params)
	}
}

func TestContext_Unit_FileSource(t *testing.T) {
	out := FileSource(FileSourceOptions{FileID: "file_abc"})
	if out.Params["file_id"] != "file_abc" || out.Params["chunk_size"] != 2000 {
		t.Fatalf("unexpected: %+v", out.Params)
	}
}

func TestContext_Unit_CustomSource(t *testing.T) {
	out := CustomSource(CustomSourceOptions{
		Type: "hackernews", Params: map[string]interface{}{"tag": "ai"}, AuthRef: "link_x",
	})
	if out.AuthRef != "link_x" || out.Params["tag"] != "ai" {
		t.Fatalf("unexpected: %+v", out)
	}
}

// ─── Unit — Strategy builders ───────────────────────────────────────────

func TestContext_Unit_AllItemsStrategy(t *testing.T) {
	out := AllItemsStrategy()
	if out.Type != "all_items" || len(out.Params) != 0 {
		t.Fatalf("unexpected: %+v", out)
	}
}

func TestContext_Unit_LLMRerankStrategyDefaults(t *testing.T) {
	out := LLMRerankStrategy(LLMRerankStrategyOptions{})
	if out.Type != "llm_rerank" {
		t.Fatalf("type: %q", out.Type)
	}
	if out.Params["top_n"] != 0 || out.Params["batch_size"] != 20 || out.Params["max_concurrency"] != 4 ||
		out.Params["content_aware"] != false || out.Params["content_chars"] != 4000 {
		t.Fatalf("defaults wrong: %+v", out.Params)
	}
}

func TestContext_Unit_LLMRerankStrategyFull(t *testing.T) {
	out := LLMRerankStrategy(LLMRerankStrategyOptions{
		TopN: 5, Instruction: "Prefer official docs.",
		Model: "anthropic/claude-sonnet-4-6",
		ScoreThreshold: 0.3, ContentAware: true, ContentChars: 6000,
	})
	if out.Params["top_n"] != 5 || out.Params["instruction"] != "Prefer official docs." ||
		out.Params["score_threshold"] != 0.3 || out.Params["content_aware"] != true ||
		out.Params["content_chars"] != 6000 {
		t.Fatalf("unexpected: %+v", out.Params)
	}
}

// ─── Unit — Synthesizer builders ────────────────────────────────────────

func TestContext_Unit_RawSynthesizer(t *testing.T) {
	out := RawSynthesizer()
	if out.Type != "raw" {
		t.Fatalf("type: %q", out.Type)
	}
	// Back-compat alias.
	if RawShape().Type != "raw" {
		t.Fatalf("RawShape alias broken")
	}
}

func TestContext_Unit_MarkdownSynthesizerSingleDefaults(t *testing.T) {
	out, err := MarkdownSynthesizer(MarkdownSynthesizerOptions{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out.Type != "markdown" || out.Params["mode"] != "single" ||
		out.Params["batch_size"] != 5 || out.Params["include_metadata"] != true ||
		out.Params["max_chars_per_item"] != 20000 {
		t.Fatalf("defaults wrong: %+v", out.Params)
	}
}

func TestContext_Unit_MarkdownSynthesizerMulti(t *testing.T) {
	out, err := MarkdownSynthesizer(MarkdownSynthesizerOptions{
		Mode: "multi", Instruction: "Summarise.",
		IncludeMetadata: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out.Params["mode"] != "multi" || out.Params["instruction"] != "Summarise." ||
		out.Params["include_metadata"] != false {
		t.Fatalf("unexpected: %+v", out.Params)
	}
}

func TestContext_Unit_MarkdownSynthesizerBadMode(t *testing.T) {
	_, err := MarkdownSynthesizer(MarkdownSynthesizerOptions{Mode: "bogus"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestContext_Unit_LLMSynthesizerByExample(t *testing.T) {
	out, err := LLMSynthesizer(LLMSynthesizerOptions{
		Instruction: "extract a knowledge graph",
		Example:     map[string]interface{}{"nodes": []interface{}{map[string]interface{}{"id": 1}}},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out.Type != "llm" || out.Params["instruction"] != "extract a knowledge graph" {
		t.Fatalf("unexpected: %+v", out.Params)
	}
	// Example is JSON-marshalled to a string.
	exampleStr, ok := out.Params["output_example"].(string)
	if !ok {
		t.Fatalf("output_example not string: %T", out.Params["output_example"])
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(exampleStr), &decoded); err != nil {
		t.Fatalf("output_example not valid JSON: %v", err)
	}
	if out.Params["output_schema"] != "" || out.Params["output_description"] != "" {
		t.Fatalf("non-example fields should be blank: %+v", out.Params)
	}
}

func TestContext_Unit_LLMSynthesizerBySchemaDict(t *testing.T) {
	out, err := LLMSynthesizer(LLMSynthesizerOptions{
		Instruction: "tabulate",
		Schema:      map[string]interface{}{"type": "object"},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	schemaStr, _ := out.Params["output_schema"].(string)
	var decoded map[string]interface{}
	_ = json.Unmarshal([]byte(schemaStr), &decoded)
	if decoded["type"] != "object" {
		t.Fatalf("schema not round-tripped: %v", schemaStr)
	}
}

func TestContext_Unit_LLMSynthesizerBySchemaString(t *testing.T) {
	out, err := LLMSynthesizer(LLMSynthesizerOptions{
		Instruction: "x", Schema: `{"type":"object"}`,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out.Params["output_schema"] != `{"type":"object"}` {
		t.Fatalf("string schema should pass through: %v", out.Params["output_schema"])
	}
}

func TestContext_Unit_LLMSynthesizerByDescription(t *testing.T) {
	out, err := LLMSynthesizer(LLMSynthesizerOptions{
		Instruction: "summarise", Description: "An object with title and body.",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if out.Params["output_description"] != "An object with title and body." {
		t.Fatalf("description missing: %v", out.Params["output_description"])
	}
}

func TestContext_Unit_LLMSynthesizerRequiresInstruction(t *testing.T) {
	_, err := LLMSynthesizer(LLMSynthesizerOptions{Example: map[string]interface{}{"a": 1}})
	if err == nil {
		t.Fatalf("expected instruction error")
	}
}

func TestContext_Unit_LLMSynthesizerExactlyOneInput(t *testing.T) {
	_, err := LLMSynthesizer(LLMSynthesizerOptions{Instruction: "x"})
	if err == nil {
		t.Fatalf("expected 'exactly one' error (no inputs)")
	}
	_, err = LLMSynthesizer(LLMSynthesizerOptions{
		Instruction: "x",
		Schema:      map[string]interface{}{},
		Example:     map[string]interface{}{},
	})
	if err == nil {
		t.Fatalf("expected 'exactly one' error (two inputs)")
	}
}

// ─── Unit — Reconciler builders ─────────────────────────────────────────

func TestContext_Unit_NoopReconciler(t *testing.T) {
	if NoopReconciler().Type != "noop" {
		t.Fatalf("type wrong")
	}
}

func TestContext_Unit_CustomReconcilerCron(t *testing.T) {
	out := CustomReconciler("cron", map[string]interface{}{"schedule": "0 6 * * *"})
	if out.Type != "cron" || out.Params["schedule"] != "0 6 * * *" {
		t.Fatalf("unexpected: %+v", out)
	}
}

// ─── Unit — Constraints ─────────────────────────────────────────────────

func TestContext_Unit_ConstraintsDefaults(t *testing.T) {
	out := ContextConstraints{}.ToMap()
	if out["max_items"] != 20 || out["max_per_source"] != 10 ||
		out["max_crawl_time_s"] != float64(120) || out["language"] != "en" {
		t.Fatalf("defaults wrong: %+v", out)
	}
	if _, present := out["freshness_days"]; present {
		t.Fatalf("freshness_days should be absent by default")
	}
}

func TestContext_Unit_ConstraintsFreshnessEmits(t *testing.T) {
	out := ContextConstraints{FreshnessDays: 7}.ToMap()
	if out["freshness_days"] != 7 {
		t.Fatalf("freshness_days: %v", out["freshness_days"])
	}
}

// ─── Unit — body composition (inline pipeline) ──────────────────────────

func TestContext_Unit_BodyMinimal(t *testing.T) {
	c := &AsyncWebCrawler{}
	body, err := c.buildContextBody(ContextOptions{Intent: "x"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["intent"] != "x" {
		t.Fatalf("intent missing")
	}
	if _, has := body["pipeline"]; has {
		t.Fatalf("pipeline should be absent on minimal body")
	}
}

func TestContext_Unit_BodyInlinePipelineBasic(t *testing.T) {
	c := &AsyncWebCrawler{}
	src := GoogleWebSource(nil)
	body, err := c.buildContextBody(ContextOptions{
		Intent: "x", Sources: []PillarConfig{src},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	pipeline, ok := body["pipeline"].(map[string]interface{})
	if !ok {
		t.Fatalf("pipeline missing: %+v", body)
	}
	srcs, ok := pipeline["sources"].([]map[string]interface{})
	if !ok {
		t.Fatalf("sources missing/wrong-type: %T", pipeline["sources"])
	}
	if len(srcs) != 1 || srcs[0]["type"] != "google_web" {
		t.Fatalf("source mismatch: %+v", srcs)
	}
}

func TestContext_Unit_BodyInlinePipelineFull(t *testing.T) {
	c := &AsyncWebCrawler{}
	src := GoogleWebSource(nil)
	strat := LLMRerankStrategy(LLMRerankStrategyOptions{TopN: 5, Instruction: "prefer docs"})
	syn, _ := MarkdownSynthesizer(MarkdownSynthesizerOptions{Mode: "single"})
	rec := NoopReconciler()
	body, err := c.buildContextBody(ContextOptions{
		Intent:      "x",
		Sources:     []PillarConfig{src},
		Strategy:    &strat,
		Synthesizer: &syn,
		Reconciler:  &rec,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	pipeline := body["pipeline"].(map[string]interface{})
	if pipeline["strategy"] != "llm_rerank" {
		t.Fatalf("strategy mismatch: %v", pipeline["strategy"])
	}
	if pipeline["synthesizer"] != "markdown" {
		t.Fatalf("synthesizer mismatch: %v", pipeline["synthesizer"])
	}
	synthParams := pipeline["synthesizer_params"].(map[string]interface{})
	if synthParams["mode"] != "single" {
		t.Fatalf("synth mode: %v", synthParams["mode"])
	}
	if pipeline["reconciler"] != "noop" {
		t.Fatalf("reconciler: %v", pipeline["reconciler"])
	}
}

func TestContext_Unit_BodyShapeAliasAccepted(t *testing.T) {
	c := &AsyncWebCrawler{}
	src := GoogleWebSource(nil)
	sh := RawSynthesizer()
	body, err := c.buildContextBody(ContextOptions{
		Intent:  "x",
		Sources: []PillarConfig{src},
		Shape:   &sh,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	pipeline := body["pipeline"].(map[string]interface{})
	if pipeline["synthesizer"] != "raw" {
		t.Fatalf("shape alias not mapped: %+v", pipeline)
	}
}

func TestContext_Unit_BodySynthesizerWinsOverShape(t *testing.T) {
	c := &AsyncWebCrawler{}
	src := GoogleWebSource(nil)
	syn, _ := MarkdownSynthesizer(MarkdownSynthesizerOptions{Mode: "multi"})
	sh := RawSynthesizer()
	body, _ := c.buildContextBody(ContextOptions{
		Intent: "x", Sources: []PillarConfig{src}, Synthesizer: &syn, Shape: &sh,
	})
	pipeline := body["pipeline"].(map[string]interface{})
	if pipeline["synthesizer"] != "markdown" {
		t.Fatalf("synthesizer should win: %v", pipeline["synthesizer"])
	}
}

func TestContext_Unit_BodyMutualExclusion(t *testing.T) {
	c := &AsyncWebCrawler{}
	src := GoogleWebSource(nil)
	_, err := c.buildContextBody(ContextOptions{
		Intent: "x", GeneratorID: "gen_x", Sources: []PillarConfig{src},
	})
	if err == nil {
		t.Fatalf("expected mutual-exclusion error")
	}
}

func TestContext_Unit_BodyInlineRequiresSource(t *testing.T) {
	c := &AsyncWebCrawler{}
	syn := RawSynthesizer()
	_, err := c.buildContextBody(ContextOptions{Intent: "x", Synthesizer: &syn})
	if err == nil {
		t.Fatalf("expected at-least-one-source error")
	}
}

func TestContext_Unit_BodyEmptyIntent(t *testing.T) {
	c := &AsyncWebCrawler{}
	_, err := c.buildContextBody(ContextOptions{Intent: "   "})
	if err == nil {
		t.Fatalf("expected empty-intent error")
	}
}

// ─── Unit — ContextOutput sugar ─────────────────────────────────────────

func TestContext_Unit_OutputRaw(t *testing.T) {
	out := ContextOutputFromMap(map[string]interface{}{
		"type": "raw",
		"data": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"url": "https://x", "title": "T", "content": "C"},
			},
		},
	})
	if out.Shape != "raw" || len(out.Items) != 1 || out.Items[0].URL != "https://x" {
		t.Fatalf("unexpected: %+v", out)
	}
	if out.Markdown != "" || out.Files != nil || out.Data != nil {
		t.Fatalf("raw shape should leave sugar fields zero")
	}
}

func TestContext_Unit_OutputMarkdownSingle(t *testing.T) {
	out := ContextOutputFromMap(map[string]interface{}{
		"type": "markdown",
		"data": map[string]interface{}{
			"mode":     "single",
			"items":    []interface{}{map[string]interface{}{"url": "https://a"}},
			"markdown": "# heading\n\nbody",
		},
	})
	if out.Shape != "markdown" || out.Markdown != "# heading\n\nbody" {
		t.Fatalf("unexpected: %+v", out)
	}
	if out.Files != nil {
		t.Fatalf("single mode should leave Files nil")
	}
}

func TestContext_Unit_OutputMarkdownMulti(t *testing.T) {
	out := ContextOutputFromMap(map[string]interface{}{
		"type": "markdown",
		"data": map[string]interface{}{
			"mode": "multi",
			"items": []interface{}{
				map[string]interface{}{"url": "https://a"},
				map[string]interface{}{"url": "https://b"},
			},
			"files": []interface{}{
				map[string]interface{}{"filename": "a.md", "markdown": "# A"},
				map[string]interface{}{"filename": "b.md", "markdown": "# B"},
			},
		},
	})
	if out.Markdown != "" || len(out.Files) != 2 {
		t.Fatalf("unexpected: markdown=%q files=%+v", out.Markdown, out.Files)
	}
	if out.Files[0].Filename != "a.md" || out.Files[0].Markdown != "# A" {
		t.Fatalf("file mismatch: %+v", out.Files[0])
	}
}

func TestContext_Unit_OutputLLM(t *testing.T) {
	out := ContextOutputFromMap(map[string]interface{}{
		"type": "llm",
		"data": map[string]interface{}{
			"items":           []interface{}{map[string]interface{}{"url": "https://a"}},
			"data":            map[string]interface{}{"runtimes": []interface{}{map[string]interface{}{"name": "tokio"}}},
			"resolved_schema": map[string]interface{}{"type": "object"},
			"notes":           []interface{}{"resolved schema from output_example (walked)"},
		},
	})
	if out.Shape != "llm" {
		t.Fatalf("shape: %q", out.Shape)
	}
	dataMap, ok := out.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Data not a map: %T", out.Data)
	}
	if _, ok := dataMap["runtimes"]; !ok {
		t.Fatalf("Data missing 'runtimes': %+v", dataMap)
	}
	if out.ResolvedSchema["type"] != "object" {
		t.Fatalf("resolved_schema: %+v", out.ResolvedSchema)
	}
	if len(out.Notes) != 1 {
		t.Fatalf("notes: %+v", out.Notes)
	}
}

// ─── Unit — SSE event parsing ───────────────────────────────────────────

func TestContext_Unit_ParseStatus(t *testing.T) {
	ev, ok := ParseContextEvent("status", map[string]interface{}{
		"type": "status", "status": "planning", "phase": "planning",
		"version": float64(1),
	})
	if !ok || ev.Type != ContextEventStatus || ev.Status != "planning" {
		t.Fatalf("unexpected: %+v ok=%v", ev, ok)
	}
}

func TestContext_Unit_ParsePhaseInit(t *testing.T) {
	ev, ok := ParseContextEvent("phase_progress", map[string]interface{}{
		"type": "phase_progress", "kind": "init", "total": float64(3),
		"items": []interface{}{map[string]interface{}{"id": "a"}},
	})
	if !ok || ev.Type != ContextEventPhaseInit || ev.PhaseInitTotal != 3 {
		t.Fatalf("unexpected: %+v ok=%v", ev, ok)
	}
}

func TestContext_Unit_ParsePhaseItem(t *testing.T) {
	ev, ok := ParseContextEvent("phase_progress", map[string]interface{}{
		"type": "phase_progress", "kind": "item_update",
		"id": "abc", "status": "done", "ms": float64(1240),
	})
	if !ok || ev.ItemID != "abc" || ev.ItemMs != 1240 {
		t.Fatalf("unexpected: %+v ok=%v", ev, ok)
	}
}

func TestContext_Unit_ParseTerminal(t *testing.T) {
	ev, ok := ParseContextEvent("terminal", map[string]interface{}{
		"type": "terminal", "status": "completed", "urls_crawled": float64(9),
	})
	if !ok || ev.URLsCrawled != 9 {
		t.Fatalf("unexpected: %+v ok=%v", ev, ok)
	}
}

func TestContext_Unit_ParseUnknownReturnsFalse(t *testing.T) {
	if _, ok := ParseContextEvent("mystery", map[string]interface{}{"type": "mystery"}); ok {
		t.Fatalf("expected ok=false")
	}
}

// ─── Unit — ContextResult helpers ───────────────────────────────────────

func TestContext_Unit_ResultIsTerminal(t *testing.T) {
	for _, s := range []string{"completed", "completed_partial", "failed", "cancelled"} {
		r := ContextResultFromMap(map[string]interface{}{"run_id": "x", "status": s, "version": float64(1)}, nil)
		if !r.IsTerminal() {
			t.Fatalf("%q should be terminal", s)
		}
	}
	r := ContextResultFromMap(map[string]interface{}{"run_id": "x", "status": "queued"}, nil)
	if r.IsTerminal() {
		t.Fatalf("queued should not be terminal")
	}
}

func TestContext_Unit_ResultIsSuccess(t *testing.T) {
	for _, s := range []string{"completed", "completed_partial"} {
		r := ContextResultFromMap(map[string]interface{}{"run_id": "x", "status": s}, nil)
		if !r.IsSuccess() {
			t.Fatalf("%q should be success", s)
		}
	}
	for _, s := range []string{"failed", "cancelled", "queued"} {
		r := ContextResultFromMap(map[string]interface{}{"run_id": "x", "status": s}, nil)
		if r.IsSuccess() {
			t.Fatalf("%q should NOT be success", s)
		}
	}
}

// ─── Live ───────────────────────────────────────────────────────────────

func TestContext_Live_DefaultGeneratorOneShot(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)
	result, err := c.Context(ContextOptions{
		Intent: "brief overview of what LangChain is, with citations",
		Constraints: &ContextConstraints{
			MaxItems: 5, MaxPerSource: 3, MaxCrawlTimeS: 60,
		},
		Timeout: 3 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Context: %v", err)
	}
	if !result.IsTerminal() {
		t.Fatalf("not terminal: %s", result.Status)
	}
	output, err := result.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if output.Shape != "raw" && output.Shape != "markdown" && output.Shape != "llm" {
		t.Fatalf("unexpected shape: %q", output.Shape)
	}
}

func TestContext_Live_InlinePipelineRaw(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)
	src := GoogleWebSource(&GoogleWebSourceOptions{TopKPerBackend: 5})
	strat := AllItemsStrategy()
	syn := RawSynthesizer()
	rec := NoopReconciler()
	result, err := c.Context(ContextOptions{
		Intent:      "what is HTTP/2",
		Sources:     []PillarConfig{src},
		Strategy:    &strat,
		Synthesizer: &syn,
		Reconciler:  &rec,
		Constraints: &ContextConstraints{MaxItems: 3, MaxCrawlTimeS: 45},
		Timeout:     3 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Context: %v", err)
	}
	if !result.IsTerminal() {
		t.Fatalf("not terminal: %s", result.Status)
	}
	output, _ := result.Output()
	if output.Shape != "raw" {
		t.Fatalf("expected raw shape, got %q", output.Shape)
	}
}

func TestContext_Live_Streaming(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	events, err := c.ContextStream(ctx, ContextStreamOptions{
		Intent: "one-line answer: what is RAG",
		Constraints: &ContextConstraints{
			MaxItems: 2, MaxPerSource: 2, MaxCrawlTimeS: 30,
		},
	})
	if err != nil {
		t.Fatalf("ContextStream: %v", err)
	}

	sawTerminal := false
	for ev := range events {
		if ev.Type == ContextEventTerminal {
			sawTerminal = true
			if !ContextTerminalStatuses[ev.Status] {
				t.Fatalf("terminal with non-terminal status: %s", ev.Status)
			}
		}
	}
	if !sawTerminal {
		t.Fatalf("stream never emitted terminal event")
	}
}

func TestContext_Live_GetAndCancel(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)

	deadline := time.Now().Add(60 * time.Second)
	var result *ContextResult
	var err error
	for time.Now().Before(deadline) {
		result, err = c.Context(ContextOptions{
			Intent:      "ignored — will be cancelled",
			Constraints: &ContextConstraints{MaxItems: 2, MaxCrawlTimeS: 30},
			NoWait:      true,
		})
		if err == nil {
			break
		}
		var qe *QuotaExceededError
		if !errors.As(err, &qe) {
			t.Fatalf("Context: %v", err)
		}
		time.Sleep(5 * time.Second)
	}
	if result == nil {
		t.Fatalf("no slot in 60s")
	}
	state, err := c.GetContextRun(result.RunID)
	if err != nil {
		t.Fatalf("GetContextRun: %v", err)
	}
	if state.RunID != result.RunID {
		t.Fatalf("run_id mismatch")
	}
	if err := c.CancelContextRun(result.RunID); err != nil {
		t.Fatalf("CancelContextRun: %v", err)
	}
}

func TestContext_Live_Catalog(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)
	catalog, err := c.ContextCatalog()
	if err != nil {
		t.Fatalf("ContextCatalog: %v", err)
	}
	hasName := func(entries []CatalogEntry, name string) bool {
		for _, e := range entries {
			if e.Name == name {
				return true
			}
		}
		return false
	}
	if !hasName(catalog.Sources, "google_web") {
		t.Fatalf("sources missing google_web")
	}
	if !hasName(catalog.Synthesizers, "raw") || !hasName(catalog.Synthesizers, "markdown") ||
		!hasName(catalog.Synthesizers, "llm") {
		t.Fatalf("synthesizers missing one of raw/markdown/llm: %+v", catalog.Synthesizers)
	}
	// Back-compat alias
	if !hasName(catalog.Shapes, "raw") {
		t.Fatalf("Shapes alias should mirror Synthesizers")
	}
	if !hasName(catalog.Reconcilers, "noop") {
		t.Fatalf("reconcilers missing noop")
	}
}
