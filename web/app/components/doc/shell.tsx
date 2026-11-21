import type { ReactNode } from 'react';
import { SkyFoot, SkyNav } from '@/components/sky/chrome';
import { WRAP } from '@/components/sky/parts';
import { cn } from '@/lib/utils';

// The same flat sky as the homepage pictures, so the mammoth blends in.
const SKY = '#9bb3fb';

/**
 * The frame for the plain pages (download, privacy, terms, not found): a band
 * of sky with the nav, the page title and the balloon mammoth, then the page
 * in one readable column on white, then the footer.
 */
export function DocShell({
  eyebrow,
  title,
  lead,
  children,
  stars,
}: {
  eyebrow?: string;
  title: string;
  lead?: ReactNode;
  children: ReactNode;
  stars?: number | null;
}) {
  return (
    <div className="home-sky flex min-h-svh flex-col bg-white">
      <div className="relative overflow-hidden" style={{ backgroundColor: SKY }}>
        <img
          src="/mammoth-balloon.webp"
          width={920}
          height={880}
          alt=""
          aria-hidden="true"
          decoding="async"
          draggable={false}
          className="pointer-events-none absolute right-0 bottom-0 h-[72%] w-auto max-w-none select-none max-md:hidden"
        />
        <SkyNav stars={stars} className="relative" />
        <div className={cn(WRAP, 'relative pt-6 pb-14 sm:pt-10 sm:pb-20')}>
          <div className="max-w-[620px]">
            {eyebrow ? <span className="inline-flex rounded-full bg-white/60 px-3 py-1 text-[13px] font-semibold text-[var(--ink)] ring-1 ring-white/80">{eyebrow}</span> : null}
            <h1 className="serif mt-5 text-balance text-[42px]/[1.05] tracking-[-0.03em] text-[var(--ink)] sm:text-[56px]/[1.03]">{title}</h1>
            {lead ? <p className="mt-5 max-w-[520px] text-[17px] leading-[1.6] text-pretty text-[var(--body)] sm:text-[18px]">{lead}</p> : null}
          </div>
        </div>
      </div>
      <main className="mx-auto w-full max-w-[760px] flex-1 px-4 pt-12 pb-24 sm:px-8 sm:pt-16">{children}</main>
      <SkyFoot />
    </div>
  );
}
