import { useId } from 'react';
import { MAMMOTH_H, MAMMOTH_RUNS, MAMMOTH_W, type Color } from '@/lib/mammoth';
import { cn } from '@/lib/utils';

/**
 * The pixel mammoth in one color, shaded with ordered dither like a print
 * halftone. It draws in currentColor, so it works on paper and on blue.
 */

// How much ink each part of the mammoth gets, from none to solid.
const INK: Record<Color, 0 | 1 | 2 | 3 | 4> = {
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
 * At icon size the dots blur into gray, so small mammoths are drawn as a flat
 * silhouette instead: the darker parts in ink, the lightest left as paper.
 */
export function DitherMammoth({ className, label, small }: { className?: string; label?: string; small?: boolean }) {
  const id = useId().replace(/:/g, '');
  return (
    <svg
      viewBox={`0 0 ${MAMMOTH_W} ${MAMMOTH_H}`}
      shapeRendering="crispEdges"
      role={label ? 'img' : undefined}
      aria-label={label}
      aria-hidden={label ? undefined : true}
      className={cn('block fill-current', className)}
    >
      <defs>
        {/* A 2 by 2 ordered dither: one, two or three of every four dots inked. */}
        <pattern id={`${id}-1`} width="0.5" height="0.5" patternUnits="userSpaceOnUse">
          <rect width="0.25" height="0.25" />
        </pattern>
        <pattern id={`${id}-2`} width="0.5" height="0.5" patternUnits="userSpaceOnUse">
          <rect width="0.25" height="0.25" />
          <rect x="0.25" y="0.25" width="0.25" height="0.25" />
        </pattern>
        <pattern id={`${id}-3`} width="0.5" height="0.5" patternUnits="userSpaceOnUse">
          <rect width="0.5" height="0.25" />
          <rect y="0.25" width="0.25" height="0.25" />
        </pattern>
      </defs>
      {MAMMOTH_RUNS.map((r) => {
        const ink = INK[r.c];
        if (ink === 0 || (small && ink < 2)) return null;
        if (small) return <rect key={`${r.x}-${r.y}`} x={r.x} y={r.y} width={r.w} height={1} fill="currentColor" />;
        return <rect key={`${r.x}-${r.y}`} x={r.x} y={r.y} width={r.w} height={1} fill={ink === 4 ? 'currentColor' : `url(#${id}-${ink})`} />;
      })}
    </svg>
  );
}
