import type { Route } from './+types/not-found';
import { DocShell } from '@/components/doc/shell';
import { DitherMammoth } from '@/components/frost/dither-mammoth';
import { Pill } from '@/components/frost/home-content';

export function meta({}: Route.MetaArgs) {
  return [{ title: 'Not Found | Woolly' }];
}

export default function NotFound() {
  return (
    <DocShell>
      <DitherMammoth className="w-[160px]" />
      <span className="mt-10 inline-block rounded-full border border-[var(--blue)] px-2.5 py-0.5 text-[12px] leading-[1.6]">404</span>
      <h1 className="mt-6 text-balance text-[44px] font-light leading-[1.02] tracking-[-0.035em] sm:text-[60px]">Even a mammoth forgets sometimes</h1>
      <p className="mt-6 max-w-[480px] text-[17px] leading-[1.6]">This page doesn’t exist, or it moved. The docs or the homepage will get you back on track.</p>
      <div className="mt-9 flex flex-wrap gap-3">
        <Pill href="/docs" solid>
          Read the docs
        </Pill>
        <Pill href="/">Go home</Pill>
      </div>
    </DocShell>
  );
}
