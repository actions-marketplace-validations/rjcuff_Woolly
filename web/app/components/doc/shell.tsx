import type { ReactNode } from 'react';
import { DocFooter } from '@/components/doc/footer';
import { HeroNav, type Mode } from '@/components/doc/rail';

/**
 * The frame for every page except the docs: one narrow centered column in
 * monospace, with the name and links on top.
 */
export function DocShell({
  children,
  stars,
  mode,
  onMode,
}: {
  children: ReactNode;
  stars?: number | null;
  mode?: Mode;
  onMode?: (mode: Mode) => void;
}) {
  return (
    <div className="home-doc flex min-h-svh flex-col bg-fd-background font-mono text-fd-foreground">
      <div className="mx-auto flex w-full max-w-[760px] flex-1 flex-col px-[max(20px,env(safe-area-inset-left))] sm:px-14">
        <HeroNav stars={stars} />
        <main className="flex-1">{children}</main>
        <DocFooter mode={mode} onMode={onMode} />
      </div>
    </div>
  );
}
