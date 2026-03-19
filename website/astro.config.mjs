import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  site: 'https://jaigouk.com',
  base: '/altoiddd/',
  integrations: [
    starlight({
      title: 'alto',
      customCss: ['./src/styles/global.css'],
      components: {
        ThemeProvider: './src/components/ThemeProvider.astro',
        ThemeSelect: './src/components/ThemeSelect.astro',
      },
      sidebar: [
        {
          label: 'Guides',
          items: [
            { label: 'Quickstart', slug: 'guides/quickstart' },
            { label: 'Installation', slug: 'guides/installation' },
            { label: 'New Project', slug: 'guides/new-project' },
            { label: 'Existing Project', slug: 'guides/existing-project' },
            { label: 'CLI Reference', slug: 'guides/cli-reference' },
            { label: 'AI Tool Integration', slug: 'guides/ai-tool-integration' },
            { label: 'Concepts', slug: 'guides/concepts' },
          ],
        },
      ],
    }),
  ],
  vite: {
    plugins: [tailwindcss()],
  },
});
