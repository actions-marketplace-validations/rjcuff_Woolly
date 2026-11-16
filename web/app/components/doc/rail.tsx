import { ArrowUpRight } from 'lucide-react';
import { Link } from 'react-router';
import { GitHubIcon, SiteThemeToggle, Wordmark } from '@/components/brand';
import { formatStars, STAR_THRESHOLD } from '@/lib/github';
import { githubUrl } from '@/lib/shared';
import { cn } from '@/lib/utils';

export type Mode = 'human' | 'agent';
/** A section of the page, with the memory file it reads as. */
export type DocSection = { id: string; label: string };

const LINKS: { label: string; href: string }[] = [
  { label: 'Docs', href: '/docs' },
  { label: 'Download', href: '/download' },
  { label: 'For teams', href: '/docs/project-memory' },
  { label: 'Benchmark', href: '/docs/benchmark' },
  { label: 'Changelog', href: `${githubUrl}/blob/main/CHANGELOG.md` },
];

function A({ href, className, children, onClick }: { href: string; className?: string; children: React.ReactNode; onClick?: () => void }) {
  return href.startsWith('/') ? (
    <Link to={href} className={className} onClick={onClick}>
      {children}
    </Link>
  ) : (
    <a href={href} className={className} onClick={onClick}>
      {children}
    </a>
  );
}

const MODES: { id: Mode; label: string }[] = [
  { id: 'human', label: 'Page' },
  { id: 'agent', label: 'Raw .md' },
];

/** Page reads as a page; Raw .md shows the same words as the Markdown files Woolly would keep. */
function ModeToggle({ mode, onMode }: { mode: Mode; onMode: (mode: Mode) => void }) {
  const index = MODES.findIndex((m) => m.id === mode);
  return (
    <div role="group" aria-label="Read as" className="relative grid grid-cols-2 rounded-lg bg-fd-muted p-1 shadow-[inset_0_0_0_1px_var(--color-fd-border)]">
      {/* One thumb that slides between the two, instead of two that swap. */}
      <span
        aria-hidden="true"
        className="absolute inset-y-1 left-1 w-[calc(50%-4px)] rounded-md bg-fd-background shadow-[0_0_0_1px_var(--color-fd-border),0_1px_3px_rgb(0_0_0/0.08)] transition-transform duration-300 ease-ui motion-reduce:transition-none"
        style={{ transform: `translateX(${index * 100}%)` }}
      />
      {MODES.map((m) => (
        <button
          key={m.id}
          type="button"
          aria-pressed={mode === m.id}
          onClick={() => onMode(m.id)}
          className={cn(
            'relative min-h-7 whitespace-nowrap rounded-md px-2.5 text-[13px] sm:px-3 font-medium outline-none transition-colors duration-200 focus-visible:ring-2 focus-visible:ring-brand',
            mode === m.id ? 'text-fd-foreground' : 'text-fd-muted-foreground hover:text-fd-foreground',
            m.id === 'agent' && 'font-mono text-[12px]',
          )}
        >
          {m.label}
        </button>
      ))}
    </div>
  );
}

function GitHubLink({ stars }: { stars?: number | null }) {
  const showStars = typeof stars === 'number' && stars >= STAR_THRESHOLD;
  return (
    <a
      href={githubUrl}
      aria-label={showStars ? `GitHub, ${stars} stars` : 'GitHub'}
      className="group inline-flex items-center gap-1.5 text-[15px] text-fd-muted-foreground transition-colors duration-150 hover:text-fd-foreground"
    >
      <GitHubIcon className="size-[15px] transition-transform duration-200 ease-ui group-hover:-rotate-12" />
      <span className="link-draw">GitHub</span>
      {showStars ? <span className="font-mono text-[12px] tabular-nums">{formatStars(stars)}</span> : null}
      <ArrowUpRight aria-hidden="true" className={ARROW} />
    </a>
  );
}

const ARROW =
  'size-3.5 opacity-50 transition-[transform,opacity] duration-200 ease-out group-hover:-translate-y-0.5 group-hover:translate-x-0.5 group-hover:opacity-100';

function Links({ vertical, stars }: { vertical?: boolean; stars?: number | null }) {
  return (
    <ul className={vertical ? 'flex flex-col gap-2.5' : 'flex flex-wrap items-center gap-x-5 gap-y-2'}>
      {LINKS.map((l) => (
        <li key={l.label}>
          <A href={l.href} className="group inline-flex items-center gap-1 text-[15px] text-fd-muted-foreground transition-colors duration-150 hover:text-fd-foreground">
            <span className="link-draw">{l.label}</span>
            <ArrowUpRight aria-hidden="true" className={ARROW} />
          </A>
        </li>
      ))}
      <li>
        <GitHubLink stars={stars} />
      </li>
    </ul>
  );
}

/**
 * The name, the links and the switches. On a wide screen they float at the
 * left, fixed to the window with no bar behind them, beside the content. When
 * there's no room beside it, they sit at the top of the column instead.
 */
export function HeroNav({ mode, onMode, stars }: { mode?: Mode; onMode?: (mode: Mode) => void; stars?: number | null }) {
  const toggle = mode && onMode ? <ModeToggle mode={mode} onMode={onMode} /> : null;
  return (
    <>
      <div className="fixed top-14 z-40 hidden w-[200px] xl:block" style={{ left: 'max(32px, calc((100vw - 720px) / 2 - 280px))' }}>
        <Link to="/" className="rounded-md outline-none focus-visible:ring-2 focus-visible:ring-brand">
          <Wordmark className="text-[30px]" />
        </Link>
        <nav aria-label="Main" className="mt-8">
          <Links vertical stars={stars} />
        </nav>
        <div className="mt-8 flex items-center gap-1.5">
          {toggle}
          <SiteThemeToggle />
        </div>
      </div>

      <div className="pt-10 sm:pt-14 xl:hidden">
        <div className="flex items-center justify-between gap-4">
          <Link to="/" className="rounded-md outline-none focus-visible:ring-2 focus-visible:ring-brand">
            <Wordmark className="text-[30px]" />
          </Link>
          <div className="flex items-center gap-1.5">
            {toggle}
            <SiteThemeToggle />
          </div>
        </div>
        <nav aria-label="Main" className="mt-6">
          <Links stars={stars} />
        </nav>
      </div>
      <div aria-hidden="true" className="hidden h-6 xl:block" />
    </>
  );
}
