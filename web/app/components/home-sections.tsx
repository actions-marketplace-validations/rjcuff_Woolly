import { Link } from 'react-router';

const RESULTS: { label: string; note: string; score: number }[] = [
  { label: 'Found the answer', note: 'Woolly returned the message that holds the answer', score: 86.7 },
  { label: 'Answered right', note: 'A 7B model on a laptop answered from those memories', score: 56.7 },
];

/** LOCOMO results, drawn as bars. */
export function Benchmark() {
  return (
    <section className="flex w-full flex-col items-center gap-10 py-20 sm:py-28">
      <div className="flex max-w-[640px] flex-col items-center gap-5 text-center">
        <p className="rounded-md bg-fd-background px-3 py-1.5 text-[14px] font-medium shadow-[0_0_0_1px_var(--color-fd-border)]">
          Measured on a laptop
        </p>
        <h2 className="text-balance text-[32px] font-semibold leading-tight tracking-[-0.02em] sm:text-[48px]">
          Benchmarking Woolly
        </h2>
        <p className="text-pretty text-[17px] leading-relaxed text-fd-muted-foreground sm:text-[19px]">
          Tested with LOCOMO, the long conversation benchmark memory services report. Everything runs on your own
          computer with Ollama, so you can check every number.
        </p>
        <Link
          to="/docs/benchmark"
          className="btn-brand flex items-center gap-2 rounded-md py-2 pr-2 pl-4 text-[15px] font-medium"
        >
          Run it yourself
          <span aria-hidden="true" className="flex size-6 items-center justify-center rounded bg-black/20 text-[13px]">
            →
          </span>
        </Link>
      </div>

      <div className="flex w-full max-w-[860px] flex-col gap-5">
        {RESULTS.map((r) => (
          <div key={r.label}>
            {/* On phones the bars are too short for the label, so it sits above. */}
            <div className="mb-2 flex items-center justify-between font-mono text-[14px] font-semibold uppercase tracking-wide sm:hidden">
              <span>{r.label}</span>
              <span className="tabular-nums">{r.score}%</span>
            </div>
            <div className="relative h-8 overflow-hidden rounded-md bg-black/[0.05] sm:h-14 dark:bg-white/[0.06]">
              <div
                className="absolute inset-y-0 left-0 rounded-md bg-[linear-gradient(90deg,var(--color-brand),var(--color-brand-light))]"
                style={{ width: `${r.score}%` }}
              />
              <div
                className="relative flex h-full items-center justify-between px-4 font-mono text-[16px] font-semibold uppercase tracking-wide text-white max-sm:hidden"
                style={{ width: `${r.score}%` }}
              >
                <span className="truncate">{r.label}</span>
                <span className="tabular-nums">{r.score}%</span>
              </div>
            </div>
            <p className="mt-2 text-[14px] text-fd-muted-foreground">{r.note}</p>
          </div>
        ))}
        <p className="mt-2 text-pretty text-[14px] text-fd-muted-foreground">
          First 30 questions of LOCOMO conversation 1, with <code className="font-mono">qwen2.5:7b</code> answering.
          Median search time 23 ms. Published scores use large hosted models, so they aren&apos;t directly comparable.
        </p>
      </div>
    </section>
  );
}

const FAQ: { q: string; a: string }[] = [
  {
    q: 'Does Woolly send my data anywhere?',
    a: 'No. Woolly has no servers, no accounts and no tracking. Your memories stay in ~/.woolly on your computer.',
  },
  {
    q: 'Which apps work with Woolly?',
    a: 'Any app that supports MCP, including Claude Code, Claude Desktop, Cursor, VS Code, Windsurf, Codex, Gemini CLI, LM Studio and opencode. Run woolly connect to set them all up at once.',
  },
  {
    q: 'Do I need Ollama?',
    a: 'No. Woolly works on its own. With Ollama running, it also searches by meaning, learns facts from your chats and saves decisions when a session ends.',
  },
  {
    q: 'How is this different from ChatGPT or Claude memory?',
    a: "Those memories live inside one app, on that company's servers. Woolly is shared by every app you connect, and it lives on your computer in files you can read.",
  },
  {
    q: 'What happens when a fact changes?',
    a: "A new memory with the same key replaces the old one. If an AI saves something that disagrees with what you said, your version stays and the AI's is saved as a conflict for you to review.",
  },
  {
    q: 'Can my team share memory?',
    a: "Yes. Run woolly init in a repo and commit .woolly/. Every teammate's agent reads and updates the same project memory, and facts about you never go in the repo.",
  },
  {
    q: 'Is Woolly free?',
    a: 'Yes. Woolly is open source under the MIT license.',
  },
];

export function Faq() {
  return (
    <section className="grid w-full gap-10 py-20 sm:py-28 md:grid-cols-[0.8fr_1.2fr]">
      <div>
        <h2 className="text-[32px] font-semibold leading-tight tracking-[-0.02em] sm:text-[40px]">Questions</h2>
        <p className="mt-3 text-[17px] text-fd-muted-foreground">
          More in the{' '}
          <Link to="/docs/faq" className="font-medium text-brand hover:underline">
            FAQ
          </Link>
          .
        </p>
      </div>
      <div className="divide-y divide-fd-border border-y border-fd-border">
        {FAQ.map((item) => (
          <details key={item.q} className="group">
            <summary className="flex cursor-pointer list-none items-center justify-between gap-6 py-5 text-[17px] font-medium [&::-webkit-details-marker]:hidden">
              {item.q}
              <span
                aria-hidden="true"
                className="text-[22px] leading-none text-fd-muted-foreground transition-transform duration-200 group-open:rotate-45"
              >
                +
              </span>
            </summary>
            <p className="-mt-1 pb-5 pr-10 text-pretty text-[16px] leading-relaxed text-fd-muted-foreground">
              {item.a}
            </p>
          </details>
        ))}
      </div>
    </section>
  );
}
