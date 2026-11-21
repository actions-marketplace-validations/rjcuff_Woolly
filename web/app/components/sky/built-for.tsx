import { Link } from 'react-router';
import { cn } from '@/lib/utils';
import { H2, WRAP } from './parts';

const CARD = 'flex h-full min-h-[190px] flex-col rounded-3xl bg-[var(--tint-blue)] p-7 sm:p-8';

function CaseText({ c }: { c: Case }) {
  return (
    <>
      <h3 className="serif text-[24px]/[1.15] text-[var(--ink)] sm:text-[26px]">{c.title}</h3>
      <p className="mt-3 max-w-[300px] text-[15.5px] leading-[1.6] text-pretty text-[var(--body)]">{c.body}</p>
    </>
  );
}

export type Case = { title: string; body: string; href?: string };

const CASES: Case[] = [
  {
    title: 'Solo developers',
    body: 'Switch between Claude Code, Cursor and Codex on the same repo and keep one memory that stays true.',
    href: '/docs/connect',
  },
  {
    title: 'Your agent files',
    body: 'Run woolly check on any repo. It flags lines in CLAUDE.md and AGENTS.md the code no longer backs up.',
    href: '/docs/memory-files',
  },
  {
    title: 'Teams',
    body: 'Project memory lives in git, so every teammate’s agent starts from the same checked facts.',
    href: '/docs/project-memory',
  },
  {
    title: 'Pull requests',
    body: 'A GitHub Action comments when a change makes a project fact stale, before it merges.',
    href: '/docs/pull-requests',
  },
  {
    title: 'Agents you build',
    body: 'A local API with TypeScript and Python clients gives your own agents memory that stays accurate.',
    href: '/docs/memory-api',
  },
];

/**
 * Who Woolly is for, as a grid of light blue cards. The top right spot holds
 * the balloon mammoth from the hero, on its own patch of sky.
 */
export function BuiltFor({ title = 'Made for the way you build', cases = CASES }: { title?: string; cases?: Case[] }) {
  return (
    <section aria-label="Built for" className="bg-white py-20 sm:py-28">
      <div className={WRAP}>
        <h2 className={H2}>{title}</h2>
        <ul className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 sm:mt-12">
          {cases.map((c) => (
            <li key={c.title}>
              {c.href ? (
                <Link to={c.href} className={cn(CARD, 'transition-colors duration-150 outline-none hover:bg-[#dce2ff] focus-visible:ring-2 focus-visible:ring-[var(--accent)]')}>
                  <CaseText c={c} />
                </Link>
              ) : (
                <div className={CARD}>
                  <CaseText c={c} />
                </div>
              )}
            </li>
          ))}
          {/* On large screens this sits in the top right; elsewhere it closes the grid. */}
          <li aria-hidden="true" className="lg:col-start-3 lg:row-start-1">
            <div className="h-[220px] overflow-hidden sm:h-full sm:min-h-[190px] rounded-3xl bg-[#9bb3fb]">
              <img
                src="/mammoth-balloon.webp"
                width={920}
                height={880}
                alt=""
                loading="lazy"
                decoding="async"
                draggable={false}
                className="pointer-events-none size-full select-none object-cover object-[50%_35%]"
              />
            </div>
          </li>
        </ul>
      </div>
    </section>
  );
}
