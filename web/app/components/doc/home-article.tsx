import { ArrowUpRight, FileText, Pin, RefreshCw, Sunrise, Users, type LucideIcon } from 'lucide-react';
import { useInView, useReducedMotion } from 'motion/react';
import { useEffect, useRef, useState, type ReactNode } from 'react';
import { Link } from 'react-router';
import { BlockButton, FrameButton, Label, MoreLink, SmartLink } from '@/components/doc/parts';
import { PixelBrain } from '@/components/doc/pixel-brain';
import type { DocSection } from '@/components/doc/rail';
import { InsightCard } from '@/components/motion/insight-card';
import { NumberTicker } from '@/components/motion/number-ticker';
import { EASE_OUT_CSS } from '@/lib/ease';
import { FAQ } from '@/lib/faq';
import { integrations } from '@/lib/integrations';
import type { Release } from '@/lib/releases';
import { githubUrl } from '@/lib/shared';
import { cn } from '@/lib/utils';

export const SECTIONS: DocSection[] = [
  { id: 'top', label: 'Overview' },
  { id: 'why', label: 'Why Woolly' },
  { id: 'what', label: 'What it does' },
  { id: 'numbers', label: 'By the numbers' },
  { id: 'anywhere', label: 'Runs anywhere' },
  { id: 'private', label: 'Private' },
  { id: 'shipping', label: 'Shipping weekly' },
  { id: 'questions', label: 'Questions' },
];

/** Setup steps a coding agent can follow on its own, shown in the Raw .md view. */
export const SETUP_PROMPT = `Set up Woolly, the local memory shared by my AI tools.

1. Install it: curl -fsSL https://usewoolly.com/install.sh | sh
   (On Windows: irm https://usewoolly.com/install.ps1 | iex)
2. Connect my AI apps: woolly connect
3. Keep it running at login: woolly service install
4. Check it works: woolly doctor

Docs for agents: https://usewoolly.com/llms.txt`;

// The apps with a logo to show, each with its name beside it.
const APPS = integrations
  .filter((app) => app.name !== 'Python')
  .map((app) => ({ ...app, name: app.name === 'Claude' ? 'Claude Code' : app.name }));

const DATE = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' });

const P = 'text-pretty text-[17px] leading-[1.7] text-fd-foreground/90';
const INLINE = 'text-brand underline-offset-4 hover:underline';
const NUDGE = 'transition-transform duration-200 ease-ui group-hover:-translate-y-0.5 group-hover:translate-x-0.5 motion-reduce:transition-none';

/** False only for a section that starts below the screen, until it scrolls in, so server HTML always shows it. */
function useReveal(ref: React.RefObject<HTMLElement | null>) {
  const [shown, setShown] = useState(true);
  useEffect(() => {
    const el = ref.current;
    if (!el || el.getBoundingClientRect().top < window.innerHeight) return;
    setShown(false);
    const io = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setShown(true);
          io.disconnect();
        }
      },
      { rootMargin: '0px 0px -12% 0px' },
    );
    io.observe(el);
    return () => io.disconnect();
  }, [ref]);
  return shown;
}

function Section({ id, children, className }: { id: string; children: ReactNode; className?: string }) {
  const ref = useRef<HTMLElement>(null);
  const shown = useReveal(ref);
  return (
    <section ref={ref} id={id} className={cn('scroll-mt-10 border-t border-fd-border pt-14 pb-16', className)}>
      <div className="reveal" data-shown={shown}>
        {children}
      </div>
    </section>
  );
}

function H2({ children }: { children: ReactNode }) {
  return <h2 className="text-balance text-[28px] font-normal leading-tight tracking-[-0.02em] sm:text-[32px]">{children}</h2>;
}

/** Rolls from zero to the value the first time it's on screen. */
function Rolling({ value, decimals = 0 }: { value: number; decimals?: number }) {
  const ref = useRef<HTMLSpanElement>(null);
  const seen = useInView(ref, { once: true, amount: 0.6 });
  const reduce = useReducedMotion();
  return (
    <span ref={ref}>
      <NumberTicker value={seen || reduce ? value : 0} format={(n) => n.toFixed(decimals)} />
    </span>
  );
}

