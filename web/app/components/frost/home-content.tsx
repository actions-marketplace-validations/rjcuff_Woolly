import { motion, useReducedMotion } from 'motion/react';
import { useEffect, useState, type ReactNode } from 'react';
import { Link } from 'react-router';
import { DitherMammoth } from '@/components/frost/dither-mammoth';
import { WRAP } from '@/components/frost/nav';
import { blueskyUrl, githubUrl, linkedinUrl, teamUrl } from '@/lib/shared';
import { cn } from '@/lib/utils';

/* Everything below the nav is one blue on ice white. Panels marked "solid"
   swap the two, the way a print run swaps ink and paper. */

const MONO = 'font-mono';

/* --------------------------------------------------------------- pieces */

export function Pill({ href, children, solid, onBlue }: { href: string; children: ReactNode; solid?: boolean; onBlue?: boolean }) {
  const cls = cn(
    'btn-press inline-flex h-12 items-center gap-4 rounded-full border px-6 text-[16px] outline-none transition-colors focus-visible:ring-2 focus-visible:ring-offset-2',
    onBlue
      ? solid
        ? 'border-[var(--paper)] bg-[var(--paper)] text-[var(--blue)] focus-visible:ring-[var(--paper)] focus-visible:ring-offset-[var(--blue)]'
        : 'border-[var(--paper)] text-[var(--paper)] hover:bg-[var(--paper)] hover:text-[var(--blue)] focus-visible:ring-[var(--paper)] focus-visible:ring-offset-[var(--blue)]'
      : solid
        ? 'border-[var(--blue)] bg-[var(--blue)] text-[var(--paper)] focus-visible:ring-[var(--blue)]'
        : 'border-[var(--blue)] text-[var(--blue)] hover:bg-[var(--blue)] hover:text-[var(--paper)] focus-visible:ring-[var(--blue)]',
  );
  const inner = (
    <>
      <span className="pill-arrow" aria-hidden="true" />
      {children}
    </>
  );
  return href.startsWith('/') ? (
    <Link to={href} className={cls}>
      {inner}
    </Link>
  ) : (
    <a href={href} className={cls}>
      {inner}
    </a>
  );
}

const INSTALL = {
  unix: 'curl -fsSL https://usewoolly.com/install.sh | sh',
  windows: 'irm https://usewoolly.com/install.ps1 | iex',
} as const;

/**
 * The copy icon turns into a check: the two sheets shrink into the middle
 * while the check draws itself in, and back again after a moment.
 */
export function CopyMorph({ done }: { done: boolean }) {
  const reduce = useReducedMotion();
  const t = reduce ? { duration: 0 } : { duration: 0.28, ease: [0.23, 1, 0.32, 1] as const };
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.75} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" className="size-[18px]">
      <motion.g initial={false} animate={done ? { scale: 0.3, opacity: 0 } : { scale: 1, opacity: 1 }} transition={t} style={{ transformBox: 'fill-box', transformOrigin: 'center' }}>
        <rect x="8" y="8" width="13" height="13" rx="2" />
        <path d="M4 16V5a1 1 0 0 1 1-1h11" />
      </motion.g>
      <motion.path d="M5 12.5l4.5 4.5L19 7.5" strokeWidth={2} initial={false} animate={done ? { pathLength: 1, opacity: 1 } : { pathLength: 0, opacity: 0 }} transition={reduce ? { duration: 0 } : { duration: 0.32, delay: done ? 0.08 : 0, ease: [0.23, 1, 0.32, 1] }} />
    </svg>
  );
}

