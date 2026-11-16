import type { Route } from './+types/download';
import { useEffect, useState } from 'react';
import { BlockButton, Label } from '@/components/doc/parts';
import { DocShell } from '@/components/doc/shell';
import { SegmentedTabs } from '@/components/ui';
import { CrossfadeLabel } from '@/components/easeui/swap-label';
import { detectOS, macApp, windowsApp, type OS } from '@/lib/download';
import { getStars } from '@/lib/github';
import { seo } from '@/lib/seo';

export function meta(_: Route.MetaArgs) {
  return seo({
    title: 'Download Woolly',
    description: 'Download Woolly for Mac, Windows and Linux. One local memory for all your AI tools.',
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
      <section className="pb-20 pt-14">
        <Label>Download</Label>
        <h1 className="mt-4 text-[40px] font-normal leading-[1.06] tracking-[-0.035em] sm:text-[56px]">Download Woolly</h1>
        <p className="mt-6 max-w-[520px] text-pretty text-[18px] leading-relaxed text-fd-muted-foreground">
          Free and open source, for Mac, Windows and Linux. Pick your computer.
        </p>

        <div className="mt-9">
          <SegmentedTabs options={OSES} active={os} onChange={setOS} label="Operating system" idPrefix="os" panelId="os-panel" />
        </div>

        <div id="os-panel" role="tabpanel" aria-labelledby={`os-${os}`} className="mt-8 flex flex-col gap-8">
          {os === 'mac' && (
            <div className="flex flex-col items-stretch gap-3 sm:items-start">
              <BlockButton href={macApp}>Download for Mac</BlockButton>
              <p className="text-[14px] text-fd-muted-foreground">Apple Silicon and Intel · macOS 11 or later</p>
            </div>
          )}

          {os === 'windows' && (
            <div className="flex flex-col items-stretch gap-3 sm:items-start">
              <BlockButton href={windowsApp}>Download for Windows</BlockButton>
              <p className="text-[14px] text-fd-muted-foreground">Windows 10 or later</p>
            </div>
          )}

          <div className="flex flex-col gap-3">
            <p className="text-[15px] text-fd-muted-foreground">
              {os === 'mac'
                ? 'Or install the command line tool:'
                : os === 'windows'
                  ? 'Or install the command line tool in PowerShell:'
                  : 'Paste this in a terminal:'}
            </p>
            <div className="flex items-center gap-2 bg-fd-muted py-2 pr-2 pl-4 shadow-[inset_0_0_0_1px_var(--color-fd-border)]">
              <code className="min-w-0 flex-1 overflow-x-auto whitespace-nowrap font-mono text-[14px]">{COMMANDS[os]}</code>
              <button
                type="button"
                onClick={copy}
                className="shrink-0 bg-fd-background px-3 py-1.5 font-mono text-[12px] uppercase tracking-[0.12em] text-fd-muted-foreground shadow-[0_0_0_1px_var(--color-fd-border)] transition-colors duration-150 hover:text-fd-foreground"
              >
                <CrossfadeLabel showing={copied ? 'copied' : 'copy'} labels={{ copy: 'Copy', copied: 'Copied' }} />
              </button>
            </div>
            {os === 'linux' && <p className="text-[14px] text-fd-muted-foreground">Works on x86 and ARM, including Raspberry Pi.</p>}
          </div>
        </div>

        <p className="mt-12 border-t border-fd-border pt-8 text-[15px] text-fd-muted-foreground">
          Then connect your AI apps with <code className="font-mono text-fd-foreground">woolly connect</code>. See the{' '}
          <a href="/docs/quickstart" className="text-brand underline-offset-4 hover:underline">
            quickstart
          </a>
          .
        </p>
      </section>
    </DocShell>
  );
}
