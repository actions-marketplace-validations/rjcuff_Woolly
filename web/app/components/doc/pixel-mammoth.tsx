import { useEffect, useRef } from 'react';

// The same four warm colors as the brain, from Woolly red through to memory yellow.
const REGIONS = ['#de3329', '#f97316', '#f59e0b', '#facc15'] as const;
const TUSK = 4;
const STALE: [number, number, number] = [222, 51, 41];

const COLS = 64;
const ROWS = 30;
// One check sweeps across every few seconds and flags a few memories as stale.
const CHECK_EVERY = 5200;
const SWEEP = 1500;
const STALE_FOR = 2600;

type Cell = { x: number; y: number; region: number; energy: number; staleAt: number };
type Ripple = { x: number; y: number; born: number; reach: number };

const near = (x: number, y: number, points: [number, number][], r: number) =>
  points.some(([px, py]) => Math.hypot(px - x, py - y) < r);

/** Points along a quadratic curve, in grid cells. */
function curve(a: [number, number], b: [number, number], c: [number, number]) {
  const out: [number, number][] = [];
  for (let t = 0; t <= 1; t += 0.02) {
    out.push([
      (1 - t) ** 2 * a[0] + 2 * (1 - t) * t * b[0] + t ** 2 * c[0],
      (1 - t) ** 2 * a[1] + 2 * (1 - t) * t * b[1] + t ** 2 * c[1],
    ]);
  }
  return out;
}

const TRUNK = curve([9, 11], [3, 21], [9, 27]);
const TUSKS = curve([13, 17], [5, 25], [1, 19]);
const TAIL = curve([57, 13], [59, 16], [59, 20]);

/** What part of the mammoth a cell is, as a region index, or -1 for empty. */
function partAt(x: number, y: number): number {
  const ellipse = (cx: number, cy: number, rx: number, ry: number) => ((x - cx) / rx) ** 2 + ((y - cy) / ry) ** 2 <= 1;
  if (Math.hypot(x - 11.5, y - 8.5) < 0.8) return -1; // the eye
  if (near(x, y, TUSKS, 0.75) && y > 16) return TUSK;
  const head = ellipse(14, 10, 7.5, 8.5);
  const hump = ellipse(21, 9.5, 7, 8);
  const body = ellipse(37, 14.5, 20.5, 9.5) && y < 21;
  const fringe = y >= 19 && y < 22 && x > 18 && x < 54 && (x + y) % 3 !== 0; // the long hair under the belly
  const legs = y >= 17 && y < 29 && ((x >= 19 && x < 24) || (x >= 25 && x < 29) || (x >= 43 && x < 48) || (x >= 49 && x < 53));
  const trunk = near(x, y, TRUNK, y > 24 ? 1.2 : 1.7);
  const tail = near(x, y, TAIL, 0.7);
  if (!(head || hump || body || fringe || legs || trunk || tail)) return -1;
  if (y >= 20) return 3;
  if (x < 17) return 0;
  if (x < 36) return 1;
  return 2;
}

