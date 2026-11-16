import type { Route } from './+types/home';
import { useEffect, useState } from 'react';
import { AgentView } from '@/components/doc/agent-view';
import { HomeArticle } from '@/components/doc/home-article';
import type { Mode } from '@/components/doc/rail';
import { DocShell } from '@/components/doc/shell';
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

export default function Home({ loaderData }: Route.ComponentProps) {
  const [os, setOS] = useState<OS>('mac');
  const [mode, setMode] = useState<Mode>('human');
  useEffect(() => {
    setOS(detectOS());
  }, []);

  const releases = loaderData?.releases ?? [];
  const switchMode = (next: Mode) => {
    setMode(next);
    window.scrollTo({ top: 0 });
  };

  return (
    <DocShell mode={mode} onMode={switchMode} stars={loaderData?.stars}>
      {mode === 'human' ? <HomeArticle download={DOWNLOADS[os]} releases={releases} /> : <AgentView releases={releases} />}
    </DocShell>
  );
}
