import type { Rule } from './types';

// The same JSON rule files consumed by the Go CLI, bundled at build time.
// This is what makes the web analyzer and the CLI provably identical.
const modules = import.meta.glob('../rules/*/*.json', {
  eager: true,
  import: 'default',
});

export const rules: Rule[] = (Object.entries(modules) as [string, Rule][])
  .filter(([path]) => !path.endsWith('schema.json'))
  .map(([, rule]) => rule)
  .sort((a, b) => (a.id < b.id ? -1 : 1));
