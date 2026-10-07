import { mergeConfig } from 'vite';
import base from './vite.config';

// Standalone review artifact; production entry and routing stay independent.
export default mergeConfig(base, {
  build: {
    outDir: 'dist/design-prototype',
    rolldownOptions: { input: 'design.html' },
  },
});
