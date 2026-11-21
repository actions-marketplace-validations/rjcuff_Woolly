// Renders Woolly's brand images from the pixel mammoth in app/lib/mammoth.ts:
// the site's icons in public/, and logos, profile pictures and banners for
// social accounts in ../assets/brand/2026/. Uses the Playwright that the site
// tests already install. Run with: node scripts/brand.ts
import { mkdirSync, writeFileSync } from 'node:fs';
import { chromium } from 'playwright';
import { BRAND, ditherSvg, MAMMOTH_H, MAMMOTH_W, silhouetteSvg } from '../app/lib/mammoth.ts';

const OUT = new URL('../../assets/brand/2026/', import.meta.url);
const PUBLIC = new URL('../public/', import.meta.url);
mkdirSync(OUT, { recursive: true });

const uri = (svg: string) => `data:image/svg+xml;base64,${Buffer.from(svg).toString('base64')}`;
const dither = { paper: uri(ditherSvg(BRAND.paper)), blue: uri(ditherSvg(BRAND.blue)) };
const flat = { paper: uri(silhouetteSvg(BRAND.paper)), blue: uri(silhouetteSvg(BRAND.blue)) };
const RATIO = MAMMOTH_H / MAMMOTH_W;
// The flat mammoth's SVG is square; the dithered one is the art's own shape.
const img = (src: string, w: number, square = false) => `<img src="${src}" width="${w}" height="${square ? w : Math.round(w * RATIO)}" style="image-rendering:pixelated;display:block">`;

const FONT = '<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter+Tight:wght@300;500&display=block">';
const page = (w: number, h: number, bg: string, body: string) =>
  `<!doctype html><html><head>${FONT}<style>*{margin:0;box-sizing:border-box}html,body{width:${w}px;height:${h}px;background:${bg};font-family:'Inter Tight',sans-serif;overflow:hidden}</style></head><body>${body}</body></html>`;
const center = (inner: string) => `<div style="display:flex;align-items:center;justify-content:center;width:100%;height:100%">${inner}</div>`;

/** A banner: the name and line on the left, the mammoth on the right. */
const banner = (w: number, h: number, scale: number) =>
  page(
    w,
    h,
    BRAND.blue,
    `<div style="display:flex;align-items:center;justify-content:space-between;width:100%;height:100%;padding:0 ${Math.round(w * 0.08)}px;color:${BRAND.paper}">
      <div>
        <div style="font-weight:500;font-size:${Math.round(h * 0.2 * scale)}px;letter-spacing:-0.03em;line-height:1">woolly</div>
        <div style="margin-top:${Math.round(h * 0.06)}px;font-weight:300;font-size:${Math.round(h * 0.085 * scale)}px;letter-spacing:-0.02em">Memory that knows when it’s wrong.</div>
      </div>
      ${img(dither.paper, Math.round(h * 0.78 * scale))}
    </div>`,
  );

/** The mammoth next to the name, for places that want a logo with words. */
const lockup = (ink: string, mark: string) =>
  page(
    1200,
    400,
    'transparent',
    `${center(`<div style="display:flex;align-items:center;gap:56px;color:${ink}">${img(mark, 300, true)}<span style="font-weight:500;font-size:190px;letter-spacing:-0.035em;line-height:1">woolly</span></div>`)}`,
  );

const jobs: { file: URL; w: number; h: number; html: string; transparent?: boolean }[] = [
  // The site's own icons. Small sizes use the flat mammoth, since dots blur at icon size.
  { file: new URL('apple-touch-icon.png', PUBLIC), w: 180, h: 180, html: page(180, 180, BRAND.blue, center(img(flat.paper, 132, true))) },
  { file: new URL('icon-192.png', PUBLIC), w: 192, h: 192, html: page(192, 192, BRAND.blue, center(img(flat.paper, 140, true))) },
  { file: new URL('icon-512.png', PUBLIC), w: 512, h: 512, html: page(512, 512, BRAND.blue, center(img(dither.paper, 340))) },

  // Profile pictures: the mammoth on blue, inside the circle most sites crop to.
  { file: new URL('profile-1024.png', OUT), w: 1024, h: 1024, html: page(1024, 1024, BRAND.blue, center(img(dither.paper, 640))) },
  { file: new URL('profile-400.png', OUT), w: 400, h: 400, html: page(400, 400, BRAND.blue, center(img(dither.paper, 260))) },
  { file: new URL('profile-flat-400.png', OUT), w: 400, h: 400, html: page(400, 400, BRAND.blue, center(img(flat.paper, 290, true))) },

  // Link previews and banners.
  { file: new URL('og-1200x630.png', OUT), w: 1200, h: 630, html: page(1200, 630, BRAND.blue, center(img(dither.paper, 546))) },
  { file: new URL('github-social-1280x640.png', OUT), w: 1280, h: 640, html: banner(1280, 640, 0.9) },
  { file: new URL('x-header-1500x500.png', OUT), w: 1500, h: 500, html: banner(1500, 500, 1) },
  { file: new URL('linkedin-banner-1584x396.png', OUT), w: 1584, h: 396, html: banner(1584, 396, 1) },

  // Logos on transparent backgrounds, for light and dark places.
  { file: new URL('mark-blue.png', OUT), w: 1024, h: 1024, transparent: true, html: page(1024, 1024, 'transparent', center(img(flat.blue, 1024, true))) },
  { file: new URL('mark-white.png', OUT), w: 1024, h: 1024, transparent: true, html: page(1024, 1024, 'transparent', center(img(flat.paper, 1024, true))) },
  { file: new URL('mark-dither-blue.png', OUT), w: 1024, h: 780, transparent: true, html: page(1024, 780, 'transparent', center(img(dither.blue, 1024))) },
  { file: new URL('mark-dither-white.png', OUT), w: 1024, h: 780, transparent: true, html: page(1024, 780, 'transparent', center(img(dither.paper, 1024))) },
  { file: new URL('logo-blue.png', OUT), w: 1200, h: 400, transparent: true, html: lockup(BRAND.blue, flat.blue) },
  { file: new URL('logo-white.png', OUT), w: 1200, h: 400, transparent: true, html: lockup(BRAND.paper, flat.paper) },
];

const browser = await chromium.launch();
for (const job of jobs) {
  const p = await browser.newPage({ viewport: { width: job.w, height: job.h } });
  await p.setContent(job.html, { waitUntil: 'networkidle' });
  await p.evaluate(() => document.fonts.ready);
  writeFileSync(job.file, await p.screenshot({ type: 'png', omitBackground: job.transparent ?? false }));
  await p.close();
  console.log('Wrote', job.file.pathname.replace(/.*\/(web|assets)\//, '$1/'));
}
await browser.close();

// The vector logos too, for anything that takes SVG.
writeFileSync(new URL('mark-blue.svg', OUT), `${silhouetteSvg(BRAND.blue)}\n`);
writeFileSync(new URL('mark-dither-blue.svg', OUT), `${ditherSvg(BRAND.blue)}\n`);
console.log('Wrote assets/brand/2026/mark-blue.svg and mark-dither-blue.svg');
