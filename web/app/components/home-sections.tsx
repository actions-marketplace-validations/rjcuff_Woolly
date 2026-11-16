import { useInView, useReducedMotion } from 'motion/react';
import { useRef } from 'react';
import { FaqWall, type FaqItem } from '@/components/easeui/faq-wall';
import { NumberTicker } from '@/components/easeui/number-ticker';
import { Mammoth } from '@/components/brand';
import { PrimaryButton, SecondaryButton } from '@/components/ui';
import { EASE_OUT_CSS } from '@/lib/ease';
import type { Release } from '@/lib/releases';
import { githubUrl } from '@/lib/shared';

const RESULTS: { label: string; score: number }[] = [
  { label: 'Recall', score: 86.7 },
  { label: 'Answer accuracy', score: 56.7 },
];

// Seconds between one bar starting and the next, so they land in order.
const STEP = 0.12;

/** LOCOMO results, as bars that draw in the first time they are seen. */
export function Benchmark() {
  const ref = useRef<HTMLDivElement>(null);
  const seen = useInView(ref, { once: true, amount: 0.4 });
  const reduce = useReducedMotion();
  const shown = seen || reduce;

  return (
    <section className="flex w-full flex-col items-center gap-10 py-20 sm:py-28">
      <div className="flex max-w-[640px] flex-col items-center gap-5 text-center">
        <h2 className="text-[32px] font-semibold leading-tight tracking-[-0.02em] sm:text-[48px]">
          Benchmarking Woolly
        </h2>
        <p className="text-pretty text-[17px] leading-relaxed text-fd-muted-foreground sm:text-[19px]">
          Scored on LOCOMO, the long conversation memory benchmark.
        </p>
        <PrimaryButton href="/docs/benchmark" size="md">
          View benchmark
        </PrimaryButton>
      </div>

      <div ref={ref} className="flex w-full max-w-[860px] flex-col gap-5">
        {RESULTS.map((r, i) => (
          <div key={r.label}>
            {/* On phones the bars are too short for the label, so it sits above. */}
            <div className="mb-2 flex items-center justify-between font-mono text-[14px] font-semibold uppercase tracking-wide sm:hidden">
              <span>LOCOMO {r.label}</span>
              <NumberTicker value={r.score} decimals={1} delay={i * STEP} entrance={false} />
            </div>
            <div className="relative h-8 rounded-md bg-black/[0.05] sm:h-14 dark:bg-white/[0.06]">
              {/* Revealed through a rounded clip-path, so the round ends never squash mid-draw. */}
              <div
                className="absolute inset-y-0 left-0 rounded-md bg-[linear-gradient(90deg,var(--color-brand),var(--color-brand-light))]"
                style={{
                  width: `${r.score}%`,
                  clipPath: shown ? 'inset(0 0 0 0 round 6px)' : 'inset(0 100% 0 0 round 6px)',
                  transition: reduce ? 'none' : `clip-path 600ms ${EASE_OUT_CSS} ${i * STEP}s`,
                }}
              >
                <div className="flex h-full items-center justify-between px-4 font-mono text-[16px] font-semibold uppercase tracking-wide text-white max-sm:hidden">
                  <span className="truncate">LOCOMO {r.label}</span>
                  <NumberTicker value={r.score} decimals={1} delay={i * STEP} entrance={false} />
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}

const DATE = new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' });

/** The newest releases from the changelog, so visitors can see how fast it moves. */
export function Shipping({ releases }: { releases: Release[] }) {
  if (releases.length === 0) return null;
  return (
    <section className="flex w-full flex-col items-center gap-10 py-20 sm:py-28">
      <div className="flex max-w-[640px] flex-col items-center gap-5 text-center">
        <Mammoth className="size-10 text-brand" />
        <h2 className="text-[32px] font-semibold leading-tight tracking-[-0.02em] sm:text-[48px]">Shipping weekly</h2>
        <p className="text-pretty text-[17px] leading-relaxed text-fd-muted-foreground sm:text-[19px]">
          Built in public, one release at a time. Here is what landed lately.
        </p>
      </div>
      <ol className="grid w-full items-stretch gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {releases.map((r) => (
          <li key={r.version}>
            <a
              href={r.href}
              className="flex h-full flex-col gap-3 rounded-2xl bg-fd-background p-5 shadow-[0_0_0_1px_var(--color-fd-border)] transition-colors hover:bg-fd-accent"
            >
              <span className="flex items-baseline justify-between gap-3">
                <span className="font-mono text-[15px] font-semibold text-brand">v{r.version}</span>
                {r.date ? (
                  <time dateTime={r.date} className="text-[13px] text-fd-muted-foreground tabular-nums">
                    {DATE.format(new Date(r.date))}
                  </time>
                ) : null}
              </span>
              <span className="line-clamp-3 text-pretty text-[15px] leading-relaxed">{r.summary}</span>
            </a>
          </li>
        ))}
      </ol>
      <SecondaryButton href={`${githubUrl}/blob/main/CHANGELOG.md`} size="md">
        Read the changelog
      </SecondaryButton>
    </section>
  );
}

const FAQ: FaqItem[] = [
  {
    question: 'Does Woolly send my data anywhere?',
    answer:
      'No. There are no servers, no accounts and no telemetry. Your memories live in ~/.woolly on your laptop, and that is the whole story.',
    keywords: ['privacy', 'cloud', 'server', 'tracking', 'data', 'offline'],
  },
  {
    question: 'Which apps work with Woolly?',
    answer:
      'Anything that speaks MCP. Claude Code, Claude Desktop, Cursor, VS Code, Windsurf, Codex, Gemini CLI, LM Studio and opencode are one woolly connect away. Ollama and other local models get memory through woolly proxy.',
    keywords: ['claude', 'cursor', 'vscode', 'windsurf', 'codex', 'gemini', 'mcp', 'apps', 'tools'],
  },
  {
    question: 'Do I need Ollama?',
    answer:
      'No. Woolly works on its own. With Ollama running, it also searches by meaning, learns facts from your chats and saves decisions when a session ends.',
    keywords: ['ollama', 'local', 'model', 'llm'],
  },
  {
    question: 'How is this different from ChatGPT or Claude memory?',
    answer:
      "Those memories are locked inside one app, on that company's servers. Claude never hears what you told Cursor. Woolly is one memory every tool shares, and it lives on your laptop in files you can read.",
    keywords: ['chatgpt', 'openai', 'mem0', 'different', 'compare'],
  },
  {
    question: 'What happens when a fact changes?',
    answer:
      "A new memory with the same key replaces the old one. If an AI saves something that disagrees with what you said, your version stays and the AI's is saved as a conflict for you to review.",
    keywords: ['update', 'conflict', 'replace', 'wrong', 'old'],
  },
  {
    question: 'Can my team share memory?',
    answer:
      "Yes. Run woolly init in a repo and commit .woolly/. Every teammate's agent reads and updates the same project memory, and facts about you never go in the repo.",
    keywords: ['team', 'git', 'repo', 'project', 'share'],
  },
  {
    question: 'Is this a memory API for my app?',
    answer:
      'No, and that is on purpose. Woolly makes your own AI workflow better. There are JavaScript and Python packages for scripts and side projects, but if you are adding memory to a product with thousands of users, a hosted memory API is the better fit.',
    keywords: ['api', 'sdk', 'mem0', 'app', 'product', 'build'],
  },
  {
    question: 'Is Woolly free?',
    answer: 'Yes. It is MIT licensed and built in public on GitHub. Issues and pull requests are welcome.',
    keywords: ['price', 'cost', 'license', 'mit', 'open', 'pay'],
  },
];

export function Faq() {
  return (
    <FaqWall
      className="px-0 sm:px-0"
      title="Questions"
      items={FAQ}
      // Opening one never closes another, so the row you tapped stays put.
      multiple
      ask={{ label: 'Ask your own', placeholder: 'Type a question, like "does it work offline"' }}
      contact={{
        title: 'Still have a question?',
        description: 'Open an issue on GitHub and we will answer it there.',
        actionLabel: 'Ask on GitHub',
        actionHref: `${githubUrl}/issues`,
      }}
    />
  );
}