function Install({ onBlue }: { onBlue?: boolean }) {
  const [os, setOS] = useState<keyof typeof INSTALL>('unix');
  const [copied, setCopied] = useState(false);
  const reduce = useReducedMotion();
  useEffect(() => {
    if (/Windows/i.test(navigator.userAgent)) setOS('windows');
  }, []);
  useEffect(() => {
    if (!copied) return;
    const t = window.setTimeout(() => setCopied(false), 1600);
    return () => window.clearTimeout(t);
  }, [copied]);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(INSTALL[os]);
      setCopied(true);
    } catch {
      // Clipboard can be blocked; the command is still there to select.
    }
  };
  return (
    <div className={cn('w-full max-w-[560px]', onBlue ? 'text-[var(--paper)]' : 'text-[var(--blue)]')}>
      <div className={cn('relative flex items-stretch overflow-hidden rounded-[14px] border', onBlue ? 'border-[var(--paper)]' : 'border-[var(--blue)]')}>
        <code className={cn(MONO, 'flex min-w-0 flex-1 items-center overflow-x-auto whitespace-nowrap px-5 py-3.5 text-[13.5px] sm:text-[14px]')}>
          <span className="mr-3 opacity-50">$</span>
          {/* Switching systems fades the new command in where the old one was. */}
          <motion.span key={os} initial={reduce ? false : { opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.18 }}>
            {INSTALL[os]}
          </motion.span>
        </code>
        <button
          type="button"
          onClick={copy}
          aria-label={copied ? 'Copied' : 'Copy install command'}
          className={cn(
            'no-press relative flex w-12 shrink-0 items-center justify-center overflow-hidden rounded-r-[13px] border-l outline-none transition-colors duration-150',
            onBlue
              ? 'border-[var(--paper)] hover:bg-[var(--paper)] hover:text-[var(--blue)] focus-visible:bg-[var(--paper)] focus-visible:text-[var(--blue)]'
              : 'border-[var(--blue)] hover:bg-[var(--blue)] hover:text-[var(--paper)] focus-visible:bg-[var(--blue)] focus-visible:text-[var(--paper)]',
          )}
        >
          <CopyMorph done={copied} />
        </button>
      </div>
      <div role="group" aria-label="Operating system" className="mt-3 flex gap-2 text-[13px]">
        {(
          [
            ['unix', 'macOS & Linux'],
            ['windows', 'Windows'],
          ] as const
        ).map(([id, label]) => (
          <button
            key={id}
            type="button"
            aria-pressed={os === id}
            onClick={() => setOS(id)}
            className={cn(
              'rounded-full border px-3 py-1 outline-none transition-[opacity,border-color] duration-150 focus-visible:ring-2',
              onBlue ? 'focus-visible:ring-[var(--paper)]' : 'focus-visible:ring-[var(--blue)]',
              os === id ? (onBlue ? 'border-[var(--paper)]' : 'border-[var(--blue)]') : 'border-transparent opacity-60 hover:opacity-100',
            )}
          >
            {label}
          </button>
        ))}
      </div>
    </div>
  );
}

/* ----------------------------------------------------------------- hero */

export function HeroContent() {
  return (
    <div className="flex flex-col gap-8 px-5 py-10 text-[var(--blue)] sm:px-10 sm:py-14 lg:px-12 lg:py-[clamp(32px,6vh,72px)]">
      <div>
        <Link to="/docs/memory-files" className="group inline-flex items-center gap-3 text-[14px] outline-none">
          <span className="shrink-0 whitespace-nowrap rounded-full border border-[var(--blue)] px-2.5 py-0.5 text-[12px] leading-[1.6]">New</span>
          <span className="link-soft">woolly check reads CLAUDE.md and AGENTS.md</span>
        </Link>
        <h1 className="mt-8 max-w-[620px] text-balance text-[44px] font-light leading-[1.02] tracking-[-0.035em] sm:text-[60px] lg:text-[clamp(48px,6.4vh,76px)]">
          Memory that knows when it’s wrong.
        </h1>
        <p className="mt-6 max-w-[500px] text-[17px] leading-[1.55] sm:text-[18px]">
          Your agents trust CLAUDE.md, AGENTS.md and their memory. Woolly checks every line against the code and flags the ones a commit made wrong. It runs locally with no setup.
        </p>
      </div>

      <div className="flex flex-col gap-6">
        <Install />
        <div className="flex flex-wrap gap-3">
          <Pill href="/docs/quickstart" solid>
            Get started
          </Pill>
          <Pill href="/docs">Read the docs</Pill>
        </div>
      </div>
    </div>
  );
}

/* --------------------------------------------------------------- panels */

function Code({ lines, onBlue }: { lines: ReactNode[]; onBlue?: boolean }) {
  return (
    <pre className={cn(MONO, 'overflow-x-auto rounded-[18px] p-5 text-[12.5px] leading-[1.8] sm:p-6', onBlue ? 'bg-[var(--paper)] text-[var(--blue)]' : 'bg-[var(--blue)] text-[var(--paper)]')}>
      <code>
        {lines.map((l, i) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: lines of a fixed sample.
          <span key={i} className="block whitespace-pre-wrap break-words">
            {l}
          </span>
        ))}
      </code>
    </pre>
  );
}


