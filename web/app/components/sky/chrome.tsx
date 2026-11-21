import { Menu, X } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { Link } from 'react-router';
import { GitHubIcon } from '@/components/brand';
import { DitherMammoth } from '@/components/frost/dither-mammoth';
import { formatStars, STAR_THRESHOLD } from '@/lib/github';
import { githubUrl, linkedinUrl } from '@/lib/shared';
import { cn } from '@/lib/utils';
import { CopyInstall, WRAP } from './parts';

const ITEM = 'inline-flex min-h-10 items-center rounded-full px-2.5 text-[14px] sm:px-3.5 sm:text-[14.5px] font-medium text-[var(--ink)] outline-none transition-colors duration-150 hover:bg-white/70 focus-visible:bg-white/70';

/** A frosted pill that floats over the sky. */
const PILL = 'flex items-center rounded-full bg-white/60 p-1 shadow-[0_1px_2px_rgb(27_35_80/0.06),0_4px_16px_-8px_rgb(27_35_80/0.18)] ring-1 ring-white/90 backdrop-blur-md';

/**
 * The nav floats over the hero sky. The name and the links sit together on the
 * left; GitHub and Get started sit in a second pill on the right.
 */
export function SkyNav({ stars, className }: { stars?: number | null; className?: string }) {
  const showStars = typeof stars === 'number' && stars >= STAR_THRESHOLD;
  return (
    <header className={cn('z-20', className)}>
      <div className={cn(WRAP, 'flex h-20 items-center justify-between gap-3 sm:h-24')}>
        <div className="flex min-w-0 items-center gap-6">
          <Link to="/" aria-label="Woolly home" className="flex shrink-0 items-center gap-2 rounded-lg text-[var(--ink)] outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]">
            <DitherMammoth small className="h-[28px] w-auto text-[var(--ink)]" />
            <span className="serif text-[24px] leading-none">woolly</span>
          </Link>
          <nav aria-label="Main" className="max-md:hidden">
            <ul className={PILL}>
              <li>
                <Link to="/docs" className={ITEM}>
                  Docs
                </Link>
              </li>
              <li>
                <Link to="/download" className={ITEM}>
                  Download
                </Link>
              </li>
              <li>
                <Link to="/team" className={ITEM}>
                  Teams
                </Link>
              </li>
              <li className="max-md:hidden">
                <a href={`${githubUrl}/blob/main/bench/RESULTS.md`} className={ITEM}>
                  Benchmarks
                </a>
              </li>
            </ul>
          </nav>
        </div>
        <div className={cn(PILL, 'shrink-0 gap-1 max-md:hidden')}>
          <a href={githubUrl} aria-label={showStars ? `GitHub, ${stars} stars` : 'GitHub'} className={cn(ITEM, 'gap-1.5')}>
            <GitHubIcon className="size-[17px]" />
            {showStars ? <span className="text-[13px] tabular-nums">{formatStars(stars)}</span> : null}
          </a>
          <Link
            to="/docs/quickstart"
            className="btn-press inline-flex min-h-10 items-center rounded-full bg-[var(--action)] px-4 text-[14.5px] font-semibold text-white outline-none transition-colors duration-150 hover:bg-[var(--action-hover)] focus-visible:ring-2 focus-visible:ring-[var(--action)] focus-visible:ring-offset-2"
          >
            Get started
          </Link>
        </div>
        <MobileMenu stars={showStars ? stars : null} />
      </div>
    </header>
  );
}

const DRAWER_LINKS: { label: string; href: string }[] = [
  { label: 'Docs', href: '/docs' },
  { label: 'Download', href: '/download' },
  { label: 'Teams', href: '/team' },
  { label: 'Benchmarks', href: `${githubUrl}/blob/main/bench/RESULTS.md` },
];

const DRAWER_ITEM = 'flex min-h-12 items-center rounded-xl px-3 text-[17px] font-medium text-[var(--ink)] outline-none transition-colors duration-150 hover:bg-white focus-visible:bg-white';

