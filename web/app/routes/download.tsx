import type { Route } from './+types/download';
import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { CopyMorph } from '@/components/frost/home-content';
import { Button } from '@/components/sky/parts';
import { InstallButtons } from '@/components/install-buttons';
import { cn } from '@/lib/utils';
import { DocShell } from '@/components/doc/shell';
import { detectOS, macApp, windowsApp, type OS } from '@/lib/download';
import { getStars } from '@/lib/github';
import { seo } from '@/lib/seo';

export function meta(_: Route.MetaArgs) {
  return seo({
    title: 'Download Woolly',
    description: 'Download Woolly for Mac, Windows and Linux. Memory that knows when it is wrong, for Claude Code, Cursor and Codex.',
    path: '/download',
    image: '/og/home.png',
  });
}

export async function loader() {
  return { stars: await getStars() };
}

const OSES: { id: OS; label: string }[] = [
  { id: 'mac', label: 'macOS' },
  { id: 'linux', label: 'Linux' },
  { id: 'windows', label: 'Windows' },
];

const COMMANDS: Record<OS, string> = {
  mac: 'curl -fsSL https://usewoolly.com/install.sh | sh',
  linux: 'curl -fsSL https://usewoolly.com/install.sh | sh',
  windows: 'irm https://usewoolly.com/install.ps1 | iex',
};

export default function Download({ loaderData }: Route.ComponentProps) {
  const [os, setOS] = useState<OS>('mac');
  const [copied, setCopied] = useState(false);

  useEffect(() => setOS(detectOS()), []);
  useEffect(() => {
    if (!copied) return;
    const t = window.setTimeout(() => setCopied(false), 1600);
    return () => window.clearTimeout(t);
  }, [copied]);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(COMMANDS[os]);
      setCopied(true);
    } catch {
      // Clipboard can be blocked; the command is still there to select.
    }
  };

  return (
    <DocShell stars={loaderData?.stars} eyebrow="Free and open source" title="Download Woolly" lead="For Mac, Windows and Linux. Pick your computer.">
      <div role="tablist" aria-label="Operating system" className="inline-flex flex-wrap gap-1 rounded-full bg-[var(--tint-blue)] p-1">
        {OSES.map((o) => (
          <button
            key={o.id}
            id={`os-${o.id}`}
            type="button"
            role="tab"
            aria-selected={os === o.id}
            aria-controls="os-panel"
            onClick={() => setOS(o.id)}
            className={cn(
              'no-press min-h-11 rounded-full px-5 text-[15px] font-semibold outline-none transition-colors duration-150 focus-visible:ring-2 focus-visible:ring-[var(--accent)]',
              os === o.id ? 'bg-white text-[var(--ink)] shadow-[0_1px_2px_rgb(27_35_80/0.08)]' : 'text-[var(--muted)] hover:bg-white/60',
            )}
          >
            {o.label}
          </button>
        ))}
      </div>

      <div id="os-panel" role="tabpanel" aria-labelledby={`os-${os}`} className="mt-10 flex flex-col gap-10">
        {os !== 'linux' && (
          <div className="flex flex-col items-start gap-3 rounded-3xl bg-[var(--tint-blue)] p-7 sm:p-8">
            <h2 className="serif text-[26px] text-[var(--ink)]">The app</h2>
            <p className="text-[15.5px] text-[var(--body)]">{os === 'mac' ? 'Apple Silicon and Intel, macOS 11 or later.' : 'Windows 10 or later.'}</p>
            <Button href={os === 'mac' ? macApp : windowsApp} className="mt-2">
              {os === 'mac' ? 'Download for Mac' : 'Download for Windows'}
            </Button>
          </div>
        )}

        <div className="flex flex-col gap-3">
          <h2 className="serif text-[26px] text-[var(--ink)]">{os === 'linux' ? 'Install' : 'Or the command line tool'}</h2>
          <p className="text-[15.5px] text-[var(--body)]">
            {os === 'mac' ? 'Paste this in Terminal.' : os === 'windows' ? 'Paste this in PowerShell.' : 'Paste this in a terminal.'}
          </p>
          <div className="mt-1 flex items-stretch overflow-hidden rounded-2xl bg-white ring-1 ring-[var(--line)] shadow-[var(--lift)]">
            <code className="flex min-w-0 flex-1 items-center overflow-x-auto whitespace-nowrap px-5 py-4 font-mono text-[13.5px] text-[var(--ink)] sm:text-[14px]">
              <span className="mr-3 text-[var(--accent)]">$</span>
              {COMMANDS[os]}
            </code>
            <button
              type="button"
              onClick={copy}
              aria-label={copied ? 'Copied' : 'Copy install command'}
              className="no-press flex w-12 shrink-0 items-center justify-center text-[var(--muted)] outline-none transition-colors duration-150 hover:bg-[var(--tint-blue)] focus-visible:bg-[var(--tint-blue)]"
            >
              <CopyMorph done={copied} />
            </button>
          </div>
          {os === 'linux' && <p className="text-[14px] text-[var(--muted)]">Works on x86 and ARM, including Raspberry Pi.</p>}
        </div>
      </div>

      <p className="mt-14 text-[17px] leading-[1.6] text-[var(--body)]">
        Then connect your AI apps with <code className="rounded-md bg-[var(--tint-blue)] px-1.5 py-0.5 font-mono text-[0.88em] whitespace-nowrap text-[var(--ink)]">woolly connect</code>, and run{' '}
        <code className="rounded-md bg-[var(--tint-blue)] px-1.5 py-0.5 font-mono text-[0.88em] whitespace-nowrap text-[var(--ink)]">woolly init --hooks</code> in your repo. See the{' '}
        <Link to="/docs/quickstart" className="font-semibold text-[var(--accent)] hover:text-[var(--ink)]">
          quickstart
        </Link>
        .
      </p>

      <div className="mt-8 flex flex-col gap-3">
        <p className="text-[15.5px] text-[var(--body)]">Using Cursor or LM Studio? Once Woolly is installed, add it in one click. If the app then can't find woolly, run woolly connect instead, which uses the full path.</p>
        <InstallButtons />
      </div>
    </DocShell>
  );
}
