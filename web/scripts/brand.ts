// Renders Woolly's brand images: the site's icons in public/ from the pixel
// mammoth in app/lib/mammoth.ts, and logos, profile pictures, banners and link
// previews for social accounts in ../assets/brand/2026/, from the pixel mammoth
// and the picture book art the website uses. Uses the Playwright that the site
// tests already install. Run with: node scripts/brand.ts
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
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

// The website's type: Fraunces for the name and headlines, Bricolage Grotesque for the rest.
const FONT = '<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,500&family=Bricolage+Grotesque:opsz,wght@12..96,400;500&display=block">';
const SERIF = "font-family:'Fraunces',Georgia,serif;font-weight:500;font-variation-settings:'SOFT' 100,'WONK' 0";
const INK = '#1b2350';
const page = (w: number, h: number, bg: string, body: string) =>
  `<!doctype html><html><head>${FONT}<style>*{margin:0;box-sizing:border-box}html,body{width:${w}px;height:${h}px;background:${bg};font-family:'Bricolage Grotesque',sans-serif;overflow:hidden;position:relative}</style></head><body>${body}</body></html>`;

// The picture book art from the homepage, and the sky color at the top of each.
const art = (file: string) => `data:image/webp;base64,${readFileSync(new URL(file, PUBLIC)).toString('base64')}`;
const SCENE = { hero: art('hero.webp'), herd: art('cta.webp'), balloon: art('mammoth-balloon.webp') };
const SKY = { hero: '#9bb4fb', herd: '#95b1f7', balloon: '#9bb3fb' };
const pic = (src: string, style: string) => `<img src="${src}" style="position:absolute;display:block;${style}">`;

/** The pixel mammoth and the name, as in the website's nav. */
const name = (size: number, ink = INK, sub = '') =>
  `<div style="display:flex;align-items:center;gap:${Math.round(size * 0.35)}px;color:${ink}">${img(ink === INK ? flat.blue : flat.paper, Math.round(size * 1.05), true)}<span style="${SERIF};font-size:${size}px;line-height:1;letter-spacing:-0.025em">woolly</span>${sub}</div>`;

/** A headline in the website's serif. */
const headline = (text: string, size: number, width: number) =>
  `<div style="${SERIF};font-size:${size}px;line-height:1.05;letter-spacing:-0.03em;color:${INK};max-width:${width}px;margin-top:${Math.round(size * 0.4)}px;text-wrap:balance">${text}</div>`;

const LINE = 'Memory that knows when it’s wrong';

/** A banner on the sky: the name and line on the left, the balloon mammoth scene on the right. */
const skyBanner = (w: number, h: number, o: { size: number; artW: number; artTop: number; left: number; top: number; width: number }) =>
  page(
    w,
    h,
    SKY.hero,
    pic(SCENE.hero, `right:0;top:${o.artTop}px;width:${o.artW}px`) +
      `<div style="position:absolute;left:${o.left}px;top:${o.top}px">${name(Math.round(o.size * 0.62))}${headline(LINE, o.size, o.width)}</div>`,
  );

/** A link preview: the name and line over the herd of mammoths. */
const herdPreview = (w: number, h: number) =>
  page(
    w,
    h,
    SKY.herd,
    pic(SCENE.herd, `left:50%;bottom:-6px;width:${Math.round(w * 1.04)}px;transform:translateX(-50%)`) +
      `<div style="position:absolute;left:${Math.round(w * 0.06)}px;top:${Math.round(h * 0.095)}px">${name(Math.round(h * 0.054))}${headline(LINE, Math.round(h * 0.095), Math.round(w * 0.6))}</div>`,
  );

/** A profile picture: the balloon mammoth, framed to sit inside a circle crop. */
const cartoonProfile = (size: number) => {
  const scale = size / 560;
  return page(size, size, SKY.balloon, pic(SCENE.balloon, `width:${Math.round(920 * scale)}px;left:${Math.round(-215 * scale)}px;top:${Math.round(-300 * scale)}px`));
};
const center = (inner: string) => `<div style="display:flex;align-items:center;justify-content:center;width:100%;height:100%">${inner}</div>`;

/** The mammoth next to the name, for places that want a logo with words. */
const lockup = (ink: string, mark: string) =>
  page(
    1200,
    400,
    'transparent',
    `${center(`<div style="display:flex;align-items:center;gap:56px;color:${ink}">${img(mark, 300, true)}<span style="${SERIF};font-size:190px;letter-spacing:-0.025em;line-height:1">woolly</span></div>`)}`,
  );

const jobs: { file: URL; w: number; h: number; html: string; transparent?: boolean }[] = [
  // The site's own icons. Small sizes use the flat mammoth, since dots blur at icon size.
  { file: new URL('apple-touch-icon.png', PUBLIC), w: 180, h: 180, html: page(180, 180, BRAND.blue, center(img(flat.paper, 132, true))) },
  { file: new URL('icon-192.png', PUBLIC), w: 192, h: 192, html: page(192, 192, BRAND.blue, center(img(flat.paper, 140, true))) },
  { file: new URL('icon-512.png', PUBLIC), w: 512, h: 512, html: page(512, 512, BRAND.blue, center(img(dither.paper, 340))) },

  // Profile pictures: the balloon mammoth on the sky, and the pixel mammoth on
  // night blue for places that want a plain mark.
  { file: new URL('profile-1024.png', OUT), w: 1024, h: 1024, html: cartoonProfile(1024) },
  { file: new URL('profile-400.png', OUT), w: 400, h: 400, html: cartoonProfile(400) },
  { file: new URL('profile-mark-1024.png', OUT), w: 1024, h: 1024, html: page(1024, 1024, BRAND.blue, center(img(dither.paper, 640))) },
  { file: new URL('profile-mark-400.png', OUT), w: 400, h: 400, html: page(400, 400, BRAND.blue, center(img(dither.paper, 260))) },
  { file: new URL('profile-flat-400.png', OUT), w: 400, h: 400, html: page(400, 400, BRAND.blue, center(img(flat.paper, 290, true))) },

  // Link previews and banners, on the homepage's sky. X and LinkedIn put the
  // profile picture over the bottom left, so the words sit high.
  { file: new URL('og-1200x630.png', OUT), w: 1200, h: 630, html: herdPreview(1200, 630) },
  { file: new URL('og/home.png', PUBLIC), w: 1200, h: 630, html: herdPreview(1200, 630) },
  { file: new URL('github-social-1280x640.png', OUT), w: 1280, h: 640, html: herdPreview(1280, 640) },
  { file: new URL('x-header-1500x500.png', OUT), w: 1500, h: 500, html: skyBanner(1500, 500, { size: 54, artW: 1060, artTop: -40, left: 90, top: 70, width: 560 }) },
  { file: new URL('linkedin-banner-1584x396.png', OUT), w: 1584, h: 396, html: skyBanner(1584, 396, { size: 44, artW: 980, artTop: -60, left: 420, top: 64, width: 560 }) },

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
