import type { ReactNode } from 'react';
import { Foot } from '@/components/frost/home-content';
import { FrostNav } from '@/components/frost/nav';

/**
 * The frame for the plain pages (download, privacy, terms, not found): the
 * homepage's nav and footer, with the page in one readable column.
 */
export function DocShell({ children, stars }: { children: ReactNode; stars?: number | null }) {
  return (
    <div className="home-frost flex min-h-svh flex-col bg-[var(--paper)] text-[var(--blue)]">
      <FrostNav stars={stars} />
      <main className="mx-auto w-full max-w-[760px] flex-1 px-5 pb-24 pt-14 sm:px-10 sm:pt-20">{children}</main>
      <Foot />
    </div>
  );
}
