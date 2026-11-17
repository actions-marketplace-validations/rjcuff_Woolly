import { useInView, useReducedMotion } from 'motion/react';
import { useRef, useState, type ReactNode } from 'react';
import { Link } from 'react-router';
import { PixelWaves } from '@/components/doc/pixel-waves';
import { UNDERLINE } from '@/components/doc/rail';
import { integrations } from '@/lib/integrations';
import type { Release } from '@/lib/releases';
import { githubUrl } from '@/lib/shared';
import { cn } from '@/lib/utils';

/** Setup steps a coding agent can follow on its own, shown in the Raw .md view. */
export const SETUP_PROMPT = `Set up Woolly, the project memory my coding agents share.

1. Install it with curl -fsSL https://usewoolly.com/install.sh | sh
   (on Windows, irm https://usewoolly.com/install.ps1 | iex)
2. Connect my AI apps with woolly connect
3. Keep it running at login with woolly service install
4. In this repo, run woolly init --hooks so every agent shares its memory and every commit checks it
5. Check it works with woolly doctor, and check facts against the code with woolly check

Docs for agents are at https://usewoolly.com/llms.txt`;

const APPS = integrations
  .filter((app) => app.name !== 'Python')
  .map((app) => ({ ...app, name: app.name === 'Claude' ? 'Claude Code' : app.name }));

const DATE = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' });

const BODY = 'text-pretty text-[15px] leading-[1.8]';
const MUTED = 'text-fd-muted-foreground';

function Section({ title, children, aside }: { title: string; children: ReactNode; aside?: ReactNode }) {
  return (
    <section className="mt-28 flex flex-col gap-8 sm:mt-32">
      <div className="flex items-baseline justify-between gap-4">
        <h2 className={cn('text-[12px] uppercase tracking-[0.16em]', MUTED)}>{title}</h2>
        {aside}
      </div>
      <div className="min-w-0">{children}</div>
    </section>
  );
}

const INSTALL = ['curl -fsSL https://usewoolly.com/install.sh | sh', 'woolly connect', 'woolly init --hooks'];

function CopyInstall() {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(INSTALL.join('\n'));
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard can be blocked; the commands are still there to select.
    }
  };
  return (
    <button type="button" onClick={copy} className={cn(UNDERLINE, 'text-[12px]', MUTED)}>
      {copied ? 'Copied' : 'Copy'}
    </button>
  );
}

function Install() {
  return (
    <div className="bg-fd-muted px-5 py-4">
      <pre className="overflow-x-auto text-[13px] leading-[1.9]">
        <code>
          {INSTALL.map((line) => (
            <span key={line} className="block whitespace-pre">
              <span className={MUTED}>$ </span>
              {line}
            </span>
          ))}
        </code>
      </pre>
    </div>
  );
}

const FEATURES: { title: string; body: string }[] = [
  {
    title: 'It knows when a fact goes stale',
    body: 'On every commit, Woolly checks each fact against the code it points at. Rename a function, delete a file or change a pin, and the fact that relied on it is flagged with the commit that broke it.',
  },
  {
    title: 'Every fact cites its source',
    body: 'A file, a function or a few lines, the commit it was saved at, and the session it came from.',
  },
  {
    title: 'One memory for every agent',
    body: 'Claude Code, Cursor and Codex read and update the same notes. Your teammates get them through git.',
  },
  {
    title: 'Sessions start warm',
    body: 'Each session opens with the project’s current facts and where you left off. Stale facts stay out.',
  },
  {
    title: 'Decisions stick',
    body: 'What you chose, why, and what you ruled out, so no agent pitches it again.',
  },
  {
    title: 'Memory you can read',
    body: 'Plain Markdown files. Nothing is ever deleted, and woolly review settles what went stale.',
  },
];