function WorksWith() {
  return (
    <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
      <Label>Works with</Label>
      <ul className="flex flex-wrap items-center gap-x-5 gap-y-2 text-fd-muted-foreground">
        {APPS.map((app) => (
          <li key={app.name}>
            <Link
              to={app.href}
              aria-label={app.name}
              className="flex size-8 items-center justify-center transition-[color,scale] duration-300 ease-ui hover:scale-[1.18] hover:text-fd-foreground focus-visible:text-fd-foreground motion-reduce:transition-none"
            >
              <svg viewBox="0 0 24 24" className="size-5" fill="currentColor" aria-hidden="true">
                <path d={app.path} />
              </svg>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

function Brain() {
  return (
    <figure className="mt-14">
      <PixelBrain />
      <figcaption className="mt-5 text-[13px] text-fd-muted-foreground">
        Every square is a memory.<span className="hidden [@media(hover:hover)]:inline"> Hover to recall one.</span>
      </figcaption>
    </figure>
  );
}

function Hero({ download, latest }: { download: { href: string; label: string }; latest?: Release }) {
  return (
    <header className="pb-16 pt-14">
      {latest ? (
        <SmartLink
          href={latest.href}
          className="group inline-flex max-w-full items-stretch bg-fd-muted text-[14px] shadow-[inset_0_0_0_1px_var(--color-fd-border)]"
        >
          <span className="flex items-center gap-2.5 px-3 py-1.5 text-fd-muted-foreground">
            <span aria-hidden="true" className="relative flex size-1.5">
              <span className="absolute inset-0 animate-ping rounded-full bg-green-500 opacity-60 motion-reduce:hidden" />
              <span className="relative size-1.5 rounded-full bg-green-500" />
            </span>
            <span className="truncate">New in v{latest.version}</span>
          </span>
          <span className="m-[3px] flex items-center gap-1.5 bg-fd-background px-2.5 font-medium shadow-[0_0_0_1px_var(--color-fd-border)]">
            Read the notes
            <ArrowUpRight aria-hidden="true" className={cn('size-3.5', NUDGE)} />
          </span>
        </SmartLink>
      ) : null}

      <h1 className="mt-8 text-[34px] font-normal leading-[1.08] tracking-[-0.03em] sm:text-[46px]">
        your ai agents <br />
        share one brain
      </h1>
      <p className="mt-6 max-w-[560px] text-pretty text-[18px] leading-relaxed text-fd-muted-foreground sm:text-[19px]">
        Claude Code, Cursor, Codex and Ollama read and write one memory. It lives on your laptop, in Markdown files you
        can read.
      </p>

      <div className="mt-9 flex flex-wrap gap-3">
        <BlockButton href="/docs/quickstart">Get started</BlockButton>
        <FrameButton href={download.href}>{download.label}</FrameButton>
      </div>

      <div className="mt-8">
        <WorksWith />
      </div>

      <Brain />
    </header>
  );
}

const FEATURES: { icon: LucideIcon; title: string; body: ReactNode }[] = [
  {
    icon: Sunrise,
    title: 'Sessions that start warm.',
    body: 'Claude Code opens every session with your profile, the project’s notes and a handoff from last time. Nobody has to say “here’s my stack” again.',
  },
  {
    icon: Pin,
    title: 'Decisions stick.',
    body: 'When a session ends, Woolly saves what you chose, why, and what you ruled out. No agent pitches the rejected option again.',
  },
  {
    icon: RefreshCw,
    title: 'Facts stay current.',
    body: 'A newer fact replaces the old one. When an AI guesses something that contradicts you, your version wins and the guess waits for your review.',
  },
  {
    icon: FileText,
    title: 'Memory you can read',
    body: 'Every memory is a Markdown file. Open it, grep it, fix it by hand, and see which tool saved it. The Woolly app shows the same thing with search.',
  },
  {
    icon: Users,
    title: 'One project memory for the team.',
    body: (
      <>
        <code className="font-mono text-[15px]">woolly init</code> keeps the repo’s stack, conventions and gotchas in{' '}
        <code className="font-mono text-[15px]">.woolly/woolly.md</code>. Commit it, and every teammate’s agent reads and
        updates the same notes.
      </>
    ),
  },
];

/** Bars that draw in once, the first time they're on screen. */
function Bars({ rows }: { rows: { label: string; value: number }[] }) {
  const ref = useRef<HTMLDivElement>(null);
  const seen = useInView(ref, { once: true, amount: 0.5 });
  const reduce = useReducedMotion();
  const shown = seen || reduce;
  return (
    <div ref={ref} className="flex flex-col gap-5">
      {rows.map((r, i) => (
        <div key={r.label}>
          <div className="flex items-baseline justify-between gap-3">
            <Label>{r.label}</Label>
            <span className="text-[32px] leading-none tracking-[-0.02em]">
              <Rolling value={r.value} decimals={1} />
              <span className="text-[16px] text-fd-muted-foreground">%</span>
            </span>
          </div>
          <div className="mt-3 h-3 bg-[repeating-linear-gradient(90deg,var(--color-fd-muted)_0_3px,transparent_3px_4px)]">
            <div
              className="h-full bg-[repeating-linear-gradient(90deg,var(--color-brand)_0_3px,transparent_3px_4px)]"
              style={{
                width: `${r.value}%`,
                clipPath: shown ? 'inset(0 0 0 0)' : 'inset(0 100% 0 0)',
                transition: reduce ? 'none' : `clip-path 700ms ${EASE_OUT_CSS} ${i * 120}ms`,
              }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}

const STAT = 'rounded-none bg-fd-background p-5 shadow-[0_0_0_1px_var(--color-fd-border)] [&>div:first-child>span]:font-mono [&>div:first-child>span]:text-[11px] [&>div:first-child>span]:uppercase [&>div:first-child>span]:tracking-[0.14em]';

function StatValue({ children, unit }: { children: ReactNode; unit?: string }) {
  return (
    <span className="mt-4 block text-[44px] font-normal leading-none tracking-[-0.03em]">
      {children}
      {unit ? <span className="ml-0.5 text-[18px] text-fd-muted-foreground">{unit}</span> : null}
    </span>
  );
}

const PLACES: { title: string; body: string; href: string }[] = [
  { title: 'On your laptop', body: 'Mac, Windows and Linux. One small binary, or the app.', href: '/download' },
  { title: 'On a Raspberry Pi', body: 'Sync your machines over your own network, encrypted.', href: '/docs/sync' },
  { title: 'In your repo', body: 'Project memory in git, shared with your whole team.', href: '/docs/project-memory' },
];

const PROMISES: { title: string; body: string }[] = [
  { title: 'No account', body: 'Nothing to sign up for.' },
  { title: 'No cloud', body: 'Your memory never leaves your machines.' },
  { title: 'No telemetry', body: 'Woolly never phones home.' },
];

export function HomeArticle({ download, releases }: { download: { href: string; label: string }; releases: Release[] }) {
  return (
    <article id="top">
      <Hero download={download} latest={releases[0]} />

      <Section id="why">
        <H2>Why Woolly</H2>
        <div className="mt-6 flex flex-col gap-5">
          <p className={P}>
            Every AI coding tool starts from zero. You explain your stack to Cursor, then again to Claude Code, then to
            Codex. Each keeps its own memory, or none, and none of them know what you decided yesterday or what your
            teammate’s agent learned this morning.
          </p>
          <p className={P}>
            Woolly is <strong className="font-medium">the memory they share</strong>. It’s one folder on your laptop,
            plain Markdown, with no cloud and no account. Any tool that speaks{' '}
            <Link to="/docs/mcp" className={INLINE}>
              MCP
            </Link>{' '}
            can save to it and search it, and{' '}
            <Link to="/docs/proxy" className={INLINE}>
              woolly proxy
            </Link>{' '}
            gives the same memory to Ollama and any app that uses the OpenAI or Anthropic API.
          </p>
          <p className={P}>
            When a session starts, your agent gets your profile, the project’s notes and{' '}
            <Link to="/docs/how-it-works" className={INLINE}>
              where you left off
            </Link>
            . When it ends, Woolly saves what you decided and why.
          </p>
        </div>
      </Section>

      <Section id="what">
        <H2>What it does</H2>
        <ol className="mt-10 flex flex-col">
          {FEATURES.map((f) => (
            <li key={f.title} className="group/feature flex gap-5 border-t border-fd-border py-6 first:border-t-0 first:pt-0">
              <span
                aria-hidden="true"
                className="flex size-10 shrink-0 items-center justify-center text-fd-muted-foreground shadow-[0_0_0_1px_var(--color-fd-border)] transition-[color,background-color,box-shadow] duration-200 ease-ui group-hover/feature:bg-brand/[0.07] group-hover/feature:text-brand group-hover/feature:shadow-[0_0_0_1px_color-mix(in_oklab,var(--color-brand)_35%,transparent)]"
              >
                <f.icon className="size-5 transition-transform duration-300 ease-ui group-hover/feature:-rotate-6 group-hover/feature:scale-110 motion-reduce:transition-none" strokeWidth={1.75} />
              </span>
              <div>
                <h3 className="text-[17px] font-medium">{f.title}</h3>
                <p className="mt-1.5 text-pretty text-[16px] leading-[1.7] text-fd-muted-foreground">{f.body}</p>
              </div>
            </li>
          ))}
        </ol>
        <div className="mt-8">
          <MoreLink href="/docs/how-it-works">Read how it works</MoreLink>
        </div>
      </Section>

      <Section id="numbers">
        <p className="text-balance text-[26px] leading-[1.3] tracking-[-0.02em] text-fd-muted-foreground sm:text-[30px]">
          Our <span className="text-fd-foreground">promise</span> is memory you can{' '}
          <span className="text-fd-foreground">check</span>, not a number you have to trust.
        </p>
        <div className="mt-10 grid gap-3 sm:grid-cols-2">
          <div className="grid gap-3">
            <InsightCard
              className={STAT}
              label="Median recall time"
              value={
                <StatValue unit="ms">
                  <Rolling value={23} />
                </StatValue>
              }
            />
            <InsightCard
              className={STAT}
              label="Apps it connects"
              value={
                <StatValue>
                  <Rolling value={9} />
                </StatValue>
              }
            />
            <InsightCard className={STAT} label="Servers to trust" value={<StatValue>0</StatValue>} />
          </div>
          <div className="flex flex-col bg-fd-background p-5 shadow-[0_0_0_1px_var(--color-fd-border)]">
            <p className="text-[16px] font-medium">LOCOMO benchmark</p>
            <Label className="mt-1 text-[10px]">Long conversation memory</Label>
            <div className="mt-8">
              <Bars
                rows={[
                  { label: 'Found the answer', value: 86.7 },
                  { label: 'Answered right', value: 56.7 },
                ]}
              />
            </div>
            <p className="mt-8 text-[14px] leading-relaxed text-fd-muted-foreground">
              Answered by qwen2.5:7b on a laptop, not a hosted model. Run the same test yourself:
            </p>
            <code className="mt-3 block overflow-x-auto bg-fd-muted px-3 py-2 font-mono text-[12px] whitespace-nowrap">
              python3 bench/locomo/run.py
            </code>
            <div className="mt-auto pt-6">
              <MoreLink href="/docs/benchmark">How we measure</MoreLink>
            </div>
          </div>
        </div>
      </Section>

      <Section id="anywhere">
        <H2>
          Woolly runs where your code does,
          <br className="max-sm:hidden" /> with no cloud at all.
        </H2>
        <p className={cn(P, 'mt-5')}>
          It’s one binary with no outside dependencies. Install it on your laptop, sync it to your other machines, or
          commit the project’s memory so your team shares it.
        </p>
        <ul className="mt-10 grid gap-3 sm:grid-cols-3">
          {PLACES.map((p) => (
            <li key={p.title}>
              <SmartLink
                href={p.href}
                className="group flex h-full flex-col bg-fd-background p-5 shadow-[0_0_0_1px_var(--color-fd-border)] transition-shadow duration-200 hover:shadow-[0_0_0_1px_var(--color-brand)]"
              >
                <div className="flex items-start justify-between gap-2">
                  <p className="text-[17px] font-medium">{p.title}</p>
                  <ArrowUpRight aria-hidden="true" className={cn('size-4 shrink-0 text-fd-muted-foreground group-hover:text-brand', NUDGE)} />
                </div>
                <p className="mt-1.5 text-[14px] leading-relaxed text-fd-muted-foreground">{p.body}</p>
              </SmartLink>
            </li>
          ))}
        </ul>
      </Section>

      <Section id="private">
        <H2>Private by design.</H2>
        <p className={cn(P, 'mt-5 text-fd-muted-foreground')}>
          Woolly refuses passwords and API keys, and an AI can’t plant instructions in your memory. Lines that come from a
          repo are checked for secrets and injected instructions before any agent sees them.
        </p>
        <ul className="mt-10 grid gap-3 sm:grid-cols-3">
          {PROMISES.map((p) => (
            <li key={p.title} className="bg-fd-background p-5 shadow-[0_0_0_1px_var(--color-fd-border)]">
              <Label className="text-fd-foreground">{p.title}</Label>
              <p className="mt-3 text-[14px] text-fd-muted-foreground">{p.body}</p>
            </li>
          ))}
        </ul>
        <div className="mt-8 flex flex-wrap gap-x-8 gap-y-3">
          <MoreLink href="/docs/security">Security</MoreLink>
          <MoreLink href={githubUrl}>Read the code</MoreLink>
        </div>
      </Section>

      <Section id="shipping">
        <H2>Shipping weekly</H2>
        <p className={cn(P, 'mt-4 text-fd-muted-foreground')}>Built in public, one release at a time. Here’s what landed lately.</p>
        <ol className="mt-8 shadow-[0_0_0_1px_var(--color-fd-border)]">
          {releases.map((r) => (
            <li key={r.version} className="border-b border-fd-border last:border-b-0">
              <a href={r.href} className="group flex items-start gap-4 px-4 py-4 transition-colors duration-150 hover:bg-fd-muted/60">
                <span aria-hidden="true" className="mt-0.5 flex size-6 shrink-0 items-center justify-center text-fd-muted-foreground shadow-[0_0_0_1px_var(--color-fd-border)] transition-colors duration-200 group-hover:bg-brand group-hover:text-brand-foreground group-hover:shadow-none">
                  <ArrowUpRight className={cn('size-3.5', NUDGE)} />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="flex flex-wrap items-baseline gap-x-3">
                    <span className="font-mono text-[14px] font-medium text-brand">v{r.version}</span>
                    {r.date ? (
                      <time dateTime={r.date} className="text-[13px] text-fd-muted-foreground tabular-nums">
                        {DATE.format(new Date(r.date))}
                      </time>
                    ) : null}
                  </span>
                  <span className="mt-1 line-clamp-2 block text-pretty text-[15px] leading-relaxed">{r.summary}</span>
                </span>
              </a>
            </li>
          ))}
        </ol>
        <div className="mt-6">
          <MoreLink href={`${githubUrl}/blob/main/CHANGELOG.md`}>Read the changelog</MoreLink>
        </div>
      </Section>

      <Section id="questions">
        <H2>Questions</H2>
        <dl className="mt-8 flex flex-col gap-8">
          {FAQ.map((q) => (
            <div key={q.question}>
              <dt className="text-[17px] font-medium">{q.question}</dt>
              <dd className="mt-2 text-pretty text-[16px] leading-[1.7] text-fd-muted-foreground">{q.answer}</dd>
            </div>
          ))}
        </dl>
        <p className="mt-10 text-[15px] text-fd-muted-foreground">
          Still have a question?{' '}
          <a href={`${githubUrl}/issues`} className={INLINE}>
            Ask on GitHub
          </a>
          .
        </p>
      </Section>

    </article>
  );
}
