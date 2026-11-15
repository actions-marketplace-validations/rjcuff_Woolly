import { useState, type ReactNode } from 'react';
import { SegmentedTabs } from '@/components/ui';
import { CrossfadeLabel } from '@/components/easeui/swap-label';

type Tab = {
  id: string;
  label: string;
  title: string;
  /** What the copy button copies. */
  copy: string;
  code: string;
};

const TABS: Tab[] = [
  {
    id: 'you',
    label: 'Your AI tools',
    title: 'Terminal',
    copy: 'curl -fsSL https://usewoolly.com/install.sh | sh\nwoolly connect',
    code: `# 1. Install Woolly. On Windows: irm https://usewoolly.com/install.ps1 | iex
curl -fsSL https://usewoolly.com/install.sh | sh

# 2. Add it to Claude Code, Cursor, VS Code, Windsurf, Codex and the rest
woolly connect

# 3. Keep it running in the background, starting at login
woolly service install

# Now any connected tool can save and search the same memory.
# Tell Claude Code something, then ask Cursor:
woolly remember "We deploy the API to Fly.io from the main branch"
woolly recall "where do we deploy"

# See what each app saved, read and deleted
woolly log`,
  },
  {
    id: 'team',
    label: 'Your team',
    title: '.woolly/woolly.md',
    copy: 'woolly init\ngit add .woolly .gitattributes && git commit -m "Add project memory"',
    code: `# Run once in your repo, then commit it
woolly init
git add .woolly .gitattributes && git commit -m "Add project memory"

# Every teammate's agent reads this file and keeps it up to date:

## Stack
- The API is written in Go, the web app in React

## Decisions
- Use SQLite instead of Postgres. Why: one file, no server.
  Rejected: Postgres: too much setup.

## Gotchas
- Never run the billing tests against production

## Commands
- Start everything with \`make dev\``,
  },
  {
    id: 'app',
    label: 'Your app',
    title: 'reply.ts',
    copy: 'npm install woolly-ai',
    code: `// npm install woolly-ai ai @ai-sdk/openai   (or: pip install woolly-ai)
import { Woolly } from 'woolly-ai';
import { woollyMemory } from 'woolly-ai/ai-sdk';
import { generateText, wrapLanguageModel } from 'ai';
import { openai } from '@ai-sdk/openai';

const woolly = new Woolly({ app: 'My app' });

// Every call starts with the memories that matter for this message
const model = wrapLanguageModel({
  model: openai('gpt-5'),
  middleware: woollyMemory({ client: woolly }),
});

export async function reply(message: string) {
  const { text } = await generateText({ model, prompt: message });

  // Keep the facts worth remembering. Newer ones replace older ones.
  await woolly.learn([
    { role: 'user', content: message },
    { role: 'assistant', content: text },
  ]);

  return text;
}`,
  },
];

const KEYWORDS = new Set(['import', 'from', 'const', 'await', 'export', 'async', 'function', 'return']);
// Shell commands get their first word colored.
const COMMAND = /^(curl|woolly|git)(?= )/;

function tokens(text: string): ReactNode[] {
  return text.split(/('[^']*'|"[^"]*"|`[^`]*`|\b[a-z]+\b)/g).map((part, i) => {
    if (/^['"`]/.test(part)) {
      return (
        <span key={i} className="text-[#b45309] dark:text-[#fbbf24]">
          {part}
        </span>
      );
    }
    if (KEYWORDS.has(part)) {
      return (
        <span key={i} className="text-brand">
          {part}
        </span>
      );
    }
    return part;
  });
}

// A small highlighter: comments muted, strings and keywords colored.
function highlight(line: string): ReactNode {
  const trimmed = line.trimStart();
  if (trimmed.startsWith('## ')) return <span className="font-semibold">{line}</span>;
  if (trimmed.startsWith('#') || trimmed.startsWith('//')) {
    return <span className="text-fd-muted-foreground">{line}</span>;
  }
  const command = line.match(COMMAND);
  if (command) {
    return (
      <>
        <span className="text-brand">{command[1]}</span>
        {tokens(line.slice(command[1].length))}
      </>
    );
  }
  return tokens(line);
}

export function SetupTabs() {
  const [active, setActive] = useState(TABS[0].id);
  const [copied, setCopied] = useState(false);
  const tab = TABS.find((t) => t.id === active) ?? TABS[0];
  const lines = tab.code.split('\n');

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(tab.copy);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard can be blocked; the text is still there to select.
    }
  };

  return (
    <div className="flex w-full flex-col items-center gap-6">
      <SegmentedTabs
        options={TABS}
        active={active}
        onChange={(id) => {
          setActive(id);
          setCopied(false);
        }}
        label="Set up Woolly for"
        idPrefix="tab"
        panelId="setup-panel"
      />

      <div
        id="setup-panel"
        role="tabpanel"
        aria-labelledby={`tab-${tab.id}`}
        className="w-full overflow-hidden rounded-xl bg-fd-background text-left shadow-[0_0_0_1px_var(--color-fd-border),0_24px_48px_-24px_rgb(0_0_0/0.18)]"
      >
        <div className="flex items-center justify-between gap-4 border-b border-fd-border px-4 py-3">
          <div className="flex items-center gap-3">
            <span className="flex gap-1.5" aria-hidden="true">
              <span className="size-3 rounded-full bg-[#ff5f57]" />
              <span className="size-3 rounded-full bg-[#febc2e]" />
              <span className="size-3 rounded-full bg-[#28c840]" />
            </span>
            <span className="font-mono text-[13px] text-fd-muted-foreground">{tab.title}</span>
          </div>
          <button
            onClick={copy}
            className="rounded-md px-2.5 py-1 text-[13px] font-medium text-fd-muted-foreground transition-colors hover:bg-fd-accent hover:text-fd-foreground"
          >
            <CrossfadeLabel showing={copied ? 'copied' : 'copy'} labels={{ copy: 'Copy', copied: 'Copied' }} />
          </button>
        </div>
        <pre className="overflow-x-auto py-5 font-mono text-[13px] leading-[1.75] sm:text-[14px]">
          <code className="grid">
            {lines.map((line, i) => (
              <span key={`${tab.id}-${i}`} className="flex min-h-[1lh]">
                <span aria-hidden="true" className="w-12 shrink-0 select-none pr-4 text-right text-fd-muted-foreground/60 tabular-nums">
                  {i + 1}
                </span>
                <span className="pr-5 whitespace-pre">{highlight(line)}</span>
              </span>
            ))}
          </code>
        </pre>
      </div>
    </div>
  );
}
