import type { KeyboardEvent, ReactNode } from 'react';
import { indicatorClass, useSlidingIndicator } from '@/components/sliding-tabs';

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
