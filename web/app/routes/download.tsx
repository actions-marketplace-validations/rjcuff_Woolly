import type { Route } from './+types/download';
import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { CopyMorph, Pill } from '@/components/frost/home-content';
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
    image: '/og/docs/image.png',
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
    <DocShell stars={loaderData?.stars}>
      <span className="inline-block rounded-full border border-[var(--blue)] px-2.5 py-0.5 text-[12px] leading-[1.6]">Free and open source</span>
      <h1 className="mt-6 text-[44px] font-light leading-[1.02] tracking-[-0.035em] sm:text-[60px]">Download Woolly</h1>
      <p className="mt-6 max-w-[520px] text-[17px] leading-[1.6]">For Mac, Windows and Linux. Pick your computer.</p>

      <div role="tablist" aria-label="Operating system" className="mt-10 flex flex-wrap gap-2">
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
              'rounded-full border px-4 py-1.5 text-[15px] outline-none transition-colors duration-150 focus-visible:ring-2 focus-visible:ring-[var(--blue)] focus-visible:ring-offset-2',
              os === o.id ? 'border-[var(--blue)] bg-[var(--blue)] text-[var(--paper)]' : 'border-[var(--blue)]/40 hover:border-[var(--blue)]',
            )}
          >
            {o.label}
          </button>
        ))}
      </div>

      <div id="os-panel" role="tabpanel" aria-labelledby={`os-${os}`} className="mt-10 flex flex-col gap-10 border-t border-[var(--blue)]/20 pt-10">
        {os !== 'linux' && (
          <div className="flex flex-col items-start gap-3">
            <Pill href={os === 'mac' ? macApp : windowsApp} solid>
              {os === 'mac' ? 'Download for Mac' : 'Download for Windows'}
            </Pill>
            <p className="text-[14px] opacity-80">{os === 'mac' ? 'Apple Silicon and Intel, macOS 11 or later.' : 'Windows 10 or later.'}</p>
          </div>
        )}

        <div className="flex flex-col gap-3">
          <p className="text-[15px]">
            {os === 'mac' ? 'Or install the command line tool.' : os === 'windows' ? 'Or install the command line tool in PowerShell.' : 'Paste this in a terminal.'}
          </p>
          <div className="flex items-stretch rounded-[14px] border border-[var(--blue)]">
            <code className="flex min-w-0 flex-1 items-center overflow-x-auto whitespace-nowrap px-5 py-3.5 font-mono text-[13.5px] sm:text-[14px]">
              <span className="mr-3 opacity-50">$</span>
              {COMMANDS[os]}
            </code>
            <button
              type="button"
              onClick={copy}
              aria-label={copied ? 'Copied' : 'Copy install command'}
              className="no-press flex w-12 shrink-0 items-center justify-center rounded-r-[13px] border-l border-[var(--blue)] outline-none transition-colors duration-150 hover:bg-[var(--blue)] hover:text-[var(--paper)] focus-visible:bg-[var(--blue)] focus-visible:text-[var(--paper)]"
            >
              <CopyMorph done={copied} />
            </button>
          </div>
          {os === 'linux' && <p className="text-[14px] opacity-80">Works on x86 and ARM, including Raspberry Pi.</p>}
        </div>
      </div>

      <p className="mt-14 text-[17px] leading-[1.6]">
        Then connect your AI apps with <code className="whitespace-nowrap font-mono text-[0.9em]">woolly connect</code>, and run{' '}
        <code className="whitespace-nowrap font-mono text-[0.9em]">woolly init --hooks</code> in your repo. See the{' '}
        <Link to="/docs/quickstart" className="link-soft text-[var(--color-brand)]">
          quickstart
        </Link>
        .
      </p>
    </DocShell>
  );
}
