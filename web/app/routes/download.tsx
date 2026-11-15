import type { Route } from './+types/download';
import { useEffect, useState } from 'react';
import { SiteFooter, SiteHeader } from '@/components/site-chrome';
import { PrimaryButton, SegmentedTabs } from '@/components/ui';
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
    <div className="relative isolate flex min-h-svh flex-col overflow-x-clip px-[max(20px,env(safe-area-inset-left))]">
      <SiteHeader stars={loaderData?.stars} />

      <main className="flex flex-1 items-center justify-center py-12">
        <section className="flex w-full max-w-[560px] flex-col items-center gap-8 text-center">
          <h1 className="text-[32px] font-semibold tracking-[-0.03em] sm:text-[48px]">Download Woolly</h1>

          <SegmentedTabs
            options={OSES}
            active={os}
            onChange={setOS}
            label="Operating system"
            idPrefix="os"
            panelId="os-panel"
          />

          <div
            id="os-panel"
            role="tabpanel"
            aria-labelledby={`os-${os}`}
            className="flex w-full flex-col items-center gap-8"
          >
            {os === 'mac' && (
              <div className="flex flex-col items-center gap-3">
                <PrimaryButton href={macApp}>Download for Mac</PrimaryButton>
                <p className="text-[14px] text-fd-muted-foreground">Apple Silicon and Intel · macOS 11 or later</p>
              </div>
            )}

            {os === 'windows' && (
              <div className="flex flex-col items-center gap-3">
                <PrimaryButton href={windowsApp}>Download for Windows</PrimaryButton>
                <p className="text-[14px] text-fd-muted-foreground">Windows 10 or later</p>
              </div>
            )}

            <div className="flex w-full flex-col gap-3">
              <p className="text-[15px] text-fd-muted-foreground">
                {os === 'mac'
                  ? 'Or install the command line tool:'
                  : os === 'windows'
                    ? 'Or install the command line tool in PowerShell:'
                    : 'Paste this in a terminal:'}
              </p>
              <div className="flex items-center gap-2 rounded-lg bg-fd-background py-2 pr-2 pl-4 text-left shadow-[0_0_0_1px_var(--color-fd-border)]">
                <code className="min-w-0 flex-1 overflow-x-auto whitespace-nowrap font-mono text-[14px]">
                  {COMMANDS[os]}
                </code>
                <button
                  onClick={copy}
                  className="shrink-0 rounded-md px-3 py-1.5 text-[14px] font-medium text-fd-muted-foreground transition-colors hover:bg-fd-accent hover:text-fd-foreground"
                >
                  <CrossfadeLabel showing={copied ? 'copied' : 'copy'} labels={{ copy: 'Copy', copied: 'Copied' }} />
                </button>
              </div>
              {os === 'linux' && (
                <p className="text-[14px] text-fd-muted-foreground">Works on x86 and ARM, including Raspberry Pi.</p>
              )}
            </div>
          </div>

          <p className="text-[14px] text-fd-muted-foreground">
            Then connect your AI apps with <code className="font-mono">woolly connect</code>. See the{' '}
            <a href="/docs/quickstart" className="font-medium text-brand hover:underline">
              quickstart
            </a>
            .
          </p>
        </section>
      </main>

      <SiteFooter />
    </div>
  );
}
