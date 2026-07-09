// Package diagnose matches log text against the shared rule set and produces
// a ranked diagnosis. Matching semantics are mirrored by the TypeScript
// engine in web/src/lib/parser — any change here must be made there too and
// is guarded by the shared fixtures in testdata/.
package diagnose

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"github.com/authhound/radius-log-analyzer/internal/parse"
	"github.com/authhound/radius-log-analyzer/internal/rule"
)

const maxEvidenceLines = 3

type compiled struct {
	rule.Rule
	anyRe []*regexp.Regexp
	allRe []*regexp.Regexp
}

type Engine struct {
	rules []compiled
}

// New loads and compiles rules from fsys (usually analyzer.RulesFS).
// All patterns are matched case-insensitively per rules/schema.json.
func New(fsys fs.FS) (*Engine, error) {
	rules, err := rule.LoadAll(fsys)
	if err != nil {
		return nil, err
	}
	e := &Engine{}
	for _, r := range rules {
		c := compiled{Rule: r}
		for _, p := range r.Match.Any {
			re, err := regexp.Compile("(?i)" + p)
			if err != nil {
				return nil, fmt.Errorf("rule %s: bad pattern %q: %w", r.ID, p, err)
			}
			c.anyRe = append(c.anyRe, re)
		}
		for _, p := range r.Match.All {
			re, err := regexp.Compile("(?i)" + p)
			if err != nil {
				return nil, fmt.Errorf("rule %s: bad pattern %q: %w", r.ID, p, err)
			}
			c.allRe = append(c.allRe, re)
		}
		e.rules = append(e.rules, c)
	}
	return e, nil
}

func (e *Engine) Rules() []rule.Rule {
	out := make([]rule.Rule, len(e.rules))
	for i, c := range e.rules {
		out[i] = c.Rule
	}
	return out
}

type RuleMatch struct {
	RuleID       string       `json:"rule_id"`
	FailureStage string       `json:"failure_stage"`
	Title        string       `json:"title"`
	PlainEnglish string       `json:"plain_english"`
	Priority     int          `json:"priority"`
	Causes       []rule.Cause `json:"causes"`
	Evidence     []string     `json:"evidence"`
	DocSlug      string       `json:"doc_slug,omitempty"`
}

type Diagnosis struct {
	Format  string            `json:"format"`
	Summary map[string]string `json:"summary"`
	Matches []RuleMatch       `json:"matches"`
}

// Analyze runs every rule for the given format over the log text.
// A rule matches when every "all" pattern hits at least one line and, if
// "any" patterns exist, at least one of them hits a line. Matches are ordered
// by priority (desc) then rule id (asc).
func (e *Engine) Analyze(text string, format parse.Format) Diagnosis {
	if format == "" || format == FormatAuto {
		format = parse.Detect(text)
	}
	d := Diagnosis{
		Format:  string(format),
		Summary: parse.Summary(text, format),
		Matches: []RuleMatch{},
	}
	if format == parse.FormatUnknown {
		return d
	}
	lines := strings.Split(text, "\n")
	for _, c := range e.rules {
		if c.Format != string(format) {
			continue
		}
		if m, evidence := matchRule(c, lines); m {
			d.Matches = append(d.Matches, RuleMatch{
				RuleID:       c.ID,
				FailureStage: c.FailureStage,
				Title:        c.Title,
				PlainEnglish: c.PlainEnglish,
				Priority:     c.Priority,
				Causes:       c.Causes,
				Evidence:     evidence,
				DocSlug:      c.DocSlug,
			})
		}
	}
	sort.SliceStable(d.Matches, func(i, j int) bool {
		if d.Matches[i].Priority != d.Matches[j].Priority {
			return d.Matches[i].Priority > d.Matches[j].Priority
		}
		return d.Matches[i].RuleID < d.Matches[j].RuleID
	})
	return d
}

// FormatAuto asks Analyze to run detection itself.
const FormatAuto parse.Format = "auto"

func matchRule(c compiled, lines []string) (bool, []string) {
	var evidence []string
	seen := map[string]bool{}
	addEvidence := func(line string) {
		line = strings.TrimSpace(line)
		if len(evidence) < maxEvidenceLines && !seen[line] {
			seen[line] = true
			evidence = append(evidence, line)
		}
	}

	for _, re := range c.allRe {
		hit := false
		for _, line := range lines {
			if re.MatchString(line) {
				if !hit {
					addEvidence(line)
				}
				hit = true
			}
		}
		if !hit {
			return false, nil
		}
	}

	if len(c.anyRe) > 0 {
		hit := false
		for _, line := range lines {
			for _, re := range c.anyRe {
				if re.MatchString(line) {
					addEvidence(line)
					hit = true
					break
				}
			}
		}
		if !hit {
			return false, nil
		}
	}

	return true, evidence
}
