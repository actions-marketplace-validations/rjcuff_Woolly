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
        <p className="text-[17px]">One command adds Woolly to the AI tools you already use</p>
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

const FEATURES: { icon: string; title: string; body: string; href: string }[] = [
  {
    icon: 'link',
    title: 'Shared',
    body: 'Claude Code, Cursor and Ollama read and write one memory. Tell one, and the rest know.',
    href: '/docs/connect',
  },
  {
    icon: 'file',
    title: 'Readable',
    body: 'Every memory is a Markdown file. Open it, edit it by hand, grep it or back it up.',
    href: '/docs/how-it-works',
  },
  {
    icon: 'refresh',
    title: 'Current',
    body: 'A newer fact replaces the old one, and what you said beats an AI’s guess.',
    href: '/docs/how-it-works',
  },
  {
    icon: 'shield',
    title: 'Private',
    body: 'Runs on your computer. No account, no cloud, and passwords and keys are refused.',
    href: '/docs/security',
  },
  {
    icon: 'users',
    title: 'Team ready',
    body: 'Project memory lives in your repo, so every teammate’s agent knows the decisions.',
    href: '/docs/project-memory',
  },
  {
    icon: 'list',
    title: 'Auditable',
    body: 'See which app saved, read or deleted each memory, and when.',
    href: '/docs/review',
  },
];

/** Why Woolly, as a grid of feature cards. */
export function Features() {
  return (
    <section className="flex w-full flex-col items-center gap-14 py-20 sm:py-28">
      <div className="flex max-w-[680px] flex-col items-center gap-5 text-center">
        <h2 className="text-balance text-[32px] font-semibold leading-[1.1] tracking-[-0.02em] sm:text-[48px]">
          Built for developers
          <br />
          Designed for privacy
        </h2>
        <p className="text-pretty text-[17px] leading-relaxed text-fd-muted-foreground sm:text-[19px]">
          Most AI memory is a black box on someone else&apos;s servers. Woolly&apos;s is a folder on your computer that
          every AI tool shares, and that you can read, check and change.
        </p>
      </div>
      <ul className="grid w-full gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {FEATURES.map((f) => (
          <li key={f.title}>
            <Link
              to={f.href}
              className="flex h-full flex-col rounded-xl bg-fd-background p-6 shadow-[0_0_0_1px_var(--color-fd-border)] transition-shadow hover:shadow-[0_0_0_1px_var(--color-brand)]"
            >
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
                  <path d={ICONS[f.icon]} />
                </svg>
              </span>
              <h3 className="mt-6 text-[17px] font-semibold">{f.title}</h3>
              <p className="mt-2 text-pretty text-[15px] leading-relaxed text-fd-muted-foreground">{f.body}</p>
            </Link>
          </li>
        ))}
      </ul>
    </section>
  );
}