/**
 * Phones get a menu button that opens the links in a drawer from the right.
 * It's a native modal dialog, so focus stays inside, Escape closes it and the
 * page behind can't be reached until it closes.
 */
function MobileMenu({ stars }: { stars: number | null }) {
  const ref = useRef<HTMLDialogElement>(null);
  const [open, setOpen] = useState(false);

  const close = useCallback(() => {
    const d = ref.current;
    if (!d?.open || d.classList.contains('closing')) return;
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      d.close();
      return;
    }
    d.classList.add('closing');
    const done = () => {
      d.classList.remove('closing');
      d.close();
    };
    // Close when the slide out ends, with a timer in case the event never comes.
    const t = window.setTimeout(done, 260);
    const onEnd = (e: AnimationEvent) => {
      if (e.target !== d) return;
      d.removeEventListener('animationend', onEnd);
      window.clearTimeout(t);
      done();
    };
    d.addEventListener('animationend', onEnd);
  }, []);

  useEffect(() => {
    const d = ref.current;
    if (!d) return;
    const onClose = () => setOpen(false);
    // Escape runs the same slide out instead of vanishing.
    const onCancel = (e: Event) => {
      e.preventDefault();
      close();
    };
    d.addEventListener('close', onClose);
    d.addEventListener('cancel', onCancel);
    return () => {
      d.removeEventListener('close', onClose);
      d.removeEventListener('cancel', onCancel);
    };
  }, [close]);

  // The page behind stays still while the drawer is open.
  useEffect(() => {
    if (!open) return;
    const root = document.documentElement;
    const prev = root.style.overflow;
    root.style.overflow = 'hidden';
    return () => {
      root.style.overflow = prev;
    };
  }, [open]);

  const show = () => {
    ref.current?.showModal();
    setOpen(true);
  };

  return (
    <div className="md:hidden">
      <button
        type="button"
        onClick={show}
        aria-label="Open menu"
        aria-expanded={open}
        aria-haspopup="dialog"
        className={cn(PILL, 'no-press size-11 justify-center p-0 text-[var(--ink)] outline-none transition-colors duration-150 hover:bg-white/80 focus-visible:ring-2 focus-visible:ring-[var(--accent)]')}
      >
        <Menu className="size-5" aria-hidden="true" />
      </button>
      <dialog
        ref={ref}
        aria-label="Menu"
        className="sky-drawer"
        // A click on the dimmed backdrop lands on the dialog itself.
        onClick={(e) => {
          if (e.target === e.currentTarget) close();
        }}
      >
        <div className="flex h-full flex-col px-4 pb-6">
          <div className="flex h-20 items-center justify-between">
            <span className="serif pl-3 text-[22px] leading-none text-[var(--ink)]">woolly</span>
            <button
              type="button"
              onClick={close}
              aria-label="Close menu"
              className="no-press flex size-11 items-center justify-center rounded-full text-[var(--ink)] outline-none transition-colors duration-150 hover:bg-white focus-visible:ring-2 focus-visible:ring-[var(--accent)]"
            >
              <X className="size-5" aria-hidden="true" />
            </button>
          </div>
          <nav aria-label="Menu links">
            <ul className="flex flex-col gap-1">
              {DRAWER_LINKS.map((l) => (
                <li key={l.label}>
                  {l.href.startsWith('/') ? (
                    <Link to={l.href} onClick={close} className={DRAWER_ITEM}>
                      {l.label}
                    </Link>
                  ) : (
                    <a href={l.href} className={DRAWER_ITEM}>
                      {l.label}
                    </a>
                  )}
                </li>
              ))}
              <li>
                <a href={githubUrl} className={cn(DRAWER_ITEM, 'gap-2.5')}>
                  <GitHubIcon className="size-[18px]" />
                  GitHub
                  {stars !== null ? <span className="text-[14px] text-[var(--muted)] tabular-nums">{formatStars(stars)}</span> : null}
                </a>
              </li>
            </ul>
          </nav>
          <Link
            to="/docs/quickstart"
            onClick={close}
            className="btn-press mt-auto inline-flex h-12 items-center justify-center rounded-full bg-[var(--action)] text-[16px] font-semibold text-white outline-none transition-colors duration-150 hover:bg-[var(--action-hover)] focus-visible:ring-2 focus-visible:ring-[var(--action)] focus-visible:ring-offset-2"
          >
            Get started
          </Link>
        </div>
      </dialog>
    </div>
  );
}

