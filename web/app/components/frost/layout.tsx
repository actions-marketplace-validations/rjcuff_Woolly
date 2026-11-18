import type { ReactNode } from 'react';
import { WRAP } from '@/components/frost/nav';
import { cn } from '@/lib/utils';

/** A labeled box that stands in for a section until it has content. */
export function Placeholder({ label, note, className }: { label: string; note?: string; className?: string }) {
  return (
    <div className={cn('flex flex-col items-start justify-between gap-2 border border-dashed border-[var(--line-strong)] bg-[var(--ice-1)] p-5', className)}>
      <span className="font-mono text-[11px] uppercase tracking-[0.14em] text-[var(--muted)]">{label}</span>
      {note ? <span className="font-mono text-[11px] text-[var(--faint)]">{note}</span> : null}
    </div>
  );
}

/**
 * The top of the homepage. On large screens it splits in two: the left half
 * stays pinned under the nav while the right half scrolls past it. When the
 * right half runs out, the left half lets go and both scroll away together.
 * On smaller screens the halves stack.
 */
export function Split({ left, right }: { left: ReactNode; right: ReactNode }) {
  return (
    <section aria-label="Intro" className="border-b border-[var(--blue)]/20 lg:grid lg:grid-cols-2">
      <div aria-label="Hero" role="region" className="flex flex-col border-[var(--blue)]/20 max-lg:border-b lg:sticky lg:top-14 lg:h-[calc(100svh-3.5rem)] lg:self-start lg:border-r">
        {left}
      </div>
      <div aria-label="Scrolling panels" role="region" className="flex min-w-0 flex-col">
        {right}
      </div>
    </section>
  );
}

/** A normal full-width section below the split. */
export function Band({ label, children, className }: { label: string; children: ReactNode; className?: string }) {
  return (
    <section aria-label={label} className={cn('border-b border-[var(--blue)]/20 py-16 sm:py-24', className)}>
      <div className={WRAP}>{children}</div>
    </section>
  );
}