/** A short list inside a panel, with a hairline between items. */
function Rows({ items, onBlue }: { items: [string, string][]; onBlue?: boolean }) {
  return (
    <dl className={cn('border-t', onBlue ? 'border-[var(--paper)]/30' : 'border-[var(--blue)]/25')}>
      {items.map(([k, v]) => (
        <div key={k} className={cn('grid gap-1 border-b py-4 sm:grid-cols-[150px_minmax(0,1fr)] sm:gap-6', onBlue ? 'border-[var(--paper)]/30' : 'border-[var(--blue)]/25')}>
          <dt className="text-[15px]">{k}</dt>
          <dd className="text-[15px] leading-[1.55] opacity-80">{v}</dd>
        </div>
      ))}
    </dl>
  );
}

const PANELS: { step: string; title: string; body: string; onBlue?: boolean; show: (onBlue?: boolean) => ReactNode }[] = [
  {
    step: 'What it is',
    title: 'A fact-checker for your agents’ notes',
    body: 'It reads every note your agents load, from every tool. Then it checks each line against your code and its history.',
    show: (onBlue) => (
      <Rows
        onBlue={onBlue}
        items={[
          ['Agent files', 'CLAUDE.md, AGENTS.md, Cursor rules'],
          ['Agent memory', 'Claude Code, Codex, claude-mem'],
          ['Answer', 'In about a second'],
        ]}
      />
    ),
  },
  {
    step: 'Real repos',
    title: 'We checked 20 open source repos',
    body: 'Agent files in TanStack, LangChain.js, the Anthropic SDK and more. Woolly flagged 11 lines naming code that’s gone, and all 11 were right.',
    onBlue: true,
    show: (onBlue) => (
      <Code
        onBlue={onBlue}
        lines={[
          '$ woolly check',
          '  packages/AGENTS.md:8',
          '    useValueStable was removed in d740d8bf2d',
          '3 lines name code that’s gone.',
        ]}
      />
    ),
  },
  {
    step: 'How it works',
    title: 'Git knows what changed',
    body: 'A line is flagged only when git shows the commit that took its code away. No model and no guessing, so it doesn’t cry wolf.',
    onBlue: true,
    show: (onBlue) => (
      <ol className={cn('flex flex-col border-t', onBlue ? 'border-[var(--paper)]/30' : 'border-[var(--blue)]/25')}>
        {[
          ['Read', 'Every note your agents load'],
          ['Flag', 'Lines naming code that’s gone'],
          ['Fix', 'woolly check --fix renames them for you'],
        ].map(([k, v], i) => (
          <li key={k} className={cn('grid grid-cols-[32px_minmax(0,1fr)] gap-4 border-b py-4', onBlue ? 'border-[var(--paper)]/30' : 'border-[var(--blue)]/25')}>
            <span className={cn('flex size-7 items-center justify-center rounded-full border text-[12px]', onBlue ? 'border-[var(--paper)]' : 'border-[var(--blue)]')}>{i + 1}</span>
            <div>
              <div className="text-[15px]">{k}</div>
              <div className="mt-1 text-[15px] leading-[1.55] opacity-80">{v}</div>
            </div>
          </li>
        ))}
      </ol>
    ),
  },
  {
    step: 'Where it runs',
    title: 'On your machine, in plain text',
    body: 'No account, no cloud and no telemetry. Project memory is a Markdown file in your repo that you can read and commit.',
    show: (onBlue) => (
      <Code
        onBlue={onBlue}
        lines={[
          '# .woolly/woolly.md',
          ' ',
          '- Use pnpm, not npm',
          '- API tests need Postgres on 5433',
        ]}
      />
    ),
  },
];

export function Panels() {
  return (
    <div className="flex flex-col">
      {PANELS.map((p, i) => (
        <section
          key={p.step}
          aria-label={p.step}
          className={cn(
            'flex min-w-0 flex-col justify-start gap-9 px-5 py-14 sm:px-10 sm:py-16 lg:min-h-[calc(100svh-3.5rem)] lg:px-12 lg:py-[clamp(32px,6vh,72px)]',
            p.onBlue ? 'bg-[var(--blue)] text-[var(--paper)]' : 'text-[var(--blue)]',
            !p.onBlue && i > 0 && 'border-t border-[var(--blue)]/20',
          )}
        >
          <span className={cn('self-start whitespace-nowrap rounded-full border px-2.5 py-0.5 text-[12px] leading-[1.6]', p.onBlue ? 'border-[var(--paper)]' : 'border-[var(--blue)]')}>{p.step}</span>
          <div>
            <h2 className="max-w-[580px] text-balance text-[36px] font-light leading-[1.04] tracking-[-0.03em] sm:text-[50px]">{p.title}</h2>
            <p className="mt-6 max-w-[540px] text-[17px] leading-[1.6] opacity-90 sm:text-[18px]">{p.body}</p>
          </div>
          <div className="max-w-[600px]">{p.show(p.onBlue)}</div>
        </section>
      ))}
    </div>
  );
}

