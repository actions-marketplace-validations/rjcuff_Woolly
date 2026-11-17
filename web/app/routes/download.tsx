import type { Route } from './+types/download';
import { useEffect, useState } from 'react';
import { UNDERLINE } from '@/components/doc/rail';
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

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(COMMANDS[os]);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard can be blocked; the command is still there to select.
    }
  };

  return (
    <DocShell stars={loaderData?.stars}>
      <section className="pb-10">
        <h1 className="text-[24px] font-normal leading-[1.35] tracking-[-0.01em] sm:text-[28px]">Download Woolly</h1>
        <p className="mt-6 max-w-[520px] text-pretty text-[15px] leading-[1.8] text-fd-muted-foreground">
          Free and open source, for Mac, Windows and Linux. Pick your computer.
        </p>

        <div role="tablist" aria-label="Operating system" className="mt-12 flex gap-8 text-[14px]">
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
                'outline-none transition-colors duration-200 focus-visible:text-fd-foreground',
                os === o.id ? 'text-fd-foreground underline decoration-1 underline-offset-[6px]' : 'text-fd-muted-foreground hover:text-fd-foreground',
              )}
            >
              {o.label}
            </button>
          ))}
        </div>

        <div id="os-panel" role="tabpanel" aria-labelledby={`os-${os}`} className="mt-10 flex flex-col gap-12">
          {os !== 'linux' && (
            <div className="flex flex-col gap-3">
              <a href={os === 'mac' ? macApp : windowsApp} className={cn(UNDERLINE, 'self-start text-[15px]')}>
                {os === 'mac' ? 'Download for Mac' : 'Download for Windows'}
              </a>
              <p className="text-[13px] text-fd-muted-foreground">
                {os === 'mac' ? 'Apple Silicon and Intel, macOS 11 or later.' : 'Windows 10 or later.'}
              </p>
            </div>
          )}

          <div className="flex flex-col gap-3">
            <div className="flex items-baseline justify-between gap-4">
              <p className="text-[13px] text-fd-muted-foreground">
                {os === 'mac'
                  ? 'Or install the command line tool.'
                  : os === 'windows'
                    ? 'Or install the command line tool in PowerShell.'
                    : 'Paste this in a terminal.'}
              </p>
              <button type="button" onClick={copy} className={cn(UNDERLINE, 'text-[12px] text-fd-muted-foreground')}>
                {copied ? 'Copied' : 'Copy'}
              </button>
            </div>
            <div className="bg-fd-muted px-5 py-4">
              <code className="block overflow-x-auto whitespace-nowrap text-[13px]">
                <span className="text-fd-muted-foreground">$ </span>
                {COMMANDS[os]}
              </code>
            </div>
            {os === 'linux' && <p className="text-[13px] text-fd-muted-foreground">Works on x86 and ARM, including Raspberry Pi.</p>}
          </div>
        </div>

        <p className="mt-16 text-[15px] leading-[1.8] text-fd-muted-foreground">
          Then connect your AI apps with <code className="whitespace-nowrap text-fd-foreground">woolly connect</code>, and run{' '}
          <code className="whitespace-nowrap text-fd-foreground">woolly init --hooks</code> in your repo. See the{' '}
          <a href="/docs/quickstart" className={cn(UNDERLINE, 'text-fd-foreground')}>
            quickstart
          </a>
          .
        </p>
      </section>
    </DocShell>
  );
}
