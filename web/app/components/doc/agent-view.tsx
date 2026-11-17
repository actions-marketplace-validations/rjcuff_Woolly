import { Check, Copy } from 'lucide-react';
import { useState } from 'react';
import { SETUP_PROMPT } from '@/components/doc/home-article';
import { Label } from '@/components/doc/parts';
import { FAQ } from '@/lib/faq';
import type { Release } from '@/lib/releases';
import { githubUrl, siteUrl } from '@/lib/shared';

/** The homepage as Markdown, for an agent to read or a person to paste into one. */
export function homeMarkdown(releases: Release[]): string {
  const shipped = releases.map((r) => `- **v${r.version}** ${r.summary} (${r.href})`).join('\n');
  const questions = FAQ.map((q) => `### ${q.question}\n\n${q.answer}`).join('\n\n');
  return `# Woolly, memory that knows when it's wrong

One local memory for Claude Code, Cursor and Codex. Every fact points at its code, and Woolly flags the ones that go stale.

[Get started](${siteUrl}/docs/quickstart) · [Download](${siteUrl}/download) · [GitHub](${githubUrl}) · [Docs for agents](${siteUrl}/llms.txt)

## Setup prompt

Paste this to your coding agent.

\`\`\`
${SETUP_PROMPT}
\`\`\`

## Why Woolly

You switch between Claude Code, Cursor and Codex on the same repo, and each starts from zero. The tools that do remember keep repeating facts the code stopped backing up. Woolly is plain Markdown on your machine, shared by every agent through MCP and hooks, with no cloud and no account. Every fact points at the code it's about, and when that code changes, Woolly marks the fact stale instead of letting your agent follow it (${siteUrl}/docs/stale-facts).

## What it does

- **It knows when a fact goes stale.** On every commit and on \`woolly check\`, each fact is checked against the code it points at. A renamed function, deleted file or changed value flags the fact with the commit that broke it.
- **Every fact cites its source.** A file, function or line range, the commit it was saved at, and the session it came from.
- **One memory for every agent on the repo.** Claude Code, Cursor and Codex share \`.woolly/woolly.md\`, and so do your teammates through git.
- **Sessions that start warm.** A short brief with the project's current facts, where you left off, and a one-line warning when something went stale.
- **Decisions stick.** When a session ends, Woolly saves what you chose, why, and what you ruled out.
- **Facts keep their history.** Nothing is deleted. \`woolly review\` confirms, fixes or retires stale facts.
- **It dreams.** Your local model merges duplicates, settles facts that changed and turns sessions into lessons.
- **Memory you can read.** Every memory is a Markdown file you can open, grep and edit.

## By the numbers

- 19 ms median recall time, on a laptop
- LOCOMO, answer in the top 20 results for 83.2% of 382 questions (95.6% for dates and time, 83.6% multi-hop), searched locally with nomic-embed-text
- No servers to trust
- MIT licensed

How we measure (${siteUrl}/docs/benchmark)

## Runs anywhere

- On your laptop, Mac, Windows or Linux (${siteUrl}/download)
- On a Raspberry Pi, synced over your own network and encrypted (${siteUrl}/docs/sync)
- In your repo, as project memory in git (${siteUrl}/docs/project-memory)

## Memory for the agents you build, too

The same engine has an API shaped like mem0's (${siteUrl}/docs/memory-api).

\`\`\`ts
import { Woolly } from 'woolly-ai';
const w = new Woolly();
await w.add('I just moved to Austin', { userId: 'alice' });
const memories = await w.search('where does she live?', { userId: 'alice' });
\`\`\`

## Private by design

No account, no cloud, no telemetry. Passwords and API keys are refused, and lines from a repo are checked for secrets and injected instructions.

## Shipping weekly

${shipped}

## Questions

${questions}
`;
}

export function AgentView({ releases }: { releases: Release[] }) {
  const text = homeMarkdown(releases);
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard can be blocked; the text is right there to select.
    }
  };
  return (
    <article className="pt-12 pb-20">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-fd-border pb-4">
        <Label>This page, as Markdown for your agent</Label>
        <button
          type="button"
          onClick={copy}
          className="inline-flex h-9 items-center gap-2 px-3 font-mono text-[12px] uppercase tracking-[0.12em] shadow-[inset_0_0_0_1px_var(--color-fd-border)] hover:bg-fd-muted"
        >
          {copied ? <Check aria-hidden="true" className="size-3.5 text-green-600" /> : <Copy aria-hidden="true" className="size-3.5" />}
          {copied ? 'Copied' : 'Copy Markdown'}
        </button>
      </div>
      <pre className="mt-6 font-mono text-[13px] leading-[1.75] whitespace-pre-wrap break-words text-fd-foreground/90">{text}</pre>
    </article>
  );
}