/* --------------------------------------------------------------- numbers */

const NUMBERS = [
  { value: '11 of 11', label: 'stale-line flags were right', sub: '20 open source repos' },
  { value: '4 of 20', label: 'repos had a line naming code that’s gone', sub: '2,990 lines checked' },
  { value: '0.7 s', label: 'for a woolly check', sub: 'median, on a laptop' },
  { value: '0', label: 'cloud calls', sub: 'no account, no telemetry' },
];

export function Numbers() {
  return (
    <div className="text-[var(--blue)]">
      <div className="flex flex-wrap items-end justify-between gap-6">
        <h2 className="max-w-[560px] text-[34px] font-light leading-[1.05] tracking-[-0.03em] sm:text-[48px]">Measured on real repos</h2>
        <a href={`${githubUrl}/blob/main/bench/RESULTS.md`} className="link-soft text-[15px] outline-none">
          How we measure
        </a>
      </div>
      <p className="mt-4 max-w-[560px] text-[16.5px] leading-[1.6] opacity-90">
        The agent files of 20 open source repos, checked with git and nothing else. Every flag is published with why it’s right.
      </p>
      <dl className="mt-12 grid grid-cols-2 border-t border-[var(--blue)]/25 lg:grid-cols-4">
        {NUMBERS.map((n, i) => (
          <div key={n.label} className={cn('flex flex-col gap-2 border-b border-[var(--blue)]/25 py-8 pr-4', i % 2 === 1 && 'max-lg:pl-5 max-lg:border-l', i > 0 && 'lg:border-l lg:pl-6')}>
            <dd className="order-1 text-[48px] font-light leading-none tracking-[-0.04em] sm:text-[64px]">{n.value}</dd>
            <dt className="order-2 text-[15px]">{n.label}</dt>
            <span className="order-3 text-[13px] opacity-70">{n.sub}</span>
          </div>
        ))}
      </dl>
    </div>
  );
}

/* -------------------------------------------------------------- audiences */

const WHO: { title: string; body: string; link: { label: string; href: string } }[] = [
  {
    title: 'For the agents you use',
    body: 'One command connects Claude Code, Cursor, Codex and more. They all share the same memory.',
    link: { label: 'Connect your apps', href: '/docs/connect' },
  },
  {
    title: 'For your team',
    body: 'Project memory lives in git. Every teammate’s agent starts from the same facts, checked against the code.',
    link: { label: 'Project memory', href: '/docs/project-memory' },
  },
  {
    title: 'For agents you build',
    body: 'A local HTTP API with TypeScript and Python clients. Give your own agents memory that knows when it’s stale.',
    link: { label: 'Memory API', href: '/docs/memory-api' },
  },
];

export function Audiences() {
  return (
    <div className="grid gap-px overflow-hidden rounded-[22px] border border-[var(--blue)] bg-[var(--blue)] text-[var(--blue)] md:grid-cols-3">
      {WHO.map((w) => (
        <div key={w.title} className="flex flex-col gap-4 bg-[var(--paper)] p-6 sm:p-8">
          <h3 className="text-[26px] font-light leading-[1.1] tracking-[-0.025em]">{w.title}</h3>
          <p className="text-[15.5px] leading-[1.6] opacity-90">{w.body}</p>
          <Link to={w.link.href} className="group mt-auto inline-flex items-center gap-3 pt-4 text-[15px] outline-none">
            <span className="pill-arrow" aria-hidden="true" />
            {w.link.label}
          </Link>
        </div>
      ))}
    </div>
  );
}

/* ------------------------------------------------------------------ teams */

const TEAM_ROWS: [string, string][] = [
  ['acme/api', '3 to review'],
  ['acme/web', '2 to review'],
  ['acme/worker', 'all reviewed'],
];

