import { motion, useReducedMotion } from 'motion/react';
import { useEffect, useState, type ReactNode } from 'react';
import { Link } from 'react-router';
import { DitherMammoth } from '@/components/frost/dither-mammoth';
import { WRAP } from '@/components/frost/nav';
import { integrations } from '@/lib/integrations';
import { blueskyUrl, githubUrl, linkedinUrl } from '@/lib/shared';
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

const APPS = [
  { name: 'Claude Code', path: integrations.find((i) => i.name === 'Claude')?.path },
  { name: 'Codex' },
  ...integrations.filter((i) => ['Cursor', 'Gemini CLI', 'Windsurf', 'opencode'].includes(i.name)),
];

export function HeroContent() {
  return (
    <div className="flex flex-col gap-8 px-5 py-10 text-[var(--blue)] sm:px-10 sm:py-14 lg:px-12 lg:py-[clamp(32px,6vh,72px)]">
      <div>
        <Link to="/docs/why" className="group inline-flex items-center gap-3 text-[14px] outline-none">
          <span className="shrink-0 whitespace-nowrap rounded-full border border-[var(--blue)] px-2.5 py-0.5 text-[12px] leading-[1.6]">New in 0.16</span>
          <span className="underline-offset-4 group-hover:underline">woolly why shows the decisions behind any line of code</span>
        </Link>
        <h1 className="mt-8 max-w-[620px] text-balance text-[44px] font-light leading-[1.02] tracking-[-0.035em] sm:text-[60px] lg:text-[clamp(48px,6.4vh,76px)]">
          Memory that knows when it’s wrong.
        </h1>
        <p className="mt-6 max-w-[500px] text-[17px] leading-[1.55] sm:text-[18px]">
          One local memory for Claude Code, Cursor and Codex. Every fact points at the code it’s about, and when a commit changes that code, Woolly flags the fact before your
          agent acts on it.
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

      <div className="mt-4 flex flex-col gap-3 border-t border-[var(--blue)]/25 pt-5">
        <span className="text-[13px] opacity-70">Works with</span>
        <ul className="flex flex-wrap items-center gap-x-5 gap-y-2 text-[14px]">
          {APPS.map((app) => (
            <li key={app.name} className="flex items-center gap-1.5">
              {app.path ? (
                <svg viewBox="0 0 24 24" aria-hidden="true" className="size-[14px] fill-current">
                  <path d={app.path} />
                </svg>
              ) : null}
              {app.name}
            </li>
          ))}
          <li>
            <Link to="/docs/connect" className="underline underline-offset-4 outline-none">
              and more
            </Link>
          </li>
        </ul>
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

const PANELS: { step: string; title: string; body: string; more: string; onBlue?: boolean; show: (onBlue?: boolean) => ReactNode }[] = [
  {
    step: 'What it is',
    title: 'One memory for all your coding agents',
    body: 'Woolly is a small program that runs on your computer. Claude Code, Cursor, Codex and the rest connect to it and share what they learn about your project, so a fact one agent saves is there for the next.',
    more: 'Run woolly connect once and every agent on your machine is set up. The first woolly init in a repo brings in what your CLAUDE.md, AGENTS.md and Cursor rules already say, so you never start from an empty page.',
    show: (onBlue) => (
      <Rows
        onBlue={onBlue}
        items={[
          ['Decisions', 'What you chose, why, and what you ruled out'],
          ['Conventions', 'How this repo does things'],
          ['Gotchas', 'The traps that cost someone an afternoon'],
          ['Commands', 'How to build, test and ship this repo'],
          ['Handoffs', 'Where the last session left off'],
        ]}
      />
    ),
  },
  {
    step: 'Why it matters',
    title: 'Agents forget, and notes go stale',
    body: 'Every new session starts from zero, so you explain the repo again. A notes file helps until the code changes. Then your agent follows a note that stopped being true, and it sounds just as sure.',
    more: 'The usual fix is a longer CLAUDE.md. It grows, nobody prunes it, and after a few weeks of refactors part of it is wrong. Your agent has no way to tell which part.',
    onBlue: true,
    show: (onBlue) => (
      <Rows
        onBlue={onBlue}
        items={[
          ['Without memory', 'You repeat yourself in every session and every tool'],
          ['With a notes file', 'It drifts from the code, and nothing tells you'],
          ['With a memory tool', 'It remembers what was said, not whether it’s still true'],
          ['With Woolly', 'Facts are checked against the code after every commit'],
        ]}
      />
    ),
  },
  {
    step: 'How it works',
    title: 'Every fact is tied to its code',
    body: 'When an agent saves a fact, Woolly records the file or function it’s about and the commit it was true at. After each commit it checks those facts against the code as it is now.',
    more: 'Checking needs no AI model. Woolly reads the code itself, so a check takes milliseconds and works offline. woolly why shows the decisions behind any line and whether the code still backs each one up.',
    show: (onBlue) => (
      <ol className={cn('flex flex-col border-t', onBlue ? 'border-[var(--paper)]/30' : 'border-[var(--blue)]/25')}>
        {[
          ['Save', 'An agent saves a fact along with the code it points at.'],
          ['Check', 'A git hook checks every fact after each commit, in the background.'],
          ['Flag', 'If the code no longer backs a fact up, it’s marked stale with the reason. Nothing is deleted.'],
          ['Review', 'Stale facts stay out of new sessions until you confirm, fix or retire them.'],
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
    body: 'No account, no cloud and no telemetry. A repo’s memory is a Markdown file you can read, edit and commit, so your teammates’ agents start from the same facts.',
    more: 'Passwords and API keys are refused, and an agent can’t plant instructions in your memory. Ollama is optional. With it, Woolly also searches by meaning and tidies up memory while you’re away.',
    onBlue: true,
    show: (onBlue) => (
      <Code
        onBlue={onBlue}
        lines={[
          '# .woolly/woolly.md',
          ' ',
          '## Conventions',
          '- Use pnpm, not npm',
          ' ',
          '## Gotchas',
          '- API tests need Postgres running on 5433',
          ' ',
          '## Decisions',
          '- Keep auth in middleware, not in each route',
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
            <p className="mt-4 max-w-[540px] text-[16px] leading-[1.6] opacity-75 sm:text-[17px]">{p.more}</p>
          </div>
          <div className="max-w-[600px]">{p.show(p.onBlue)}</div>
        </section>
      ))}
    </div>
  );
}

/* --------------------------------------------------------------- numbers */

const NUMBERS = [
  { value: '100%', label: 'of stale flags were right', sub: '28 of 28' },
  { value: '93%', label: 'of wrong facts caught', sub: '28 of 30' },
  { value: '35 ms', label: 'for a woolly check', sub: 'median, on a laptop' },
  { value: '0', label: 'cloud calls', sub: 'no account, no telemetry' },
];

export function Numbers() {
  return (
    <div className="text-[var(--blue)]">
      <div className="flex flex-wrap items-end justify-between gap-6">
        <h2 className="max-w-[560px] text-[34px] font-light leading-[1.05] tracking-[-0.03em] sm:text-[48px]">Measured on a laptop</h2>
        <a href={`${githubUrl}/blob/main/bench/RESULTS.md`} className="text-[15px] underline underline-offset-4 outline-none">
          How we measure
        </a>
      </div>
      <p className="mt-4 max-w-[560px] text-[16.5px] leading-[1.6] opacity-90">
        50 scripted changes to a small ESP32 drone firmware repo: renames, deleted files, changed values, comments and moved code. No hosted model, no API key.
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
    body: 'woolly connect sets up Claude Code, Cursor, Codex, Gemini CLI and the rest. They all read and write the same memory over MCP.',
    link: { label: 'Connect your apps', href: '/docs/connect' },
  },
  {
    title: 'For your team',
    body: 'Project memory is plain Markdown in .woolly/woolly.md, shared through git. Teammates’ agents start from the same facts, and you can read every line.',
    link: { label: 'Project memory', href: '/docs/project-memory' },
  },
  {
    title: 'For agents you build',
    body: 'The same memory is a local HTTP API with TypeScript and Python clients. It also speaks mem0’s routes, so code written for mem0 can point at your machine.',
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

/* -------------------------------------------------------------------- end */

export function Closing() {
  return (
    <div className="relative grid items-end gap-12 text-[var(--paper)] lg:grid-cols-[minmax(0,1fr)_minmax(0,360px)]">
      <div className="flex min-w-0 flex-col gap-8">
        <h2 className="max-w-[620px] text-balance text-[40px] font-light leading-[1.02] tracking-[-0.035em] sm:text-[60px]">Try it on the repo you’re in.</h2>
        <p className="max-w-[480px] text-[17px] leading-[1.6] opacity-90">Install it, run woolly init --hooks, and keep working. When a commit breaks a fact, you’ll see which one and why.</p>
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
                  <Link to={l.href} className="underline-offset-4 outline-none hover:underline">
                    {l.label}
                  </Link>
                ) : (
                  <a href={l.href} className="underline-offset-4 outline-none hover:underline">
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
