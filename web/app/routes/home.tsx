import type { Route } from './+types/home';
import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { Benchmark, Faq } from '@/components/home-sections';
import { Features, Integrations } from '@/components/home-features';
import { SetupTabs } from '@/components/setup-tabs';
import { SiteFooter, SiteHeader } from '@/components/site-chrome';
import { detectOS, macApp, windowsApp, type OS } from '@/lib/download';
import { getStars } from '@/lib/github';
import { seo } from '@/lib/seo';
import { appName, blueskyUrl, githubUrl, linkedinUrl, siteDescription, tagline } from '@/lib/shared';

export function meta(_: Route.MetaArgs) {
  return [
    ...seo({
      title: `${appName} | ${tagline}`,
      description: siteDescription,
      path: '/',
      image: '/og/docs/image.png',
    }),
    {
      'script:ld+json': {
        '@context': 'https://schema.org',
        '@type': 'SoftwareApplication',
        name: appName,
        description: siteDescription,
        applicationCategory: 'DeveloperApplication',
        operatingSystem: 'macOS, Linux, Windows',
        offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
        license: 'https://opensource.org/licenses/MIT',
        sameAs: [githubUrl, blueskyUrl, linkedinUrl],
      },
    },
  ];
}

export async function loader() {
  return { stars: await getStars() };
}

// Linux has no app yet, so it goes to the download page for the install command.
const DOWNLOADS: Record<OS, { href: string; label: string }> = {
  mac: { href: macApp, label: 'Download for Mac' },
  windows: { href: windowsApp, label: 'Download for Windows' },
  linux: { href: '/download', label: 'Download for Linux' },
};

function Buttons({ download }: { download: { href: string; label: string } }) {
  return (
    <div className="flex flex-wrap items-center justify-center gap-3">
      <Link
        to="/docs/quickstart"
        className="btn-brand flex items-center gap-2 rounded-md py-2.5 pr-2.5 pl-5 text-[16px] font-medium"
      >
        Get started
        <span aria-hidden="true" className="flex size-7 items-center justify-center rounded bg-black/20 text-[15px]">
          →
        </span>
      </Link>
      <a
        href={download.href}
        className="rounded-md bg-fd-background px-5 py-3 text-[16px] font-medium shadow-[0_0_0_1px_var(--color-fd-border)] transition-colors hover:bg-fd-accent"
      >
        {download.label}
      </a>
    </div>
  );
}

export default function Home({ loaderData }: Route.ComponentProps) {
  const [os, setOS] = useState<OS>('mac');
  useEffect(() => setOS(detectOS()), []);
  const download = DOWNLOADS[os];

  return (
    <div className="relative isolate flex min-h-dvh flex-col overflow-x-clip px-[max(20px,env(safe-area-inset-left))]">
      <SiteHeader stars={loaderData?.stars} />

      <main className="mx-auto flex w-full max-w-[1040px] flex-col items-center">
        {/* The first screen: the pitch and the setup panel, nothing else. It
            fills the screen below the header, and its blue wash ends with it. */}
        <div className="relative flex min-h-[calc(100dvh-4rem)] w-full flex-col items-center justify-center pt-12 pb-16 sm:pt-16">
          <div
            aria-hidden="true"
            className="hero-wash pointer-events-none absolute inset-y-0 left-1/2 -z-10 w-screen -translate-x-1/2"
          />
          <section className="flex w-full flex-col items-center gap-6 pb-12 text-center">
            <p className="flex items-center gap-2 rounded-md bg-fd-background px-3 py-1.5 text-[14px] font-medium shadow-[0_0_0_1px_var(--color-fd-border)]">
              <span className="size-2 rounded-full bg-brand" aria-hidden="true" />
              Free and open source
            </p>
            <h1 className="max-w-[820px] text-balance text-[36px] font-semibold leading-[1.08] tracking-[-0.03em] sm:text-[56px]">
              One memory for every AI tool you use
            </h1>
            <p className="max-w-[620px] text-pretty text-[18px] leading-relaxed text-fd-muted-foreground sm:text-[21px]">
              Claude Code, Cursor and Ollama share what they learn about you and your project. It lives in plain files
              on your computer, and your team shares project memory through git.
            </p>
            <div className="mt-2">
              <Buttons download={download} />
            </div>
          </section>

          <div className="w-full max-w-[820px]">
            <SetupTabs />
          </div>
        </div>

        <Integrations />

        <Features />

        <Benchmark />

        <Faq />
      </main>

      <SiteFooter />
    </div>
  );
}
