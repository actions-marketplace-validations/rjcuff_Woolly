import type { Route } from './+types/home';
import { Audiences, Closing, Foot, HeroContent, Numbers, Panels, TeamCallout } from '@/components/frost/home-content';
import { Band, Split } from '@/components/frost/layout';
import { FrostNav } from '@/components/frost/nav';
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

// The homepage: a split intro (the hero stays while the steps scroll past),
// then full-width sections.
export default function Home({ loaderData }: Route.ComponentProps) {
  return (
    <div className="home-frost flex min-h-svh flex-col bg-[var(--snow)] text-[var(--deep)]">
      <FrostNav stars={loaderData?.stars} />
      <main className="flex-1">
        <Split left={<HeroContent />} right={<Panels />} />
        <Band label="Measured on real repos">
          <Numbers />
        </Band>
        <Band label="Who it’s for">
          <Audiences />
        </Band>
        <Band label="Woolly Team">
          <TeamCallout />
        </Band>
        <Band label="Get started" className="bg-[var(--blue)]">
          <Closing />
        </Band>
      </main>
      <Foot />
    </div>
  );
}
