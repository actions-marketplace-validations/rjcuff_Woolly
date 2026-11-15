import type { Route } from './+types/not-found';
import { Mammoth } from '@/components/brand';
import { SiteFooter, SiteHeader } from '@/components/site-chrome';
import { PrimaryButton, SecondaryButton } from '@/components/ui';

export function meta({}: Route.MetaArgs) {
  return [{ title: 'Not Found | Woolly' }];
}

export default function NotFound() {
  return (
    <div className="flex min-h-svh flex-col px-[max(20px,env(safe-area-inset-left))]">
      <SiteHeader />
      <main className="flex flex-1 flex-col items-center justify-center gap-5 py-16 text-center">
        <Mammoth className="size-12 text-brand" />
        <p className="font-mono text-[15px] font-semibold text-brand">404</p>
        <h1 className="text-[32px] font-semibold tracking-[-0.03em] sm:text-[44px]">
          Even a mammoth forgets sometimes
        </h1>
        <p className="max-w-[440px] text-pretty text-[17px] text-fd-muted-foreground">
          This page doesn&apos;t exist, or it moved. The docs or the homepage will get you back on track.
        </p>
        <div className="mt-2 flex flex-wrap items-center justify-center gap-3">
          <PrimaryButton href="/docs">Read the docs</PrimaryButton>
          <SecondaryButton href="/">Go home</SecondaryButton>
        </div>
      </main>
      <SiteFooter />
    </div>
  );
}
