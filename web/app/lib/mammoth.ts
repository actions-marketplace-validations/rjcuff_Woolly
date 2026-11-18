/**
 * Woolly's mascot: a pixel mammoth facing right, worked out from a few shapes
 * on a grid, as runs of same-colored cells. Plain TypeScript with no imports,
 * so the site, the favicon script and the link preview images share it.
 */

export const MAMMOTH_W = 42;
export const MAMMOTH_H = 32;

export const MAMMOTH_COLORS = {
  outline: '#1a0f0a',
  fur: '#b8642e',
  dark: '#8c451d',
  light: '#d98b4f',
  ear: '#9a4c22',
  tusk: '#f6ecd8',
  eye: '#120c09',
  shine: '#ffffff',
  cheek: '#ff8fa3',
} as const;

export type Color = keyof typeof MAMMOTH_COLORS;

const ellipse = (x: number, y: number, cx: number, cy: number, rx: number, ry: number) =>
  ((x - cx) / rx) ** 2 + ((y - cy) / ry) ** 2 <= 1;

/** Points along a cubic curve, for the trunk and tusk. */
function curve(p: [number, number][], steps = 60) {
  const out: [number, number, number][] = [];
  for (let i = 0; i <= steps; i++) {
    const t = i / steps;
    const m = 1 - t;
    const [a, b, c, d] = p;
    out.push([
      m ** 3 * a[0] + 3 * m ** 2 * t * b[0] + 3 * m * t ** 2 * c[0] + t ** 3 * d[0],
      m ** 3 * a[1] + 3 * m ** 2 * t * b[1] + 3 * m * t ** 2 * c[1] + t ** 3 * d[1],
      t,
    ]);
  }
  return out;
}

const TRUNK = curve([
  [33, 14],
  [37.5, 19],
  [37, 27],
  [33, 28],
]);
const TUSK = curve([
  [30.5, 18.5],
  [31, 23.5],
  [33.5, 24.5],
  [36.5, 22],
]);

const near = (pts: [number, number, number][], x: number, y: number, r: (t: number) => number) =>
  pts.some(([px, py, t]) => (px - x) ** 2 + (py - y) ** 2 <= r(t) ** 2);

function cellColor(x: number, y: number): Color | null {
  const cx = x + 0.5;
  const cy = y + 0.5;
  if ((cx > 30 && cx < 32 && cy > 9 && cy < 11) || (cx > 30 && cx < 31 && cy > 11 && cy < 12)) return cx < 31 && cy < 10 ? 'shine' : 'eye';
  if (cy > 11 && cy < 12 && cx > 31 && cx < 32) return 'eye';
  if (cy > 13 && cy < 14 && cx > 32 && cx < 34) return 'cheek';
  if (near(TUSK, cx, cy, (t) => 1.25 - t * 0.55)) return 'tusk';
  const head = ellipse(cx, cy, 28.5, 12, 8, 8);
  const body = ellipse(cx, cy, 17, 18, 13, 9.5);
  const tuft = ellipse(cx, cy, 27, 3.8, 3.2, 1.9);
  const trunk = near(TRUNK, cx, cy, (t) => 1.9 - t * 0.6);
  const leg = cy > 23 && cy < 30 && [6, 12, 19, 24].some((l) => cx > l && cx < l + 4);
  const tail = cx > 3 && cx < 5 && cy > 15 && cy < 19.5;
  if (!(head || body || tuft || trunk || leg || tail)) return null;
  if (tuft || tail) return 'dark';
  if (ellipse(cx, cy, 24.5, 14, 2.4, 3.4) && !trunk) return 'ear';
  if (leg && !body) return cy > 29 ? 'dark' : 'fur';
  if (body && !head && cy > 24) return 'dark';
  if (ellipse(cx, cy, 11, 11.5, 5, 1.6) || ellipse(cx, cy, 26.5, 6.5, 3, 1.4)) return 'light';
  return 'fur';
}

export type Run = { x: number; y: number; w: number; c: Color };

/** The mammoth as runs of same-colored cells, with a dark outline around it. */
export const MAMMOTH_RUNS: Run[] = (() => {
  const grid: (Color | null)[][] = Array.from({ length: MAMMOTH_H }, (_, y) => Array.from({ length: MAMMOTH_W }, (_, x) => cellColor(x, y)));
  const filled = (x: number, y: number) => x >= 0 && y >= 0 && x < MAMMOTH_W && y < MAMMOTH_H && grid[y][x] !== null && grid[y][x] !== 'outline';
  const withOutline = grid.map((row, y) =>
    row.map((c, x) => c ?? (filled(x - 1, y) || filled(x + 1, y) || filled(x, y - 1) || filled(x, y + 1) ? 'outline' : null)),
  );
  const runs: Run[] = [];
  withOutline.forEach((row, y) => {
    let x = 0;
    while (x < MAMMOTH_W) {
      const c = row[x];
      if (!c) {
        x++;
        continue;
      }
      let w = 1;
      while (x + w < MAMMOTH_W && row[x + w] === c) w++;
      runs.push({ x, y, w, c });
      x += w;
    }
  });
  return runs;
})();

/** The parts the small one-color mammoth inks in. The lightest parts stay as paper. */
const SILHOUETTE: ReadonlySet<Color> = new Set(['outline', 'eye', 'dark', 'ear', 'fur']);

/** The flat one-color mammoth from the nav, as a square SVG for the favicon. */
export function silhouetteSvg(color: string): string {
  const pad = (MAMMOTH_W - MAMMOTH_H) / 2;
  const rects = MAMMOTH_RUNS.filter((r) => SILHOUETTE.has(r.c))
    .map((r) => `<rect x="${r.x}" y="${r.y + pad}" width="${r.w}" height="1"/>`)
    .join('');
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${MAMMOTH_W} ${MAMMOTH_W}" shape-rendering="crispEdges" fill="${color}">${rects}</svg>`;
}

/** How much ink each part gets in the dithered mammoth, from none to solid. */
export const DITHER_INK: Record<Color, 0 | 1 | 2 | 3 | 4> = {
  outline: 4,
  eye: 4,
  dark: 3,
  ear: 3,
  fur: 2,
  light: 1,
  cheek: 1,
  tusk: 0,
  shine: 0,
};

/**
 * The dithered mammoth from the homepage's closing section as a standalone
 * SVG, in one ink color on a transparent background. Shades are a 2 by 2
 * ordered dither drawn as plain squares, so any SVG renderer draws it the same.
 */
export function ditherSvg(ink: string): string {
  const dots: Record<1 | 2 | 3, [number, number][]> = {
    1: [[0, 0]],
    2: [
      [0, 0],
      [0.25, 0.25],
    ],
    3: [
      [0, 0],
      [0.25, 0],
      [0, 0.25],
    ],
  };
  let out = '';
  for (const r of MAMMOTH_RUNS) {
    const level = DITHER_INK[r.c];
    if (level === 0) continue;
    if (level === 4) {
      out += `<rect x="${r.x}" y="${r.y}" width="${r.w}" height="1"/>`;
      continue;
    }
    for (let x = r.x; x < r.x + r.w; x += 0.5) {
      for (let y = r.y; y < r.y + 1; y += 0.5) {
        for (const [dx, dy] of dots[level]) out += `<rect x="${x + dx}" y="${y + dy}" width="0.25" height="0.25"/>`;
      }
    }
  }
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${MAMMOTH_W} ${MAMMOTH_H}" shape-rendering="crispEdges" fill="${ink}">${out}</svg>`;
}

/** Woolly's two colors. */
export const BRAND = { blue: '#1a3fd1', paper: '#eef4fb' } as const;
