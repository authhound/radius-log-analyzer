package analyzer_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	analyzer "github.com/authhound/radius-log-analyzer"
	"github.com/authhound/radius-log-analyzer/internal/diagnose"
)

type expected struct {
	Format  string   `json:"format"`
	Top     string   `json:"top"`
	Matched []string `json:"matched"`
}

// TestFixtures runs the engine over every testdata/<format>/<name>.log and
// compares against <name>.expected.json. The TypeScript engine runs the same
// corpus (web/test/parity.test.ts) — together they pin both implementations
// to identical behaviour.
func TestFixtures(t *testing.T) {
	eng, err := diagnose.New(analyzer.RulesFS)
	if err != nil {
		t.Fatal(err)
	}

	logs, err := filepath.Glob(filepath.Join("testdata", "*", "*.log"))
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) == 0 {
		t.Fatal("no fixtures found")
	}

	for _, logPath := range logs {
		logPath := logPath
		t.Run(logPath, func(t *testing.T) {
			text, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			expPath := strings.TrimSuffix(logPath, ".log") + ".expected.json"
			expData, err := os.ReadFile(expPath)
			if err != nil {
				t.Fatalf("fixture %s has no expected file: %v", logPath, err)
			}
			var exp expected
			if err := json.Unmarshal(expData, &exp); err != nil {
				t.Fatal(err)
			}

			d := eng.Analyze(string(text), diagnose.FormatAuto)
			if d.Format != exp.Format {
				t.Errorf("format = %q, want %q", d.Format, exp.Format)
			}
			var got []string
			for _, m := range d.Matches {
				got = append(got, m.RuleID)
			}
			if len(got) == 0 {
				t.Fatalf("no matches, want top %q", exp.Top)
			}
			if got[0] != exp.Top {
				t.Errorf("top match = %q, want %q", got[0], exp.Top)
			}
			if strings.Join(got, ",") != strings.Join(exp.Matched, ",") {
				t.Errorf("matched = %v, want %v", got, exp.Matched)
			}
		})
	}
}