function hexToRgb(hex: string): [number, number, number] {
  const n = Number.parseInt(hex.slice(1), 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

/**
 * A woolly mammoth drawn in pixels, where each pixel is a memory. Memories
 * light up in small waves as tools save and recall them. Every few seconds a
 * check runs, and a few memories whose code changed are flagged red
 * as stale before they settle. Drawn on a canvas, paused while off screen,
 * and still under reduced motion.
 */
export function PixelMammoth() {
  const canvas = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const el = canvas.current;
    if (!el) return;
    const ctx = el.getContext('2d');
    if (!ctx) return;

    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const cells: Cell[] = [];
    const grid: (Cell | undefined)[] = [];
    for (let y = 0; y < ROWS; y++) {
      for (let x = 0; x < COLS; x++) {
        const region = partAt(x + 0.5, y + 0.5);
        if (region < 0) continue;
        const cell = { x, y, region, energy: 0, staleAt: -Infinity };
        cells.push(cell);
        grid[y * COLS + x] = cell;
      }
    }
    const memories = cells.filter((c) => c.region !== TUSK);
    const ripples: Ripple[] = [];

    let size = 0;
    let idle: [number, number, number] = [236, 236, 238];
    let tusk: [number, number, number] = [214, 211, 209];
    let colors = REGIONS.map(hexToRgb);
    const resize = () => {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      const width = el.clientWidth;
      size = width / COLS;
      el.width = Math.round(width * dpr);
      el.height = Math.round(size * ROWS * dpr);
      el.style.height = `${size * ROWS}px`;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      // The resting colors follow the theme.
      const dark = document.documentElement.classList.contains('dark');
      idle = dark ? [34, 34, 36] : [236, 236, 238];
      tusk = dark ? [120, 113, 108] : [214, 211, 209];
      colors = REGIONS.map(hexToRgb);
    };

    const spawn = (x: number, y: number, now: number, reach = 3 + Math.random() * 5) => {
      ripples.push({ x, y, born: now, reach });
    };

    // A check: which memories it will flag, picked when it starts.
    let checkStart = -Infinity;
    let flagged: Cell[] = [];
    const startCheck = (now: number) => {
      checkStart = now;
      flagged = [];
      const n = 2 + Math.floor(Math.random() * 3);
      for (let i = 0; i < n; i++) flagged.push(memories[Math.floor(Math.random() * memories.length)]);
    };

    const draw = (now: number) => {
      ctx.clearRect(0, 0, el.width, el.height);
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
            if (!cell || cell.region === TUSK || r.x + dx < 0 || r.x + dx >= COLS) continue;
            if (Math.abs(Math.hypot(dx, dy) - radius) < 0.75) cell.energy = Math.max(cell.energy, strength);
          }
        }
      }

      // A check runs across from front to back, unseen, flagging memories as it reaches them.
      const sweep = (now - checkStart) / SWEEP;
      const sweepX = sweep >= 0 && sweep <= 1 ? sweep * (COLS + 4) - 2 : -99;
      for (const c of flagged) {
        if (c.staleAt < checkStart && sweepX >= c.x) c.staleAt = now;
      }

      const gap = Math.max(1, size * 0.16);
      for (const cell of cells) {
        let rgb: [number, number, number];
        if (cell.region === TUSK) {
          rgb = tusk;
        } else {
          const [r, g, b] = colors[cell.region];
          const t = 0.07 + cell.energy * 0.93;
          const mix = (a: number, c: number) => Math.round(a + (c - a) * t);
          rgb = [mix(idle[0], r), mix(idle[1], g), mix(idle[2], b)];
        }
        const x = cell.x * size + gap / 2;
        const y = cell.y * size + gap / 2;
        const age = now - cell.staleAt;
        if (age >= 0 && age < STALE_FOR) {
          // A stale memory: hollow and red, blinking, then it fades back.
          const on = age < STALE_FOR * 0.6 ? Math.floor(age / 220) % 2 === 0 : true;
          const fade = age < STALE_FOR * 0.6 ? 1 : 1 - (age - STALE_FOR * 0.6) / (STALE_FOR * 0.4);
          ctx.fillStyle = `rgb(${idle[0]},${idle[1]},${idle[2]})`;
          ctx.fillRect(x, y, size - gap, size - gap);
          if (on) {
            const line = Math.max(1, size * 0.18);
            ctx.strokeStyle = `rgba(${STALE[0]},${STALE[1]},${STALE[2]},${fade})`;
            ctx.lineWidth = line;
            ctx.strokeRect(x + line / 2, y + line / 2, size - gap - line, size - gap - line);
          }
        } else {
          ctx.fillStyle = `rgb(${rgb[0]},${rgb[1]},${rgb[2]})`;
          ctx.fillRect(x, y, size - gap, size - gap);
        }
        cell.energy *= 0.978;
      }
    };

    resize();

    if (reduce) {
      // One still frame: a few memories lit, and two flagged as stale.
      for (let i = 0; i < 40; i++) memories[Math.floor(Math.random() * memories.length)].energy = 0.4 + Math.random() * 0.6;
      memories[Math.floor(memories.length * 0.3)].staleAt = 0;
      memories[Math.floor(memories.length * 0.7)].staleAt = 0;
      draw(1);
      return;
    }

    let frame = 0;
    let visible = true;
    let last = 0;
    let lastCheck = performance.now() - CHECK_EVERY + 1600;
    const loop = (now: number) => {
      if (now - last > 320) {
        last = now;
        const c = memories[Math.floor(Math.random() * memories.length)];
        spawn(c.x, c.y, now);
      }
      if (now - lastCheck > CHECK_EVERY) {
        lastCheck = now;
        startCheck(now);
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
      const cell = grid[y * COLS + x];
      if (cell && cell.region !== TUSK) spawn(x, y, now, 4);
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
      aria-label="A woolly mammoth drawn in small squares, each one a memory. Squares light up as memories are saved and recalled, and every few seconds a few are flagged in red as stale."
      className="block w-full touch-pan-y"
    />
  );
}
