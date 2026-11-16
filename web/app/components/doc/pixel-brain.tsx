import { useEffect, useRef } from 'react';

// Four warm colors, one per region, from Woolly red through to memory yellow.
const REGIONS = ['#de3329', '#f97316', '#f59e0b', '#facc15'] as const;

const COLS = 64;
const ROWS = 30;

type Cell = { x: number; y: number; region: number; energy: number };
type Ripple = { x: number; y: number; born: number; reach: number };

/** Inside a brain seen from the side: the cerebrum, the cerebellum tucked under its back, and the stem. */
function inside(u: number, v: number) {
  const ellipse = (cx: number, cy: number, rx: number, ry: number) => ((u - cx) / rx) ** 2 + ((v - cy) / ry) ** 2 <= 1;
  const cerebrum = ellipse(0.47, 0.44, 0.4, 0.4) && v < 0.8 - Math.max(0, u - 0.62) * 0.9;
  const cerebellum = ellipse(0.7, 0.76, 0.15, 0.12);
  const stem = u > 0.52 && u < 0.6 && v > 0.66 && v < 0.98;
  return cerebrum || cerebellum || stem;
}

// The folds of the brain, as curves in the same 0 to 1 space. Cells on a fold are left out.
const FOLDS: [number, number][][] = [
  [[0.2, 0.62], [0.4, 0.5], [0.63, 0.47]], // the long fold between the temporal lobe and the rest
  [[0.52, 0.05], [0.47, 0.25], [0.42, 0.48]], // the central fold, front from back
  [[0.16, 0.22], [0.24, 0.3], [0.34, 0.4]],
  [[0.1, 0.42], [0.2, 0.44], [0.28, 0.52]],
  [[0.3, 0.1], [0.32, 0.2], [0.28, 0.3]],
  [[0.62, 0.1], [0.6, 0.22], [0.66, 0.38]],
  [[0.76, 0.2], [0.8, 0.32], [0.78, 0.48]],
  [[0.3, 0.7], [0.45, 0.64], [0.6, 0.62]],
  [[0.57, 0.72], [0.7, 0.7], [0.83, 0.73]],
  [[0.58, 0.8], [0.7, 0.79], [0.82, 0.8]],
];

/** Points along each fold, in grid cells, for a quick distance check. */
function foldPoints() {
  const points: [number, number][] = [];
  for (const [a, b, c] of FOLDS) {
    for (let t = 0; t <= 1; t += 0.02) {
      const u = (1 - t) ** 2 * a[0] + 2 * (1 - t) * t * b[0] + t ** 2 * c[0];
      const v = (1 - t) ** 2 * a[1] + 2 * (1 - t) * t * b[1] + t ** 2 * c[1];
      points.push([u * COLS, v * ROWS]);
    }
  }
  return points;
}

/** Which lobe a point is in, roughly: front, top, back, and the lower part with the stem. */
function region(u: number, v: number) {
  if (v > 0.62) return 3;
  if (u < 0.36) return 0;
  if (u < 0.6) return 1;
  return 2;
}

