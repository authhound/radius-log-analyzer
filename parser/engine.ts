import type { Diagnosis, Format, Rule, RuleMatch } from './types';
import { rules } from './rules';
import { detect, summarize } from './detect';

// Mirrors internal/diagnose/engine.go — matching semantics must stay
// identical; both implementations are pinned by the fixtures in testdata/.

const MAX_EVIDENCE_LINES = 3;

interface Compiled {
  rule: Rule;
  any: RegExp[];
  all: RegExp[];
}

// All patterns are matched case-insensitively per rules/schema.json.
const compiled: Compiled[] = rules.map((rule) => ({
  rule,
  any: (rule.match.any ?? []).map((p) => new RegExp(p, 'i')),
  all: (rule.match.all ?? []).map((p) => new RegExp(p, 'i')),
}));

function matchRule(c: Compiled, lines: string[]): string[] | null {
  const evidence: string[] = [];
  const seen = new Set<string>();
  const addEvidence = (line: string) => {
    const trimmed = line.trim();
    if (evidence.length < MAX_EVIDENCE_LINES && !seen.has(trimmed)) {
      seen.add(trimmed);
      evidence.push(trimmed);
    }
  };

  for (const re of c.all) {
    let hit = false;
    for (const line of lines) {
      if (re.test(line)) {
        if (!hit) addEvidence(line);
        hit = true;
      }
    }
    if (!hit) return null;
  }

  if (c.any.length > 0) {
    let hit = false;
    for (const line of lines) {
      for (const re of c.any) {
        if (re.test(line)) {
          addEvidence(line);
          hit = true;
          break;
        }
      }
    }
    if (!hit) return null;
  }

  return evidence;
}

// A rule matches when every "all" pattern hits at least one line and, if
// "any" patterns exist, at least one of them hits a line. Matches are
// ordered by priority (desc) then rule id (asc).
export function analyze(text: string, format: Format | 'auto' = 'auto'): Diagnosis {
  const fmt: Format = format === 'auto' ? detect(text) : format;
  const diagnosis: Diagnosis = {
    format: fmt,
    summary: summarize(text, fmt),
    matches: [],
  };
  if (fmt === 'unknown') return diagnosis;

  const lines = text.split('\n');
  const matches: RuleMatch[] = [];
  for (const c of compiled) {
    if (c.rule.format !== fmt) continue;
    const evidence = matchRule(c, lines);
    if (evidence !== null) {
      matches.push({
        rule_id: c.rule.id,
        failure_stage: c.rule.failure_stage,
        title: c.rule.title,
        plain_english: c.rule.plain_english,
        priority: c.rule.priority,
        causes: c.rule.causes,
        evidence,
        ...(c.rule.doc_slug ? { doc_slug: c.rule.doc_slug } : {}),
      });
    }
  }
  matches.sort((a, b) => {
    if (a.priority !== b.priority) return b.priority - a.priority;
    return a.rule_id < b.rule_id ? -1 : 1;
  });
  diagnosis.matches = matches;
  return diagnosis;
}
