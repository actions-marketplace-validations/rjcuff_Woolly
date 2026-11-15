import { useInView, useReducedMotion } from 'motion/react';
import { useRef } from 'react';
import { FaqWall, type FaqItem } from '@/components/easeui/faq-wall';
import { NumberTicker } from '@/components/easeui/number-ticker';
import { PrimaryButton } from '@/components/ui';
import { EASE_OUT_CSS } from '@/lib/ease';
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

const FAQ: FaqItem[] = [
  {
    question: 'Does Woolly send my data anywhere?',
    answer: 'No. Woolly has no servers, no accounts and no tracking. Your memories stay in ~/.woolly on your computer.',
    keywords: ['privacy', 'cloud', 'server', 'tracking', 'data', 'offline'],
  },
  {
    question: 'Which apps work with Woolly?',
    answer:
      'Any app that supports MCP, including Claude Code, Claude Desktop, Cursor, VS Code, Windsurf, Codex, Gemini CLI, LM Studio and opencode. Run woolly connect to set them all up at once.',
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
      "Those memories live inside one app, on that company's servers. Woolly is shared by every app you connect, and it lives on your computer in files you can read.",
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
    question: 'Is Woolly free?',
    answer: 'Yes. Woolly is open source under the MIT license.',
    keywords: ['price', 'cost', 'license', 'mit', 'open', 'pay'],
  },
];

export function Faq() {
  return (
    <FaqWall
      className="px-0 sm:px-0"
      title="Questions"
      items={FAQ}
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
