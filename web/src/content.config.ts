import { defineCollection, z } from 'astro:content';
import { glob } from 'astro/loaders';

const kb = defineCollection({
  loader: glob({ pattern: '**/*.md', base: './src/content/kb' }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    format: z.enum(['freeradius', 'nps']),
    signature: z.string(),
    rule_id: z.string().optional(),
  }),
});

export const collections = { kb };
