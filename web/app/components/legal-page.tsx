import type { ReactNode } from 'react';
import { DocShell } from '@/components/doc/shell';

export function LegalPage({ title, updated, children }: { title: string; updated: string; children: ReactNode }) {
  return (
    <DocShell>
      <section className="pb-10">
        <h1 className="text-[24px] font-normal leading-[1.35] tracking-[-0.01em] sm:text-[28px]">{title}</h1>
        <p className="mt-3 text-[13px] text-fd-muted-foreground">Last updated {updated}</p>
        <div className="legal mt-14">{children}</div>
      </section>
    </DocShell>
  );
}
