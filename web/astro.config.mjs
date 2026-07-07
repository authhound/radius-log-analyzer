// @ts-check
import { defineConfig } from 'astro/config';

export default defineConfig({
  // Update once a custom domain is attached in Cloudflare Pages.
  site: 'https://radius-log-analyzer.pages.dev',
  output: 'static',
  vite: {
    server: {
      fs: {
        // rules/ and testdata/ live at the repo root, one level above web/.
        allow: ['..'],
      },
    },
  },
});
