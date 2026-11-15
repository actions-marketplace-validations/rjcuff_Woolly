import { Link } from 'react-router';
import { integrations } from '@/lib/integrations';

/** The apps Woolly works with, as a row of logos with their names. */
export function Integrations() {
  return (
    <section className="flex w-full flex-col items-center gap-10 py-16 sm:py-20">
      <div className="flex flex-col items-center gap-3 text-center">
        <p className="rounded-md bg-fd-foreground px-3 py-1.5 font-mono text-[20px] font-semibold text-fd-background sm:text-[24px]">
          woolly connect
        </p>
        <p className="text-[17px]">One command, and the tools you already use share a memory</p>
      </div>
      <ul className="grid w-full grid-cols-2 gap-x-6 gap-y-8 sm:grid-cols-4">
        {integrations.map((app) => (
          <li key={app.name} className="flex justify-center">
            <Link
              to={app.href}
              className="flex items-center gap-2.5 text-fd-muted-foreground opacity-80 transition hover:text-fd-foreground hover:opacity-100"
            >
              <svg viewBox="0 0 24 24" className="size-7 shrink-0" fill="currentColor" aria-hidden="true">
                <path d={app.path} />
              </svg>
              <span className="text-[20px] font-semibold tracking-[-0.01em] sm:text-[22px]">{app.name}</span>
            </Link>
          </li>
        ))}
      </ul>
    </section>
  );
}

// Simple 24px line icons, drawn with the current text color.
const ICONS: Record<string, string> = {
  file: 'M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8zM14 3v5h5M9 13h6M9 17h4',
  refresh: 'M20 11a8 8 0 0 0-14.9-3M4 4v4h4M4 13a8 8 0 0 0 14.9 3M20 20v-4h-4',
  shield: 'M12 3l7 3v5c0 4.5-3 8.3-7 10-4-1.7-7-5.5-7-10V6zM9 12l2 2 4-4',
  link: 'M10 14a4 4 0 0 0 5.7 0l3-3a4 4 0 0 0-5.7-5.7l-1 1M14 10a4 4 0 0 0-5.7 0l-3 3a4 4 0 0 0 5.7 5.7l1-1',
  users:
    'M16 20v-1a4 4 0 0 0-4-4H7a4 4 0 0 0-4 4v1M9.5 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7zM21 20v-1a4 4 0 0 0-3-3.9M15.5 4.1a3.5 3.5 0 0 1 0 6.8',
  list: 'M9 6h11M9 12h11M9 18h11M4 6h.01M4 12h.01M4 18h.01',
};

type Feature = { icon: string; title: string; body: string; href: string };

const SMALL: Feature[] = [
  {
    icon: 'refresh',
    title: 'Decisions stick',
    body: 'When a session ends, Woolly saves what you decided and what you ruled out, so no agent pitches it again.',
    href: '/docs/how-it-works',
  },
  {
    icon: 'shield',
    title: 'Private',
    body: 'Runs on your laptop. No account, no cloud, no telemetry. Passwords and API keys get turned away.',
    href: '/docs/security',
  },
  {
    icon: 'users',
    title: 'Team ready',
    body: 'Commit .woolly and every teammate’s agent starts with the same stack, conventions and gotchas.',
    href: '/docs/project-memory',
  },
];

const TILE = 'flex h-full flex-col rounded-2xl bg-fd-background p-6 shadow-[0_0_0_1px_var(--color-fd-border)] sm:p-7';

function Icon({ name }: { name: string }) {
  return (
    <span className="flex size-10 items-center justify-center rounded-lg bg-brand/10 text-brand">
      <svg
        viewBox="0 0 24 24"
        className="size-5"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        <path d={ICONS[name]} />
      </svg>
    </span>
  );
}

function More({ href }: { href: string }) {
  return (
    <Link to={href} className="mt-auto pt-6 text-[15px] font-medium text-brand hover:underline">
      Learn more <span aria-hidden="true">→</span>
    </Link>
  );
}

const FILE_LINES: [string, string][] = [
  ['key', 'user.location'],
  ['source', 'user'],
  ['saved_by', 'Claude Code'],
  ['updated_at', '2026-11-10'],
];

/** Why Woolly, as a bento grid: two wide tiles that show it, then three that say it. */
export function Features() {
  return (
    <section className="flex w-full flex-col items-center gap-14 py-20 sm:py-28">
      <div className="flex max-w-[680px] flex-col items-center gap-5 text-center">
        <h2 className="text-[32px] font-semibold leading-[1.1] tracking-[-0.02em] sm:text-[48px]">
          Stop re-explaining
          <br />
          your project
        </h2>
        <p className="text-pretty text-[17px] leading-relaxed text-fd-muted-foreground sm:text-[19px]">
          You switch from Cursor to Claude Code and start over. Again. Woolly is the memory they share, a folder on your
          laptop that you can read, grep and fix.
        </p>
      </div>

      <div className="grid w-full items-stretch gap-5 lg:grid-cols-3">
        <div className={`${TILE} lg:col-span-2`}>
          <Icon name="file" />
          <h3 className="mt-6 text-[19px] font-semibold">Memory you can read</h3>
          <p className="mt-2 max-w-[460px] text-pretty text-[15px] leading-relaxed text-fd-muted-foreground">
            Every memory is a Markdown file. Open it, edit it by hand, grep it, back it up, and see which tool saved it.
          </p>
          <div
            role="img"
            aria-label="A memory file that says Lives in Denver, Colorado"
            className="pointer-events-none mt-6 select-none overflow-hidden rounded-xl bg-fd-card font-mono text-[13px] leading-[1.9] shadow-[0_0_0_1px_var(--color-fd-border)]"
          >
            <p className="border-b border-fd-border px-4 py-2 text-[12px] text-fd-muted-foreground">
              ~/.woolly/memories/20261110-1a2b3c4d.md
            </p>
            <div className="px-4 py-3">
              {FILE_LINES.map(([k, v]) => (
                <p key={k} className="truncate">
                  <span className="text-brand">{k}</span>
                  <span className="text-fd-muted-foreground">: {v}</span>
                </p>
              ))}
              <p className="mt-2 font-semibold">Lives in Denver, Colorado.</p>
            </div>
          </div>
        </div>

        <div className={TILE}>
          <Icon name="link" />
          <h3 className="mt-6 text-[19px] font-semibold">One brain, every tool</h3>
          <p className="mt-2 text-pretty text-[15px] leading-relaxed text-fd-muted-foreground">
            Tell Claude Code once. Cursor, Codex and Gemini already know.
          </p>
          <ul className="mt-6 flex flex-wrap gap-2">
            {integrations.map((app) => (
              <li
                key={app.name}
                className="flex items-center gap-1.5 rounded-md bg-fd-card px-2.5 py-1.5 text-[13px] font-medium shadow-[0_0_0_1px_var(--color-fd-border)]"
              >
                <svg
                  viewBox="0 0 24 24"
                  className="size-3.5 text-fd-muted-foreground"
                  fill="currentColor"
                  aria-hidden="true"
                >
                  <path d={app.path} />
                </svg>
                {app.name}
              </li>
            ))}
          </ul>
          <More href="/docs/connect" />
        </div>

        {SMALL.map((f) => (
          <div key={f.title} className={TILE}>
            <Icon name={f.icon} />
            <h3 className="mt-6 text-[19px] font-semibold">{f.title}</h3>
            <p className="mt-2 text-pretty text-[15px] leading-relaxed text-fd-muted-foreground">{f.body}</p>
            <More href={f.href} />
          </div>
        ))}
      </div>
    </section>
  );
}
