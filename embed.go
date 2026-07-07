// Package analyzer embeds the shared diagnosis rules so the CLI ships as a
// single self-contained binary. The same JSON files under rules/ are consumed
// by the browser parser in web/.
package analyzer

import "embed"

//go:embed rules/freeradius/*.json rules/nps/*.json
var RulesFS embed.FS
