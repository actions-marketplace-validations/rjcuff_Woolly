import { Link } from 'react-router';
import { BlueskyIcon, GitHubIcon, LinkedInIcon, SiteThemeToggle, Wordmark } from '@/components/brand';
import { formatStars, STAR_THRESHOLD } from '@/lib/github';
import { blueskyUrl, githubUrl, linkedinUrl } from '@/lib/shared';

export function SiteHeader({ stars }: { stars?: number | null }) {
  const showStars = typeof stars === 'number' && stars >= STAR_THRESHOLD;
  const link = 'rounded-md px-3 py-1.5 text-[15px] font-medium transition-colors hover:bg-fd-accent';
  return (
    // Bleeds past the page gutter so the bottom line runs edge to edge.
    <header className="sticky top-0 z-40 -mx-[max(20px,env(safe-area-inset-left))] border-b border-fd-border bg-fd-background/80 px-[max(20px,env(safe-area-inset-left))] backdrop-blur-md">
      <nav aria-label="Main" className="mx-auto grid h-16 max-w-[1200px] grid-cols-[1fr_auto] items-center gap-4 md:grid-cols-[1fr_auto_1fr]">
        <Link to="/" aria-label="Woolly home" className="justify-self-start">
          <Wordmark className="text-[26px] sm:text-[30px]" />
        </Link>

        <div className="flex items-center gap-1 max-md:hidden">
          <Link to="/docs" className={link}>
            Docs
          </Link>
          <Link to="/download" className={link}>
            Download
          </Link>
          <Link to="/docs/project-memory" className={link}>
            For teams
          </Link>
          <Link to="/docs/benchmark" className={link}>
            Benchmark
          </Link>
        </div>

        <div className="flex items-center gap-2 justify-self-end">
          <Link to="/docs" className={`${link} md:hidden`}>
            Docs
          </Link>
          <Link to="/download" className={`${link} md:hidden max-[380px]:hidden`}>
            Download
          </Link>
          <a
            href={githubUrl}
            aria-label={showStars ? `GitHub, ${stars} stars` : 'GitHub'}
            className="flex h-9 items-center overflow-hidden rounded-md text-[14px] font-medium shadow-[0_0_0_1px_var(--color-fd-border)] transition-colors hover:bg-fd-accent"
          >
            <span className="flex items-center gap-1.5 px-2.5">
              <GitHubIcon className="size-4" />
              <span className="max-sm:hidden">{showStars ? 'Star' : 'GitHub'}</span>
            </span>
            {showStars ? (
              <span className="flex h-full items-center border-l border-fd-border px-2.5 tabular-nums">{formatStars(stars)}</span>
            ) : null}
          </a>
          <SiteThemeToggle />
          <Link
            to="/docs/quickstart"
            className="btn-brand flex h-9 items-center gap-1.5 rounded-md px-3.5 text-[14px] font-medium max-sm:hidden"
          >
            Get started <span aria-hidden="true">→</span>
          </Link>
        </div>
      </nav>
    </header>
  );
}

export function SiteFooter() {
  const year = new Date().getFullYear();
  return (
    <footer className="flex flex-wrap items-center justify-between gap-x-6 gap-y-2 py-5 text-xs text-fd-muted-foreground">
      <span>© {year} Woolly™</span>
      <nav aria-label="Footer" className="flex items-center gap-5">
        <Link to="/privacy" className="hover:text-fd-foreground">
          Privacy
        </Link>
        <Link to="/terms" className="hover:text-fd-foreground">
          Terms
        </Link>
        <a href={githubUrl} className="hover:text-fd-foreground">
          GitHub
        </a>
        <a href={blueskyUrl} aria-label="Bluesky" title="Woolly on Bluesky" className="hover:text-fd-foreground">
          <BlueskyIcon className="size-4" />
        </a>
        <a href={linkedinUrl} aria-label="LinkedIn" title="Woolly on LinkedIn" className="hover:text-fd-foreground">
          <LinkedInIcon className="size-[15px]" />
        </a>
      </nav>
    </footer>
  );
}
