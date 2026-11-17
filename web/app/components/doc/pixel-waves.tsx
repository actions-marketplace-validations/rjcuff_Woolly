import { useEffect, useRef } from 'react';

// Woolly red through to memory yellow.
const PALETTE = ['#de3329', '#f97316', '#f59e0b', '#facc15'] as const;

const COLS = 72;
const ROWS = 26;

function hexToRgb(hex: string): [number, number, number] {
  const n = Number.parseInt(hex.slice(1), 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

/** A color along the palette, t from 0 (red) to 1 (yellow). */
function along(colors: [number, number, number][], t: number): [number, number, number] {
  const x = Math.min(0.9999, Math.max(0, t)) * (colors.length - 1);
  const i = Math.floor(x);
  const f = x - i;
  const a = colors[i];
  const b = colors[i + 1];
  return [a[0] + (b[0] - a[0]) * f, a[1] + (b[1] - a[1]) * f, a[2] + (b[2] - a[2]) * f];
}

/**
 * A full rectangle of pixels with slow waves rolling through it, in Woolly's
 * warm colors. Drawn on a canvas, paused while off screen, and still under
 * reduced motion.
 */
export function PixelWaves() {
  const canvas = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const el = canvas.current;
    if (!el) return;
    const ctx = el.getContext('2d');
    if (!ctx) return;

    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const colors = PALETTE.map(hexToRgb);
    const seed = Array.from({ length: COLS * ROWS }, () => Math.random() * Math.PI * 2);
    let size = 0;
    let base: [number, number, number] = [214, 216, 220];

    const resize = () => {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      const width = el.clientWidth;
      size = width / COLS;
      el.width = Math.round(width * dpr);
      el.height = Math.round(size * ROWS * dpr);
      el.style.height = `${size * ROWS}px`;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      base = document.documentElement.classList.contains('dark') ? [30, 30, 32] : [214, 216, 220];
      // Resizing a canvas clears it, so draw again right away.
      if (drawn) draw(reduce ? 4000 : performance.now());
    };
    let drawn = false;

    const draw = (now: number) => {
      drawn = true;
      const t = now / 1000;
      ctx.clearRect(0, 0, el.width, el.height);
      const gap = Math.max(1, size * 0.14);
      for (let y = 0; y < ROWS; y++) {
        for (let x = 0; x < COLS; x++) {
          // Three slow waves at different angles and speeds, summed.
          const u = x / COLS;
          const v = y / ROWS;
          let h =
            Math.sin(u * 7.0 + t * 0.9 + Math.sin(v * 3 + t * 0.3) * 1.2) * 0.5 +
            Math.sin((u + v) * 9.0 - t * 1.3) * 0.3 +
            Math.sin(v * 11.0 - u * 4.0 + t * 0.6) * 0.2;
          // A little drift per square, so the field never repeats exactly.
          h += Math.sin(seed[y * COLS + x] + t * (0.8 + (seed[y * COLS + x] % 1) * 0.6)) * 0.12;
          // Rows near the top run warmer, like a horizon.
          const level = (h + 1) / 2;
          const lit = Math.max(0, level - 0.18) / 0.82;
          const [r, g, b] = along(colors, Math.min(1, v * 0.7 + level * 0.5));
          const k = 0.2 + Math.pow(lit, 1.5) * 0.8;
          ctx.fillStyle = `rgb(${Math.round(base[0] + (r - base[0]) * k)},${Math.round(base[1] + (g - base[1]) * k)},${Math.round(base[2] + (b - base[2]) * k)})`;
          ctx.fillRect(x * size + gap / 2, y * size + gap / 2, size - gap, size - gap);
        }
      }
    };

    resize();
    if (reduce) {
      draw(4000);
      return;
    }

    let frame = 0;
    let visible = true;
    const loop = (now: number) => {
      draw(now);
      frame = visible ? requestAnimationFrame(loop) : 0;
    };
    frame = requestAnimationFrame(loop);

    const io = new IntersectionObserver(([entry]) => {
      visible = entry.isIntersecting;
      if (visible && !frame) frame = requestAnimationFrame(loop);
    });
    io.observe(el);

    const ro = new ResizeObserver(resize);
    ro.observe(el);
    const themeWatch = new MutationObserver(resize);
    themeWatch.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });

    return () => {
      cancelAnimationFrame(frame);
      io.disconnect();
      ro.disconnect();
      themeWatch.disconnect();
    };
  }, []);

  return (
    <canvas
      ref={canvas}
      role="img"
      aria-label="A rectangle of small squares in red, orange and yellow, with slow waves rolling through it."
      className="block w-full"
    />
  );
}