function hexToRgb(hex: string): [number, number, number] {
  const n = Number.parseInt(hex.slice(1), 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

/**
 * A brain drawn in pixels, where each pixel is a memory. Memories light up in
 * small waves, as if a tool just saved or recalled them, and the cursor sets
 * off waves of its own. Drawn on a canvas, paused while off screen, and still
 * under reduced motion.
 */
export function PixelBrain() {
  const canvas = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const el = canvas.current;
    if (!el) return;
    const ctx = el.getContext('2d');
    if (!ctx) return;

    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const cells: Cell[] = [];
    const grid: (Cell | undefined)[] = [];
    const folds = foldPoints();
    for (let y = 0; y < ROWS; y++) {
      for (let x = 0; x < COLS; x++) {
        const u = (x + 0.5) / COLS;
        const v = (y + 0.5) / ROWS;
        if (!inside(u, v)) continue;
        if (folds.some(([fx, fy]) => Math.hypot(fx - (x + 0.5), fy - (y + 0.5)) < 0.55)) continue;
        const cell = { x, y, region: region(u, v), energy: 0 };
        cells.push(cell);
        grid[y * COLS + x] = cell;
      }
    }
    const colors = REGIONS.map(hexToRgb);
    const ripples: Ripple[] = [];

    let size = 0;
    let idle: [number, number, number] = [236, 236, 236];
    const resize = () => {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      const width = el.clientWidth;
      size = width / COLS;
      el.width = Math.round(width * dpr);
      el.height = Math.round(size * ROWS * dpr);
      el.style.height = `${size * ROWS}px`;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      // The resting color follows the theme.
      const dark = document.documentElement.classList.contains('dark');
      idle = dark ? [34, 34, 36] : [236, 236, 238];
    };

    const spawn = (x: number, y: number, now: number, reach = 3 + Math.random() * 5) => {
      ripples.push({ x, y, born: now, reach });
    };

    const draw = (now: number) => {
      ctx.clearRect(0, 0, el.width, el.height);
      // Each ripple is a ring moving outward; cells it passes over light up.
      for (let i = ripples.length - 1; i >= 0; i--) {
        const r = ripples[i];
        const radius = (now - r.born) / 160;
        if (radius > r.reach) {
          ripples.splice(i, 1);
          continue;
        }
        const strength = 1 - radius / r.reach;
        const span = Math.ceil(radius) + 1;
        for (let dy = -span; dy <= span; dy++) {
          for (let dx = -span; dx <= span; dx++) {
            const cell = grid[(r.y + dy) * COLS + (r.x + dx)];
            if (!cell || r.x + dx < 0 || r.x + dx >= COLS) continue;
            if (Math.abs(Math.hypot(dx, dy) - radius) < 0.75) cell.energy = Math.max(cell.energy, strength);
          }
        }
      }
      const gap = Math.max(1, size * 0.16);
      for (const cell of cells) {
        const [r, g, b] = colors[cell.region];
        // At rest a cell carries a faint wash of its region's color.
        const t = 0.07 + cell.energy * 0.93;
        const mix = (a: number, c: number) => Math.round(a + (c - a) * t);
        ctx.fillStyle = `rgb(${mix(idle[0], r)},${mix(idle[1], g)},${mix(idle[2], b)})`;
        ctx.fillRect(cell.x * size + gap / 2, cell.y * size + gap / 2, size - gap, size - gap);
        cell.energy *= 0.978;
      }
    };

    resize();

    if (reduce) {
      // One still frame with a few memories lit, so it isn't blank.
      for (let i = 0; i < 40; i++) cells[Math.floor(Math.random() * cells.length)].energy = 0.4 + Math.random() * 0.6;
      draw(0);
      return;
    }

    let frame = 0;
    let visible = true;
    let last = 0;
    const loop = (now: number) => {
      if (now - last > 320) {
        last = now;
        const c = cells[Math.floor(Math.random() * cells.length)];
        spawn(c.x, c.y, now);
      }
      draw(now);
      frame = visible ? requestAnimationFrame(loop) : 0;
    };
    frame = requestAnimationFrame(loop);

    const io = new IntersectionObserver(([entry]) => {
      visible = entry.isIntersecting;
      if (visible && !frame) frame = requestAnimationFrame(loop);
    });
    io.observe(el);

    let lastPointer = 0;
    const onMove = (event: PointerEvent) => {
      const now = performance.now();
      if (now - lastPointer < 60) return;
      lastPointer = now;
      const rect = el.getBoundingClientRect();
      const x = Math.floor(((event.clientX - rect.left) / rect.width) * COLS);
      const y = Math.floor(((event.clientY - rect.top) / rect.height) * ROWS);
      if (grid[y * COLS + x]) spawn(x, y, now, 4);
    };
    el.addEventListener('pointermove', onMove);

    const ro = new ResizeObserver(resize);
    ro.observe(el);
    const themeWatch = new MutationObserver(resize);
    themeWatch.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

    return () => {
      cancelAnimationFrame(frame);
      io.disconnect();
      ro.disconnect();
      themeWatch.disconnect();
      el.removeEventListener('pointermove', onMove);
    };
  }, []);

  return (
    <canvas
      ref={canvas}
      role="img"
      aria-label="A brain drawn in small squares. Squares light up in red, orange, amber and yellow, like memories being saved and recalled."
      className="block w-full touch-pan-y"
    />
  );
}
