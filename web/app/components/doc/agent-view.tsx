import { Check, Copy } from 'lucide-react';
import { useState } from 'react';
import { SETUP_PROMPT } from '@/components/doc/home-article';
import { Label } from '@/components/doc/parts';
import { FAQ } from '@/lib/faq';
import type { Release } from '@/lib/releases';
import { githubUrl, siteUrl } from '@/lib/shared';

/** The homepage as Markdown, for an agent to read or a person to paste into one. */
export function homeMarkdown(releases: Release[]): string {
  const shipped = releases.map((r) => `- v${r.version}: ${r.summary} (${r.href})`).join('\n');
  const questions = FAQ.map((q) => `### ${q.question}\n\n${q.answer}`).join('\n\n');
  return `# Woolly: your agents share one brain

The open, local memory engine for agents. Your coding tools share one memory, and the agents you build get an API that keeps facts current and learns from every session, all on your laptop in Markdown you can read.

[Get started](${siteUrl}/docs/quickstart) · [Download](${siteUrl}/download) · [GitHub](${githubUrl}) · [Docs for agents](${siteUrl}/llms.txt)

## Setup prompt

Paste this to your coding agent:

\`\`\`
${SETUP_PROMPT}
\`\`\`

## Why Woolly

Every AI coding tool starts from zero. Each keeps its own memory, or none. Woolly is the memory they share: one folder on your laptop, plain Markdown, no cloud and no account. Any MCP tool can save to it and search it, and \`woolly proxy\` gives the same memory to Ollama and any app that uses the OpenAI or Anthropic API.

## Build agents that remember

The memory API uses mem0's routes and field names (${siteUrl}/docs/memory-api).

\`\`\`ts
import { Woolly } from 'woolly-ai';
const w = new Woolly();
await w.add('I just moved to Austin', { userId: 'alice' });
const memories = await w.search('where does she live?', { userId: 'alice' });
\`\`\`

## What it does

- **Sessions that start warm.** Every session opens with your profile, the project's notes and a handoff from last time.
- **Decisions stick.** When a session ends, Woolly saves what you chose, why, and what you ruled out.
- **Agents that remember.** An API shaped like mem0's: a local model decides whether each fact is new, an update or no longer true, per user.
- **It dreams.** While you're away, Woolly merges duplicates, settles facts that changed and turns sessions into lessons.
- **Facts keep their history.** A newer fact replaces the old one, which is kept with when it stopped being true. Your version beats an AI's guess.
- **Memory you can read.** Every memory is a Markdown file you can open, grep and edit.
- **One project memory for the team.** \`woolly init\` keeps \`.woolly/woolly.md\` in the repo for every teammate's agent.

## By the numbers

- Recall time: 23 ms median, on a laptop
- LOCOMO: found the message with the answer 86.7%, answered right 56.7% (qwen2.5:7b, local)
- Servers to trust: 0
- License: MIT

Method: ${siteUrl}/docs/benchmark

## Runs anywhere

- On your laptop: Mac, Windows and Linux (${siteUrl}/download)
- On a Raspberry Pi: sync over your own network, encrypted (${siteUrl}/docs/sync)
- In your repo: project memory in git (${siteUrl}/docs/project-memory)

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
