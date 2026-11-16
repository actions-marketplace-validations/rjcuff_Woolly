import { BlueskyIcon, GitHubIcon, LinkedInIcon, Mammoth, SiteThemeToggle, Wordmark } from '@/components/brand';
import { SiteFooter as Footer } from '@/components/easeui/site-footer';
import { SiteNavbar } from '@/components/easeui/site-navbar';
import { formatStars, STAR_THRESHOLD } from '@/lib/github';
import { blueskyUrl, githubUrl, linkedinUrl } from '@/lib/shared';

// Bleeds past the page gutter so the bar and footer run edge to edge.
const BLEED = '-mx-[max(20px,env(safe-area-inset-left))] w-auto';

function GitHubButton({ stars }: { stars?: number | null }) {
  const showStars = typeof stars === 'number' && stars >= STAR_THRESHOLD;
  return (
    <a
      href={githubUrl}
      aria-label={showStars ? `GitHub, ${stars} stars` : 'GitHub'}
      className="btn-press flex h-9 items-center overflow-hidden rounded-md text-[14px] font-medium shadow-[0_0_0_1px_var(--color-fd-border)] transition-colors hover:bg-fd-accent"
    >
      <span className="flex items-center gap-1.5 px-2.5">
        <GitHubIcon className="size-4" />
        <span className="max-sm:hidden">{showStars ? 'Star' : 'GitHub'}</span>
      </span>
      {showStars ? (
        <span className="flex h-full items-center border-l border-fd-border px-2.5 tabular-nums">
          {formatStars(stars)}
        </span>
      ) : null}
    </a>
  );
}

export function SiteHeader({ stars }: { stars?: number | null }) {
  return (
    <SiteNavbar
      brand={{ label: <Wordmark className="text-[26px] sm:text-[30px]" />, href: '/' }}
      links={[
        { label: 'Docs', href: '/docs' },
        { label: 'Download', href: '/download' },
        { label: 'For teams', href: '/docs/project-memory' },
        { label: 'Benchmark', href: '/docs/benchmark' },
      ]}
      extras={
        <>
          <GitHubButton stars={stars} />
          <SiteThemeToggle />
        </>
      }
      primaryAction={{ label: 'Get started', href: '/docs/quickstart' }}
    />
  );
}

export function SiteFooter() {
  return (
    <Footer
      className={BLEED}
      brand={{
        label: (
          <span className="inline-flex items-center gap-2">
            <Mammoth className="size-7" />
            <Wordmark className="text-[26px]" />
          </span>
        ),
        href: '/',
      }}
      tagline="Your AI agents share one brain. Local, private and MIT licensed."
      columns={[
        {
          title: 'Product',
          links: [
            { label: 'Download', href: '/download' },
            { label: 'Project memory', href: '/docs/project-memory' },
            { label: 'Benchmark', href: '/docs/benchmark' },
            { label: 'Changelog', href: `${githubUrl}/blob/main/CHANGELOG.md` },
          ],
        },
        {
          title: 'Developers',
          links: [
            { label: 'Quickstart', href: '/docs/quickstart' },
            { label: 'Connect your apps', href: '/docs/connect' },
            { label: 'JavaScript', href: '/docs/javascript' },
            { label: 'Python', href: '/docs/python' },
            { label: 'Contributing', href: `${githubUrl}/blob/main/CONTRIBUTING.md` },
          ],
        },
        {
          title: 'Help',
          links: [
            { label: 'FAQ', href: '/docs/faq' },
            { label: 'Security', href: '/docs/security' },
            { label: 'Report an issue', href: `${githubUrl}/issues`, external: true },
          ],
        },
      ]}
      legal={[
        { label: 'Privacy', href: '/privacy' },
        { label: 'Terms', href: '/terms' },
      ]}
      social={[
        { label: 'GitHub', href: githubUrl, icon: <GitHubIcon className="size-4" /> },
        { label: 'Bluesky', href: blueskyUrl, icon: <BlueskyIcon className="size-4" /> },
        { label: 'LinkedIn', href: linkedinUrl, icon: <LinkedInIcon className="size-[15px]" /> },
      ]}
      copyright={`© ${new Date().getFullYear()} Woolly™`}
    />
  );
}
