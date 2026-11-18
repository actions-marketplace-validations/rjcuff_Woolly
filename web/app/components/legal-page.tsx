import type { ReactNode } from 'react';
import { DocShell } from '@/components/doc/shell';

export function LegalPage({ title, updated, children }: { title: string; updated: string; children: ReactNode }) {
  return (
    <DocShell>
      <span className="inline-block rounded-full border border-[var(--blue)] px-2.5 py-0.5 text-[12px] leading-[1.6]">Last updated {updated}</span>
      <h1 className="mt-6 text-[44px] font-light leading-[1.02] tracking-[-0.035em] sm:text-[60px]">{title}</h1>
      <div className="legal mt-10">{children}</div>
    </DocShell>
  );
}
