import { Plus } from 'lucide-react';
import type { ReactNode } from 'react';
import { Link } from 'react-router';
import { integrations } from '@/lib/integrations';
import { cn } from '@/lib/utils';
import { ArrowLink, H2, LEAD, WRAP } from './parts';

const TILE =
  'flex aspect-[5/4] flex-col items-center justify-center gap-3 rounded-2xl bg-white px-2 text-center text-[13.5px] font-semibold text-[var(--ink)] shadow-[0_1px_2px_rgb(27_35_80/0.05)] ring-1 ring-[var(--line)] outline-none transition-colors duration-150 hover:bg-[var(--tint-blue)] focus-visible:ring-2 focus-visible:ring-[var(--accent)] sm:text-[15px]';

export type WorksWithTile = { name: string; href: string; icon: ReactNode };

/** A simple-icons logo, drawn in the text color. */
export function LogoIcon({ path }: { path: string }) {
  return (
    <svg viewBox="0 0 24 24" className="size-7 fill-current sm:size-8" aria-hidden="true">
      <path d={path} />
    </svg>
  );
}

const APPS: WorksWithTile[] = integrations.map((app) => ({ name: app.name, href: app.href, icon: <LogoIcon path={app.path} /> }));

/**
 * Text on the left, a grid of tiles with real logos on the right. The
 * homepage shows the apps Woolly works with, ending in a tile for any other
 * MCP app.
 */
export function WorksWith({
  title = 'One set of notes for every agent you use',
  lead = 'Woolly connects over MCP, so Claude, Cursor and the rest all read the same checked memory. Switch tools and nothing gets lost.',
  link = { label: 'See how to connect', href: '/docs/connect' },
  tiles = APPS,
  more = { label: 'Any MCP app', href: '/docs/connect' },
}: {
  title?: string;
  lead?: string;
  link?: { label: string; href: string };
  tiles?: WorksWithTile[];
  more?: { label: string; href: string } | null;
}) {
  return (
    <section aria-label="Works with" className="bg-[var(--mist)] py-20 sm:py-28">
      <div className={cn(WRAP, 'grid items-center gap-12 lg:grid-cols-[minmax(0,1fr)_minmax(0,560px)] lg:gap-16')}>
        <div className="flex flex-col items-start">
          <h2 className={H2}>{title}</h2>
          <p className={cn(LEAD, 'mt-5 max-w-[460px]')}>{lead}</p>
          <ArrowLink href={link.href} className="mt-6">
            {link.label}
          </ArrowLink>
        </div>
        <ul className="grid grid-cols-3 gap-2.5 sm:gap-3">
          {tiles.map((t) => (
            <li key={t.name}>
              <Link to={t.href} className={TILE}>
                {t.icon}
                {t.name}
              </Link>
            </li>
          ))}
          {more ? (
            <li>
              <Link to={more.href} className={cn(TILE, 'border-2 border-dashed border-[#c3cbe8] bg-transparent text-[var(--muted)] shadow-none ring-0 hover:bg-white')}>
                <Plus className="size-7 sm:size-8" aria-hidden="true" />
                {more.label}
              </Link>
            </li>
          ) : null}
        </ul>
      </div>
    </section>
  );
}
