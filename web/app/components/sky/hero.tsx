import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';
import { WRAP } from './parts';

// The hero picture's sky is one flat color. The section uses that exact color,
// so wherever the picture doesn't reach, the sky simply carries on.
const HERO_SKY = '#9bb3fb';

/**
 * The sky at the top of a page: the ice world in the bottom right and the
 * words on the open sky. Phones get 80% of the screen with the scene along
 * the bottom; larger screens get 60% with the scene in the corner.
 */
interface SkyHeroProps {
  title: ReactNode;
  lead: ReactNode;
  actions: ReactNode;
  note?: ReactNode;
  /** Overrides the phone title size, for longer titles that should still fit on two lines. */
  titleClassName?: string;
}

export function SkyHero({ title, lead, actions, note, titleClassName }: SkyHeroProps) {
  return (
    <section aria-label="Hero" className="relative h-[80svh] min-h-[560px] w-full overflow-hidden md:h-[64vh] md:min-h-[540px]" style={{ backgroundColor: HERO_SKY }}>
      {/* The picture only ever scales, never stretches, and is pinned to the
          bottom right. On phones it's sized by width so its scene spans the
          screen. From tablets up it's as tall as the hero but never wider than
          62% of it, so it can't run under the words; if the cap kicks in, it
          shrinks and the sky fills in above it. */}
      <img
        src="/hero.webp"
        width={2752}
        height={1536}
        alt=""
        aria-hidden="true"
        fetchPriority="high"
        decoding="async"
        draggable={false}
        className="pointer-events-none absolute right-0 bottom-0 h-auto w-[155%] max-w-none select-none [mask-image:linear-gradient(to_bottom,transparent_0%,black_35%)] md:h-full md:w-auto md:max-w-[62%] md:object-contain md:object-right-bottom md:[mask-image:none]"
      />
      <div className={cn(WRAP, 'relative z-10 flex h-full flex-col items-start pt-36 sm:pt-40 md:justify-center md:pt-24')}>
        <h1 className={cn('serif max-w-[640px] text-balance text-[40px]/[1.06] tracking-[-0.03em] text-[var(--ink)] sm:text-[48px]/[1.05] lg:text-[54px]/[1.04]', titleClassName)}>{title}</h1>
        <p className="mt-4 max-w-[460px] text-[16px] leading-[1.55] text-pretty text-[var(--body)] sm:mt-5 sm:text-[18px]">{lead}</p>
        <div className="mt-6 flex flex-row items-center gap-3 sm:mt-8">{actions}</div>
        {note ? <p className="mt-3 text-[13.5px] text-[var(--body)]">{note}</p> : null}
      </div>
    </section>
  );
}
