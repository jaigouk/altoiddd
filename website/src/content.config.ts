import { defineCollection } from 'astro:content';
import { docsSchema } from '@astrojs/starlight/schema';
import { glob } from 'astro/loaders';

const docs = defineCollection({
  loader: glob({
    pattern: '**/*.md',
    base: '../docs/guides',
    generateId: ({ entry }) => `guides/${entry.replace(/\.md$/, '')}`,
  }),
  schema: docsSchema(),
});

export const collections = { docs };
