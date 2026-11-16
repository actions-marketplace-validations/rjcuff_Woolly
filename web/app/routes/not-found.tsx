import type { Route } from './+types/not-found';
import { BlockButton, FrameButton, Label } from '@/components/doc/parts';
import { DocShell } from '@/components/doc/shell';

export function meta({}: Route.MetaArgs) {
  return [{ title: 'Not Found | Woolly' }];
}

export default function NotFound() {
  return (
    <DocShell>
      <section className="pb-24 pt-14">
        <Label className="text-brand">404</Label>
        <h1 className="mt-4 text-[40px] font-normal leading-[1.06] tracking-[-0.035em] sm:text-[56px]">
          Even a mammoth forgets sometimes
        </h1>
        <p className="mt-6 max-w-[480px] text-pretty text-[18px] leading-relaxed text-fd-muted-foreground">
          This page doesn&apos;t exist, or it moved. The docs or the homepage will get you back on track.
        </p>
        <div className="mt-9 flex flex-wrap gap-3">
          <BlockButton href="/docs">Read the docs</BlockButton>
          <FrameButton href="/">Go home</FrameButton>
        </div>
      </section>
    </DocShell>
  );
}
