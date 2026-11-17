import { Link } from 'react-router';
import { QUIET_LINK, type Mode } from '@/components/doc/rail';
import { blueskyUrl, githubUrl, linkedinUrl } from '@/lib/shared';
import { cn } from '@/lib/utils';

const LINKS: { label: string; href: string }[] = [
  { label: 'Privacy', href: '/privacy' },
  { label: 'Terms', href: '/terms' },
  { label: 'Changelog', href: `${githubUrl}/blob/main/CHANGELOG.md` },
  { label: 'GitHub', href: githubUrl },
  { label: 'Bluesky', href: blueskyUrl },
  { label: 'LinkedIn', href: linkedinUrl },
];

/** The bottom of every page except the docs: small links, and the switches. */
export function DocFooter({ mode, onMode }: { mode?: Mode; onMode?: (mode: Mode) => void }) {
  return (
    <footer className="mt-28 pb-12 text-[14px] text-fd-muted-foreground">
      <nav aria-label="Footer">
        <ul className="flex flex-wrap items-center gap-x-5 gap-y-2">
          <li className="text-fd-foreground">© {new Date().getFullYear()} Woolly™</li>
          {LINKS.map((l) => (
            <li key={l.label}>
              {l.href.startsWith('/') ? (
                <Link to={l.href} className={QUIET_LINK}>
                  {l.label}
                </Link>
              ) : (
                <a href={l.href} className={QUIET_LINK}>
                  {l.label}
                </a>
              )}
            </li>
          ))}
        </ul>
      </nav>
      {mode && onMode ? (
        <div className="mt-6">
          <div role="group" aria-label="Read as" className="flex items-center gap-4">
            {(['human', 'agent'] as const).map((m) => (
              <button
                key={m}
                type="button"
                aria-pressed={mode === m}
                onClick={() => onMode(m)}
                className={cn('outline-none focus-visible:ring-2 focus-visible:ring-brand', mode === m ? 'text-fd-foreground' : QUIET_LINK)}
              >
                {m === 'human' ? 'Page' : 'Raw .md'}
              </button>
            ))}
          </div>
        </div>
      ) : null}
    </footer>
  );
}
