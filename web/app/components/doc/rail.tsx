import { Link } from 'react-router';
import { SiteThemeToggle } from '@/components/brand';
import { formatStars, STAR_THRESHOLD } from '@/lib/github';
import { githubUrl } from '@/lib/shared';
import { cn } from '@/lib/utils';

export type Mode = 'human' | 'agent';
/** A section of the page, with the memory file it reads as. */
export type DocSection = { id: string; label: string };

export const UNDERLINE =
  'underline decoration-fd-foreground/35 decoration-1 underline-offset-[6px] transition-[text-decoration-color] duration-150 hover:decoration-fd-foreground focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand';

const NAV_LINK = 'group relative inline-block py-1 text-fd-muted-foreground outline-none transition-colors duration-200 hover:text-fd-foreground focus-visible:text-fd-foreground';
const BRACKET = 'pointer-events-none absolute top-0 opacity-0 transition-[opacity,translate] duration-200 group-hover:translate-x-0 group-hover:opacity-100 group-focus-visible:translate-x-0 group-focus-visible:opacity-100 motion-reduce:transition-none';

/** A nav link: no underline, and brackets that slide in around it on hover. */
function NavLink({ href, children, label }: { href: string; children: React.ReactNode; label?: string }) {
  const inner = (
    <>
      <span aria-hidden="true" className={cn(BRACKET, '-left-3 translate-x-1.5')}>
        [
      </span>
      {children}
      <span aria-hidden="true" className={cn(BRACKET, '-right-3 -translate-x-1.5')}>
        ]
      </span>
    </>
  );
  return href.startsWith('/') ? (
    <Link to={href} className={NAV_LINK} aria-label={label}>
      {inner}
    </Link>
  ) : (
    <a href={href} className={NAV_LINK} aria-label={label}>
      {inner}
    </a>
  );
}

/** The name, linking home. */
export function Logo() {
  return (
    <Link to="/" className="font-sans text-[19px] font-medium tracking-[-0.02em] outline-none focus-visible:ring-2 focus-visible:ring-brand">
      Woolly
    </Link>
  );
}

/** The top of every page except the docs: the name, and three plain links. */
export function HeroNav({ stars }: { stars?: number | null }) {
  const showStars = typeof stars === 'number' && stars >= STAR_THRESHOLD;
  return (
    <header className="flex items-center justify-between gap-6 pt-10 pb-16 sm:pt-16 sm:pb-24">
      <Logo />
      <nav aria-label="Main">
        <ul className="flex items-center gap-5 text-[14px] sm:gap-9">
          <li>
            <NavLink href="/docs">Docs</NavLink>
          </li>
          <li>
            <NavLink href="/download">Download</NavLink>
          </li>
          <li>
            <NavLink href={githubUrl} label={showStars ? `GitHub, ${stars} stars` : undefined}>
              GitHub{showStars ? <span className="ml-1.5 opacity-70">{formatStars(stars)}</span> : null}
            </NavLink>
          </li>
          <li className="-my-2 -mr-2">
            <SiteThemeToggle />
          </li>
        </ul>
      </nav>
    </header>
  );
}
