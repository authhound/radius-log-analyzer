// @ts-check
import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';

export default defineConfig({
  // Custom domain; attach it to the Pages project on first deploy.
  site: 'https://authhound.com',
  output: 'static',
  integrations: [sitemap()],
  vite: {
    server: {
      fs: {
        // rules/ and testdata/ live at the repo root, one level above web/.
        allow: ['..'],
      },
    },
  },
});
