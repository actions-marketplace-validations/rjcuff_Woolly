import { motion, useReducedMotion } from 'motion/react';
import { ArrowRight, ChevronRight } from 'lucide-react';
import { useEffect, useState, type ReactNode } from 'react';
import { Link } from 'react-router';
import { CopyMorph } from '@/components/frost/home-content';
import { cn } from '@/lib/utils';

export const WRAP = 'mx-auto w-full max-w-[1120px] px-4 sm:px-8';

/** Section heading sizes, shared so every section reads the same. */
export const H2 = 'serif text-balance text-[30px]/[1.12] sm:text-[36px]/[1.1] md:text-[44px]/[1.06] text-[var(--ink)]';
export const LEAD = 'text-[17px] leading-[1.6] text-pretty text-[var(--body)]';

type Tone = 'action' | 'white' | 'ghost' | 'night';

const TONES: Record<Tone, string> = {
  // Buttons change only their background on hover.
  action: 'bg-[var(--action)] text-white shadow-[inset_0_-2px_6px_rgb(0_0_0/0.35),0_6px_14px_-6px_rgb(30_39_102/0.5)] hover:bg-[var(--action-hover)] focus-visible:ring-[var(--action)]',
  white: 'bg-white text-[var(--ink)] shadow-[inset_0_-3px_6px_-2px_#e5e8f4,0_1px_2px_rgb(27_35_80/0.08)] ring-1 ring-[var(--line)] hover:bg-[#f3f5fc] focus-visible:ring-[var(--accent)]',
  ghost: 'text-[var(--ink)] hover:bg-white/50 focus-visible:ring-[var(--accent)]',
  night: 'bg-white/5 text-white ring-1 ring-white/25 hover:bg-white/12 focus-visible:ring-white',
};

/** A pill button. Internal paths use the router, the rest are plain links. */
export function Button({ href, children, tone = 'action', className }: { href: string; children: ReactNode; tone?: Tone; className?: string }) {
  const cls = cn(
    'btn-press inline-flex h-11 items-center gap-1.5 rounded-full px-5 text-[15px] font-semibold outline-none transition-colors duration-150 focus-visible:ring-2 focus-visible:ring-offset-2',
    TONES[tone],
    className,
  );
  const inner = (
    <>
      {children}
      <ChevronRight className="nudge size-4" aria-hidden="true" />
    </>
  );
  return href.startsWith('/') ? (
    <Link to={href} className={cls}>
      {inner}
    </Link>
  ) : (
    <a href={href} className={cls}>
      {inner}
    </a>
  );
}

/** A violet text link with an arrow that steps right on hover. */
export function ArrowLink({ href, children, className }: { href: string; children: ReactNode; className?: string }) {
  const cls = cn('inline-flex min-h-11 items-center gap-1.5 text-[15px] font-semibold text-[var(--accent)] outline-none hover:text-[var(--ink)] focus-visible:underline', className);
  const inner = (
    <>
      {children}
      <ArrowRight className="nudge size-4" aria-hidden="true" />
    </>
  );
  return href.startsWith('/') ? (
    <Link to={href} className={cls}>
      {inner}
    </Link>
  ) : (
    <a href={href} className={cls}>
      {inner}
    </a>
  );
}

const INSTALL = {
  unix: 'curl -fsSL https://usewoolly.com/install.sh | sh',
  windows: 'irm https://usewoolly.com/install.ps1 | iex',
} as const;

type OS = keyof typeof INSTALL;

/** The one line install, on a white card, with a switch for Windows. */
export function Install({ className }: { className?: string }) {
  const [os, setOS] = useState<OS>('unix');
  const [copied, setCopied] = useState(false);
  const reduce = useReducedMotion();
  useEffect(() => {
    if (/Windows/i.test(navigator.userAgent)) setOS('windows');
  }, []);
  useEffect(() => {
    if (!copied) return;
    const t = window.setTimeout(() => setCopied(false), 1600);
    return () => window.clearTimeout(t);
  }, [copied]);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(INSTALL[os]);
      setCopied(true);
    } catch {
      // Clipboard can be blocked; the command is still there to select.
    }
  };
  return (
    <div className={cn('w-full max-w-[540px]', className)}>
      <div className="flex items-stretch overflow-hidden rounded-2xl bg-white/85 shadow-[var(--lift)] ring-1 ring-white backdrop-blur">
        <code className="flex min-w-0 flex-1 items-center overflow-x-auto whitespace-nowrap px-4 py-3.5 font-mono text-[13px] text-[var(--ink)] sm:text-[13.5px]">
          <span className="mr-3 text-[var(--accent)]">$</span>
          <motion.span key={os} initial={reduce ? false : { opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.18 }}>
            {INSTALL[os]}
          </motion.span>
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
      <div role="group" aria-label="Operating system" className="mt-2.5 flex gap-1 text-[13px]">
        {(
          [
            ['unix', 'macOS & Linux'],
            ['windows', 'Windows'],
          ] as const
        ).map(([id, label]) => (
          <button
            key={id}
            type="button"
            aria-pressed={os === id}
            onClick={() => setOS(id)}
            className={cn(
              'no-press rounded-full px-3 py-1.5 font-medium outline-none transition-colors duration-150 focus-visible:ring-2 focus-visible:ring-[var(--accent)]',
              os === id ? 'bg-white/70 text-[var(--ink)]' : 'text-[var(--muted)] hover:bg-white/40',
            )}
          >
            {label}
          </button>
        ))}
      </div>
    </div>
  );
}

/**
 * A text button that copies the install command for the visitor's system.
 * Only the icon changes: it turns into a check for a moment after copying.
 */
export function CopyInstall({ className }: { className?: string }) {
  const [os, setOS] = useState<OS>('unix');
  const [copied, setCopied] = useState(false);
  useEffect(() => {
    if (/Windows/i.test(navigator.userAgent)) setOS('windows');
  }, []);
  useEffect(() => {
    if (!copied) return;
    const t = window.setTimeout(() => setCopied(false), 1600);
    return () => window.clearTimeout(t);
  }, [copied]);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(INSTALL[os]);
      setCopied(true);
    } catch {
      // Clipboard can be blocked; the docs have the command too.
    }
  };
  return (
    <button
      type="button"
      onClick={copy}
      title={INSTALL[os]}
      className={cn(
        'no-press inline-flex min-h-11 items-center gap-2 rounded-lg px-1 text-[15px] font-semibold text-[var(--ink)] outline-none transition-colors duration-150 hover:text-[var(--accent)] focus-visible:ring-2 focus-visible:ring-[var(--accent)]',
        className,
      )}
    >
      <CopyMorph done={copied} />
      Copy install command
      <span className="sr-only" aria-live="polite">
        {copied ? 'Copied' : ''}
      </span>
    </button>
  );
}

/** A small rounded label above a heading. */
export function Tag({ children, className }: { children: ReactNode; className?: string }) {
  return <span className={cn('inline-flex self-start rounded-full bg-white/70 px-3 py-1 text-[12.5px] font-semibold text-[var(--accent)] ring-1 ring-[var(--line)]', className)}>{children}</span>;
}
