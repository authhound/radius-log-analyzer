package analyzer_test

import (
	"strings"
	"testing"

	analyzer "github.com/authhound/radius-log-analyzer"
	"github.com/authhound/radius-log-analyzer/internal/diagnose"
)

// TestRulesWellFormed validates every shipped rule: parses, regexes compile
// (diagnose.New does both), ids/format/stage/causes present and consistent.
func TestRulesWellFormed(t *testing.T) {
	eng, err := diagnose.New(analyzer.RulesFS)
	if err != nil {
		t.Fatal(err)
	}
	stages := map[string]bool{
		"client_hello": true, "cert_validation": true, "inner_auth": true,
		"timeout": true, "shared_secret": true, "eap_fragmentation": true,
		"policy_match": true, "client_config": true, "user_account": true,
		"backend": true, "proxy": true, "success": true, "unsupported_input": true,
	}
	rules := eng.Rules()
	if len(rules) < 15 {
		t.Fatalf("expected at least 15 rules, got %d", len(rules))
	}
	for _, r := range rules {
		if !strings.HasPrefix(r.ID, r.Format+"/") {
			t.Errorf("%s: id must start with %q", r.ID, r.Format+"/")
		}
		if !stages[r.FailureStage] {
			t.Errorf("%s: unknown failure_stage %q", r.ID, r.FailureStage)
		}
		if r.Title == "" || r.PlainEnglish == "" {
			t.Errorf("%s: missing title or plain_english", r.ID)
		}
		if len(r.Causes) == 0 || len(r.Causes) > 3 {
			t.Errorf("%s: want 1-3 causes, got %d", r.ID, len(r.Causes))
		}
		for i, c := range r.Causes {
			if c.Cause == "" || c.NextCheck == "" {
				t.Errorf("%s: cause %d missing cause or next_check", r.ID, i)
			}
		}
		if len(r.Match.Any) == 0 && len(r.Match.All) == 0 {
			t.Errorf("%s: rule has no match patterns", r.ID)
		}
	}
}
