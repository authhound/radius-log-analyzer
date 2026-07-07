import { describe, expect, it } from 'vitest';
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { analyze } from '../src/lib/parser';

// Runs the TS engine over the same fixture corpus as the Go tests
// (parity_test.go) — together they pin both implementations to identical
// behaviour.
const testdataDir = join(__dirname, '..', '..', 'testdata');

interface Expected {
  format: string;
  top: string;
  matched: string[];
}

const fixtures: { name: string; log: string; expected: Expected }[] = [];
for (const dir of readdirSync(testdataDir)) {
  const dirPath = join(testdataDir, dir);
  for (const file of readdirSync(dirPath)) {
    if (!file.endsWith('.log')) continue;
    const base = file.slice(0, -'.log'.length);
    fixtures.push({
      name: `${dir}/${base}`,
      log: readFileSync(join(dirPath, file), 'utf8'),
      expected: JSON.parse(
        readFileSync(join(dirPath, `${base}.expected.json`), 'utf8'),
      ),
    });
  }
}

describe('fixture parity', () => {
  it('found fixtures', () => {
    expect(fixtures.length).toBeGreaterThan(0);
  });

  for (const f of fixtures) {
    it(f.name, () => {
      const d = analyze(f.log);
      expect(d.format).toBe(f.expected.format);
      const got = d.matches.map((m) => m.rule_id);
      expect(got[0]).toBe(f.expected.top);
      expect(got).toEqual(f.expected.matched);
    });
  }
});
