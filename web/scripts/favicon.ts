// Writes public/favicon.svg: the flat blue mammoth from the homepage nav,
// drawn from the pixel art in app/lib/mammoth.ts.
// Run with: node scripts/favicon.ts
import { writeFileSync } from 'node:fs';
import { BRAND, silhouetteSvg } from '../app/lib/mammoth.ts';

writeFileSync(new URL('../public/favicon.svg', import.meta.url), `${silhouetteSvg(BRAND.blue)}\n`);
console.log('Wrote public/favicon.svg');
