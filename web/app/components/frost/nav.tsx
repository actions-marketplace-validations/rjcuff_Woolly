import { Link } from 'react-router';
import { GitHubIcon } from '@/components/brand';
import { DitherMammoth } from '@/components/frost/dither-mammoth';
import { formatStars, STAR_THRESHOLD } from '@/lib/github';
import { githubUrl } from '@/lib/shared';
import { cn } from '@/lib/utils';

export const WRAP = 'mx-auto w-full max-w-[1080px] px-4 sm:px-8';

const ITEM = 'link-soft text-[15px] outline-none';

/** The homepage nav, in the page's one blue: the dithered mammoth, the name, and a few plain links. */
export function FrostNav({ stars }: { stars?: number | null }) {
  const showStars = typeof stars === 'number' && stars >= STAR_THRESHOLD;
  return (
    <header className="sticky top-0 z-30 border-b border-[var(--blue)]/20 bg-[var(--paper)]/90 font-['Inter_Tight',sans-serif] text-[var(--blue)] backdrop-blur">
      <div className="flex h-14 items-center justify-between gap-4 px-5 sm:px-10 lg:px-12">
        <Link to="/" className="flex shrink-0 items-center gap-2.5 outline-none focus-visible:ring-2 focus-visible:ring-[var(--blue)]">
          <DitherMammoth small className="h-[26px] w-auto" />
          <span className="text-[20px] font-medium tracking-[-0.03em]">woolly</span>
        </Link>
        <nav aria-label="Main" className="flex items-center gap-4 sm:gap-7">
          <ul className="flex items-center gap-4 sm:gap-7">
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
            <li className="max-sm:hidden">
              <a href={`${githubUrl}/blob/main/bench/RESULTS.md`} className={ITEM}>
                Benchmarks
              </a>
            </li>
            <li>
              <a href={githubUrl} aria-label={showStars ? `GitHub, ${stars} stars` : 'GitHub'} className={cn(ITEM, 'flex items-center gap-1.5')}>
                <GitHubIcon className="size-[17px]" />
                {showStars ? <span className="text-[13px] tabular-nums">{formatStars(stars)}</span> : null}
              </a>
            </li>
          </ul>
          <Link
            to="/docs/quickstart"
            className="btn-press inline-flex h-9 items-center rounded-full border border-[var(--blue)] bg-[var(--blue)] px-4 text-[14px] text-[var(--paper)] outline-none focus-visible:ring-2 focus-visible:ring-[var(--blue)] focus-visible:ring-offset-2 max-md:hidden"
          >
            Get started
          </Link>
        </nav>
      </div>
    </header>
  );
}
