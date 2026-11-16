import type { ReactNode } from 'react';
import { DocFooter } from '@/components/doc/footer';
import { HeroNav, type Mode } from '@/components/doc/rail';

/**
 * The frame for every page except the docs: one centered column, the name and
 * links floating at the left (or on top when there's no room), and the footer.
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
    <div className="home-doc flex min-h-svh flex-col bg-fd-background px-[max(20px,env(safe-area-inset-left))]">
      <div className="mx-auto flex w-full max-w-[720px] flex-1 flex-col">
        <HeroNav mode={mode} onMode={onMode} stars={stars} />
        <main className="flex-1">{children}</main>
        <DocFooter />
      </div>
    </div>
  );
}