/** Bars made of small squares, filled in Woolly's orange and yellow, that draw in the first time they're on screen. */
function Bars({ title, rows }: { title: string; rows: { label: string; value: number }[] }) {
  const ref = useRef<HTMLDivElement>(null);
  const seen = useInView(ref, { once: true, amount: 0.4 });
  const reduce = useReducedMotion();
  const shown = seen || reduce;
  return (
    <div ref={ref}>
      <h3 className="text-[15px]">{title}</h3>
      <div className="mt-6 flex flex-col gap-5">
        {rows.map((r, i) => (
          <div key={r.label}>
            <div className="flex items-baseline justify-between gap-3 text-[13px]">
              <span className={MUTED}>{r.label}</span>
              <span className="tabular-nums">{r.value.toFixed(1)}%</span>
            </div>
            <div className="mt-2 h-2.5 bg-[repeating-linear-gradient(90deg,var(--color-fd-muted)_0_4px,transparent_4px_5px)]">
              <div
                className="h-full [mask-image:repeating-linear-gradient(90deg,#000_0_4px,transparent_4px_5px)] bg-[linear-gradient(90deg,#f97316,#f59e0b_55%,#facc15)]"
                style={{
                  width: `${r.value}%`,
                  clipPath: shown ? 'inset(0 0 0 0)' : 'inset(0 100% 0 0)',
                  transition: reduce ? 'none' : `clip-path 800ms cubic-bezier(0.23, 1, 0.32, 1) ${i * 110}ms`,
                }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

const STALE_BARS = [
  { label: 'Precision', value: 100 },
  { label: 'Recall', value: 93.3 },
  { label: 'Renames caught', value: 100 },
  { label: 'Changed values caught', value: 86.7 },
  { label: 'Safe edits left alone', value: 100 },
];

const LOCOMO_BARS = [
  { label: 'Overall', value: 83.2 },
  { label: 'Dates and time', value: 95.6 },
  { label: 'Multi-hop', value: 83.6 },
  { label: 'Single fact', value: 79.5 },
];

export function HomeArticle({ download, releases }: { download: { href: string; label: string }; releases: Release[] }) {
  return (
    <article id="top">
      <h1 className="text-balance text-[24px] font-normal leading-[1.35] tracking-[-0.01em] sm:text-[28px]">
        Memory that knows when it’s wrong.
      </h1>
      <p className={cn(BODY, 'mt-6 max-w-[560px]', MUTED)}>
        Woolly gives Claude Code, Cursor and Codex one local memory for your repo. Every fact points at the code it came
        from. When that code changes, Woolly flags the fact instead of letting your agent follow it.
      </p>
      <div className="mt-10 flex flex-wrap gap-x-8 gap-y-3 text-[14px]">
        <Link to="/docs/quickstart" className={UNDERLINE}>
          Get started
        </Link>
        <a href={download.href} className={UNDERLINE}>
          {download.label}
        </a>
        <Link to="/docs/stale-facts" className={UNDERLINE}>
          How it works
        </Link>
      </div>

      <figure className="mt-20 sm:mt-24">
        <PixelWaves />
      </figure>

      <Section title="Install" aside={<CopyInstall />}>
        <Install />
      </Section>

      <Section title="What it does">
        <ul className="flex flex-col gap-9">
          {FEATURES.map((f) => (
            <li key={f.title}>
              <h3 className="text-[15px]">{f.title}</h3>
              <p className={cn('mt-2 text-pretty text-[14px] leading-[1.75]', MUTED)}>{f.body}</p>
            </li>
          ))}
        </ul>
      </Section>

      <Section title="Numbers">
        <div className="flex flex-col gap-14">
          <Bars title="Stale facts, 50 scripted code changes" rows={STALE_BARS} />
          <Bars title="LOCOMO, answer in the top 20 results" rows={LOCOMO_BARS} />
          <p className={cn('text-[13px] leading-relaxed', MUTED)}>
            35 ms for a woolly check, 19 ms median recall. On a laptop with no hosted model.{' '}
            <a href={`${githubUrl}/blob/main/bench/RESULTS.md`} className={UNDERLINE}>
              How we measure
            </a>
          </p>
        </div>
      </Section>

      <Section title="Works with">
        <ul className="flex flex-wrap gap-x-6 gap-y-3 text-[14px]">
          {APPS.map((app) => (
            <li key={app.name}>
              <Link to={app.href} className={UNDERLINE}>
                {app.name}
              </Link>
            </li>
          ))}
        </ul>
      </Section>

      <Section title="Also">
        <p className={cn(BODY, MUTED)}>
          The same engine has an{' '}
          <Link to="/docs/memory-api" className={cn(UNDERLINE, 'text-fd-foreground')}>
            API shaped like mem0’s
          </Link>{' '}
          for the agents you build, and{' '}
          <Link to="/docs/proxy" className={cn(UNDERLINE, 'text-fd-foreground')}>
            woolly proxy
          </Link>{' '}
          gives Ollama a memory. No account, no cloud, no telemetry.
        </p>
      </Section>

      <Section title="Shipping weekly">
        <ol className="flex flex-col">
          {releases.map((r) => (
            <li key={r.version}>
              <a href={r.href} className="group grid grid-cols-[72px_1fr] gap-x-4 py-2.5 text-[14px] sm:grid-cols-[72px_64px_1fr]">
                <span>v{r.version}</span>
                {r.date ? (
                  <time dateTime={r.date} className={cn('max-sm:hidden', MUTED)}>
                    {DATE.format(new Date(r.date))}
                  </time>
                ) : (
                  <span className="max-sm:hidden" />
                )}
                <span className={cn('line-clamp-2 leading-relaxed group-hover:text-fd-foreground', MUTED)}>{r.summary}</span>
              </a>
            </li>
          ))}
        </ol>
      </Section>
    </article>
  );
}
