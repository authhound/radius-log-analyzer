// Package rule defines the shared diagnosis-rule format (rules/schema.json)
// and loads rule files from a filesystem.
package rule

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
)

type Cause struct {
	Cause     string `json:"cause"`
	NextCheck string `json:"next_check"`
}

type Matcher struct {
	Any []string `json:"any,omitempty"`
	All []string `json:"all,omitempty"`
}

type Rule struct {
	ID           string  `json:"id"`
	Format       string  `json:"format"`
	FailureStage string  `json:"failure_stage"`
	Title        string  `json:"title"`
	PlainEnglish string  `json:"plain_english"`
	Match        Matcher `json:"match"`
	Priority     int     `json:"priority"`
	Causes       []Cause `json:"causes"`
	DocSlug      string  `json:"doc_slug,omitempty"`
}

// LoadAll reads every rules/<format>/*.json file in fsys and returns the rules
// sorted by id for deterministic ordering.
func LoadAll(fsys fs.FS) ([]Rule, error) {
	var rules []Rule
	seen := map[string]bool{}
	err := fs.WalkDir(fsys, "rules", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path.Ext(p) != ".json" || path.Base(p) == "schema.json" {
			return nil
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		var r Rule
		if err := json.Unmarshal(data, &r); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		if r.ID == "" {
			return fmt.Errorf("%s: missing id", p)
		}
		if seen[r.ID] {
			return fmt.Errorf("%s: duplicate rule id %q", p, r.ID)
		}
		seen[r.ID] = true
		rules = append(rules, r)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })
	return rules, nil
}
