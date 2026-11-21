import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';
import { Button, CopyInstall, WRAP } from './parts';

// The closing picture's sky is one flat color, so the section uses it too.
const CTA_SKY = '#96b2f8';

/**
 * The last call to action: a headline on open sky with the ice world along
 * the bottom. The picture is pinned to the bottom center and only scales.
 */
export function Closing({
  title = 'Give your agents a memory they can trust',
  line = 'Free and open source. Runs on your machine.',
  actions,
}: {
  title?: string;
  line?: string;
  actions?: ReactNode;
}) {
  return (
    <section aria-label="Get started" className="relative overflow-hidden" style={{ backgroundColor: CTA_SKY }}>
      <div className={cn(WRAP, 'relative z-10 flex flex-col items-center pt-20 text-center sm:pt-28')}>
        <h2 className="serif max-w-[760px] text-balance text-[36px]/[1.08] tracking-[-0.025em] text-[var(--ink)] sm:text-[48px]/[1.05] lg:text-[56px]/[1.04]">
          {title}
        </h2>
        <p className="mt-4 text-[17px] text-[var(--body)]">{line}</p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-2">
          {actions ?? (
            <>
              <Button href="/docs">Get started</Button>
              <CopyInstall />
            </>
          )}
        </div>
      </div>
      <img
        src="/cta.webp"
        width={3168}
        height={1344}
        alt=""
        aria-hidden="true"
        loading="lazy"
        decoding="async"
        draggable={false}
        className="pointer-events-none relative -mt-14 block h-auto w-full min-w-[720px] select-none max-sm:left-1/2 max-sm:-translate-x-1/2 sm:-mt-[12%]"
      />
    </section>
  );
}
