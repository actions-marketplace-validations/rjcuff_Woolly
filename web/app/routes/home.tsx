import type { Route } from './+types/home';
import { useEffect, useState } from 'react';
import { Benchmark, Faq, Shipping } from '@/components/home-sections';
import { Features, Integrations } from '@/components/home-features';
import { SetupTabs } from '@/components/setup-tabs';
import { SiteFooter, SiteHeader } from '@/components/site-chrome';
import { PrimaryButton, SecondaryButton } from '@/components/ui';
import { detectOS, macApp, windowsApp, type OS } from '@/lib/download';
import { getStars } from '@/lib/github';
import { changelog } from '@/lib/changelog';
import { getReleases } from '@/lib/releases';
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
  const [stars, releases] = await Promise.all([getStars(), getReleases(changelog)]);
  return { stars, releases };
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
      <PrimaryButton href="/docs/quickstart">Get started</PrimaryButton>
      <SecondaryButton href={download.href}>{download.label}</SecondaryButton>
    </div>
  );
}

export default function Home({ loaderData }: Route.ComponentProps) {
  const [os, setOS] = useState<OS>('mac');
  useEffect(() => {
    setOS(detectOS());
    // The hero entrance plays once a session; the root script reads this on the next load.
    try {
      sessionStorage.setItem('intro-seen', '1');
    } catch {
      // Storage can be blocked; the entrance just plays again.
    }
  }, []);
  const download = DOWNLOADS[os];

  return (
    <div className="relative isolate flex min-h-svh flex-col overflow-x-clip px-[max(20px,env(safe-area-inset-left))]">
      <SiteHeader stars={loaderData?.stars} />

      <main className="mx-auto flex w-full max-w-[1040px] flex-col items-center">
        {/* The first screen: the pitch and the setup panel, nothing else. It
            fills the screen below the header, and its blue wash ends with it. */}
        <div className="relative flex min-h-[calc(100svh-4rem)] w-full flex-col items-center justify-center pt-12 pb-16 sm:pt-16">
          <div
            aria-hidden="true"
            className="hero-wash pointer-events-none absolute inset-y-0 left-1/2 -z-10 w-screen -translate-x-1/2"
          />
          <section className="flex w-full flex-col items-center gap-6 pb-12 text-center">
            <h1 className="max-w-[820px] text-balance text-[36px] font-semibold leading-[1.08] tracking-[-0.03em] sm:text-[56px]">
              <span className="animate-rise block">Your AI agents</span>{' '}
              <span className="animate-rise block" style={{ animationDelay: '80ms' }}>
                <span className="text-sweep">share one brain</span>
              </span>
            </h1>
            <p
              className="animate-rise max-w-[620px] text-pretty text-[18px] leading-relaxed text-fd-muted-foreground sm:text-[21px]"
              style={{ animationDelay: '160ms' }}
            >
              Every AI coding tool starts from zero. Woolly gives Claude Code, Cursor, Codex and Ollama one memory, so
              you explain your project once. It lives on your laptop, in files you can read.
            </p>
            <div className="animate-rise mt-2" style={{ animationDelay: '240ms' }}>
              <Buttons download={download} />
            </div>
          </section>

          <div className="animate-rise w-full max-w-[820px]" style={{ animationDelay: '320ms' }}>
            <SetupTabs />
          </div>
        </div>

        <Integrations />

        <Features />

        <Benchmark />

        <Shipping releases={loaderData?.releases ?? []} />

        <Faq />
      </main>

      <SiteFooter />
    </div>
  );
}
