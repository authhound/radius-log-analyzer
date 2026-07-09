// @ts-check
import { defineConfig } from 'astro/config';

export default defineConfig({
  // Custom domain; attach it to the Pages project on first deploy.
  site: 'https://authhound.com',
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
