import type { ReactNode } from 'react';
import { Label } from '@/components/doc/parts';
import { DocShell } from '@/components/doc/shell';

export function LegalPage({ title, updated, children }: { title: string; updated: string; children: ReactNode }) {
  return (
    <DocShell>
      <section className="pb-20 pt-14">
        <Label>Last updated {updated}</Label>
        <h1 className="mt-4 text-[40px] font-normal leading-[1.06] tracking-[-0.035em] sm:text-[48px]">{title}</h1>
        <div className="legal mt-10">{children}</div>
      </section>
    </DocShell>
  );
}
