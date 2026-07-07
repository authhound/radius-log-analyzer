export type Format = 'freeradius' | 'nps' | 'unknown';

export interface Cause {
  cause: string;
  next_check: string;
}

export interface Matcher {
  any?: string[];
  all?: string[];
}

export interface Rule {
  id: string;
  format: 'freeradius' | 'nps';
  failure_stage: string;
  title: string;
  plain_english: string;
  match: Matcher;
  priority: number;
  causes: Cause[];
  doc_slug?: string;
}

export interface RuleMatch {
  rule_id: string;
  failure_stage: string;
  title: string;
  plain_english: string;
  priority: number;
  causes: Cause[];
  evidence: string[];
  doc_slug?: string;
}

export interface Diagnosis {
  format: string;
  summary: Record<string, string>;
  matches: RuleMatch[];
}