const COLUMNS: { title: string; links: { label: string; href: string }[] }[] = [
  {
    title: 'Product',
    links: [
      { label: 'Docs', href: '/docs' },
      { label: 'Download', href: '/download' },
      { label: 'Woolly Team', href: '/team' },
      { label: 'Benchmarks', href: `${githubUrl}/blob/main/bench/RESULTS.md` },
    ],
  },
  {
    title: 'Resources',
    links: [
      { label: 'Quickstart', href: '/docs/quickstart' },
      { label: 'Changelog', href: `${githubUrl}/blob/main/CHANGELOG.md` },
      { label: 'GitHub', href: githubUrl },
    ],
  },
  {
    title: 'Follow',
    links: [
      { label: 'LinkedIn', href: linkedinUrl },
    ],
  },
  {
    title: 'Legal',
    links: [
      { label: 'Privacy', href: '/privacy' },
      { label: 'Terms', href: '/terms' },
    ],
  },
];

/**
 * The footer, in our deep night blue: the name on the left, link columns on
 * the right, then a line to install Woolly and the copyright.
 */
export function SkyFoot() {
  return (
    <footer className="bg-[var(--action)] text-white">
      <div className={cn(WRAP, 'pt-16 pb-10 md:pt-20')}>
        <div className="grid gap-12 md:grid-cols-[1fr_3fr]">
          <Link to="/" className="flex items-center gap-2 self-start rounded-lg outline-none focus-visible:ring-2 focus-visible:ring-white">
            <DitherMammoth small className="h-[28px] w-auto text-[#c3cdff]" />
            <span className="serif text-[24px] leading-none">woolly</span>
          </Link>
          <nav aria-label="Footer" className="grid grid-cols-2 gap-x-8 gap-y-10 sm:grid-cols-4">
            {COLUMNS.map((c) => (
              <div key={c.title}>
                <p className="mb-3 text-[12px] font-bold uppercase tracking-[0.1em] text-white">{c.title}</p>
                <ul>
                  {c.links.map((l) => {
                    const cls = 'inline-flex min-h-9 items-center text-[15px] text-[#c3cdff] outline-none transition-colors duration-150 hover:text-white focus-visible:text-white';
                    return (
                      <li key={l.label}>
                        {l.href.startsWith('/') ? (
                          <Link to={l.href} className={cls}>
                            {l.label}
                          </Link>
                        ) : (
                          <a href={l.href} className={cls}>
                            {l.label}
                          </a>
                        )}
                      </li>
                    );
                  })}
                </ul>
              </div>
            ))}
          </nav>
        </div>
        <div className="mt-16 flex flex-col gap-4 border-t border-white/15 pt-8 md:flex-row md:items-center md:justify-between">
          <div>
            <p className="text-[12px] font-bold uppercase tracking-[0.1em]">Install Woolly</p>
            <p className="mt-1.5 text-[15px] text-[#c3cdff]">One line on macOS, Linux or Windows. No account needed.</p>
          </div>
          <CopyInstall className="self-start text-white hover:text-[#c3cdff] focus-visible:ring-white md:self-auto" />
        </div>
        <p className="mt-10 text-center text-[13px] text-white/55">© {new Date().getFullYear()} Woolly™ · MIT licensed</p>
      </div>
    </footer>
  );
}
