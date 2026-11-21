import type { Route } from './+types/home';
import { getStars } from '@/lib/github';
import { seo } from '@/lib/seo';
import { SkyFoot, SkyNav } from '@/components/sky/chrome';
import { Closing } from '@/components/sky/closing';
import { SkyHero } from '@/components/sky/hero';
import { Button, CopyInstall } from '@/components/sky/parts';
import { Benchmarks } from '@/components/sky/benchmarks';
import { BuiltFor } from '@/components/sky/built-for';
import { Faq } from '@/components/sky/faq';
import { WorksWith } from '@/components/sky/works-with';
import { appName, githubUrl, linkedinUrl, siteDescription, tagline } from '@/lib/shared';

export function meta(_: Route.MetaArgs) {
  return [
    ...seo({
      title: `${appName} | ${tagline}`,
      description: siteDescription,
      path: '/',
      image: '/og/home.png',
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
        sameAs: [githubUrl, linkedinUrl],
      },
    },
  ];
}

export async function loader() {
  return { stars: await getStars() };
}

// The homepage, top to bottom.
export default function Home({ loaderData }: Route.ComponentProps) {
  return (
    <div className="home-sky relative min-h-svh bg-white">
      {/* The nav floats over the top of the hero sky. */}
      <SkyNav stars={loaderData.stars} className="absolute inset-x-0 top-0" />
      <SkyHero
        title="Memory that knows when it’s wrong"
        lead="Woolly checks your agent’s notes against your code and flags every line a commit made inaccurate."
        actions={
          <>
            <Button href="/docs">Get started</Button>
            <CopyInstall className="max-sm:hidden" />
          </>
        }
      />
      <WorksWith />
      <BuiltFor />
      <Benchmarks />
      <Faq />
      <Closing />
      <SkyFoot />
    </div>
  );
}
