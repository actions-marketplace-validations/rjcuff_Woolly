import type { KeyboardEvent, ReactNode } from 'react';
import { Link } from 'react-router';
import { indicatorClass, useSlidingIndicator } from '@/components/sliding-tabs';

type ButtonProps = {
  /** Site paths use the router; anything else is a plain link. */
  href: string;
  children: ReactNode;
  size?: 'md' | 'lg';
  className?: string;
};

function SmartLink({ href, className, children }: { href: string; className: string; children: ReactNode }) {
  return href.startsWith('/') ? (
    <Link to={href} className={className}>
      {children}
    </Link>
  ) : (
    <a href={href} className={className}>
      {children}
    </a>
  );
}

/** The main call to action: Woolly blue, with a small arrow box. */
export function PrimaryButton({ href, children, size = 'lg', className = '' }: ButtonProps) {
  const sizes = size === 'lg' ? 'py-2.5 pr-2.5 pl-5 text-[16px]' : 'py-2 pr-2 pl-4 text-[15px]';
  const box = size === 'lg' ? 'size-7 text-[15px]' : 'size-6 text-[13px]';
  return (
    <SmartLink
      href={href}
      className={`btn-brand btn-press inline-flex items-center gap-2 rounded-md font-medium ${sizes} ${className}`}
    >
      {children}
      <span aria-hidden="true" className={`flex items-center justify-center rounded bg-black/20 ${box}`}>
        →
      </span>
    </SmartLink>
  );
}

/** The quieter button that sits next to a primary one. */
export function SecondaryButton({ href, children, size = 'lg', className = '' }: ButtonProps) {
  const sizes = size === 'lg' ? 'px-5 py-3 text-[16px]' : 'px-4 py-2 text-[15px]';
  return (
    <SmartLink
      href={href}
      className={`btn-press inline-flex items-center rounded-md bg-fd-background font-medium shadow-[0_0_0_1px_var(--color-fd-border)] transition-colors hover:bg-fd-accent ${sizes} ${className}`}
    >
      {children}
    </SmartLink>
  );
}

export type TabOption<T extends string> = { id: T; label: string };

/** Tabs in a bordered box, with a blue highlight that slides to the chosen one. */
export function SegmentedTabs<T extends string>({
  options,
  active,
  onChange,
  label,
  idPrefix,
  panelId,
}: {
  options: TabOption<T>[];
  active: T;
  onChange: (id: T) => void;
  label: string;
  idPrefix: string;
  panelId: string;
}) {
  const { register, style } = useSlidingIndicator(active);
  // Arrow keys, Home and End move between tabs, as a tab list should.
  const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    const i = options.findIndex((o) => o.id === active);
    const last = options.length - 1;
    const next =
      event.key === 'ArrowRight'
        ? i === last
          ? 0
          : i + 1
        : event.key === 'ArrowLeft'
          ? i === 0
            ? last
            : i - 1
          : event.key === 'Home'
            ? 0
            : event.key === 'End'
              ? last
              : -1;
    if (next < 0) return;
    event.preventDefault();
    onChange(options[next].id);
    document.getElementById(`${idPrefix}-${options[next].id}`)?.focus();
  };
  return (
    <div
      role="tablist"
      aria-label={label}
      onKeyDown={onKeyDown}
      // 10px outside, 4px inset, 6px on the tabs, so the corners stay concentric.
      className="relative flex max-w-full overflow-x-auto rounded-[10px] bg-fd-background p-1 shadow-[0_0_0_1px_var(--color-fd-border)]"
    >
      <span aria-hidden="true" className={`${indicatorClass} inset-y-1 rounded-md bg-brand/10`} style={style} />
      {options.map((o) => (
        <button
          key={o.id}
          ref={register(o.id)}
          role="tab"
          id={`${idPrefix}-${o.id}`}
          aria-selected={o.id === active}
          aria-controls={panelId}
          tabIndex={o.id === active ? 0 : -1}
          onClick={() => onChange(o.id)}
          className={`relative min-h-10 whitespace-nowrap rounded-md px-4 text-[15px] font-medium outline-none transition-colors duration-150 focus-visible:ring-2 focus-visible:ring-foreground/40 sm:px-5 ${
            o.id === active ? 'text-brand' : 'text-fd-muted-foreground hover:text-fd-foreground'
          }`}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}