/** Woolly Team: the view across a team's repos, for when one repo's pull request comments aren't enough. */
export function TeamCallout() {
  return (
    <div className="grid items-center gap-12 text-[var(--blue)] lg:grid-cols-2">
      <div className="flex min-w-0 flex-col gap-6">
        <span className="self-start whitespace-nowrap rounded-full border border-[var(--blue)] px-2.5 py-0.5 text-[12px] leading-[1.6]">Woolly Team</span>
        <h2 className="max-w-[580px] text-balance text-[36px] font-light leading-[1.04] tracking-[-0.03em] sm:text-[50px]">Running agents across a team’s repos?</h2>
        <p className="max-w-[480px] text-[16.5px] leading-[1.6] opacity-90">See every repo on one page, what each pull request made stale, and a review queue your team clears together. It never reads your code.</p>
        <p className="text-[15px]">
          <span className="text-[28px] font-light tracking-[-0.03em]">$299</span> a month per organization. Free during the pilot.
        </p>
        <div className="flex flex-wrap gap-3">
          <Pill href={teamUrl} solid>
            Try Woolly Team
          </Pill>
          <Pill href={`${teamUrl}/#pricing`}>See pricing</Pill>
        </div>
      </div>
      <div aria-hidden="true" className="rounded-[22px] border border-[var(--blue)] bg-[var(--blue)] p-5 text-[var(--paper)] sm:p-6">
        <div className="grid grid-cols-3 gap-2.5">
          {[
            ['Needs review', '5'],
            ['Coverage', '83%'],
            ['Repos', '12'],
          ].map(([k, v]) => (
            <div key={k} className="flex flex-col gap-1 rounded-[12px] bg-[var(--paper)]/10 p-3">
              <span className="text-[12px] opacity-80">{k}</span>
              <span className="text-[30px] font-light leading-none tracking-[-0.03em]">{v}</span>
            </div>
          ))}
        </div>
        <ul className="mt-3">
          {TEAM_ROWS.map(([repo, state]) => (
            <li key={repo} className="flex justify-between gap-3 border-t border-[var(--paper)]/25 px-0.5 py-2.5 text-[15px]">
              <span>{repo}</span>
              <span className="opacity-80">{state}</span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

/* -------------------------------------------------------------------- end */

export function Closing() {
  return (
    <div className="relative grid items-end gap-12 text-[var(--paper)] lg:grid-cols-[minmax(0,1fr)_minmax(0,360px)]">
      <div className="flex min-w-0 flex-col gap-8">
        <h2 className="max-w-[620px] text-balance text-[40px] font-light leading-[1.02] tracking-[-0.035em] sm:text-[60px]">Try it on the repo you’re in.</h2>
        <p className="max-w-[480px] text-[17px] leading-[1.6] opacity-90">Install it and keep working. Woolly tells you when a commit breaks a note.</p>
        <Install onBlue />
        <div className="flex flex-wrap gap-3">
          <Pill href="/docs/quickstart" solid onBlue>
            Get started
          </Pill>
          <Pill href="/download" onBlue>
            Download the app
          </Pill>
        </div>
      </div>
      <DitherMammoth className="w-full max-w-[360px] justify-self-end max-lg:hidden" />
    </div>
  );
}

/* ----------------------------------------------------------------- footer */

const FOOT: { label: string; href: string }[] = [
  { label: 'Docs', href: '/docs' },
  { label: 'Download', href: '/download' },
  { label: 'Changelog', href: `${githubUrl}/blob/main/CHANGELOG.md` },
  { label: 'GitHub', href: githubUrl },
  { label: 'Bluesky', href: blueskyUrl },
  { label: 'LinkedIn', href: linkedinUrl },
  { label: 'Privacy', href: '/privacy' },
  { label: 'Terms', href: '/terms' },
];

export function Foot() {
  return (
    <footer className="page-foot py-10 text-[var(--blue)]">
      <div className={cn(WRAP, 'flex flex-col gap-6 md:flex-row md:items-center md:justify-between')}>
        <span className="text-[14px]">© {new Date().getFullYear()} Woolly™ · MIT licensed</span>
        <nav aria-label="Footer">
          <ul className="flex flex-wrap gap-x-5 gap-y-2 text-[14px]">
            {FOOT.map((l) => (
              <li key={l.label}>
                {l.href.startsWith('/') ? (
                  <Link to={l.href} className="link-soft outline-none">
                    {l.label}
                  </Link>
                ) : (
                  <a href={l.href} className="link-soft outline-none">
                    {l.label}
                  </a>
                )}
              </li>
            ))}
          </ul>
        </nav>
      </div>
    </footer>
  );
}
