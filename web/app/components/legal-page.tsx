import type { ReactNode } from 'react';
import { DocShell } from '@/components/doc/shell';

export function LegalPage({ title, updated, children }: { title: string; updated: string; children: ReactNode }) {
  return (
    <DocShell eyebrow={`Last updated ${updated}`} title={title}>
      <div className="legal">{children}</div>
    </DocShell>
  );
}
