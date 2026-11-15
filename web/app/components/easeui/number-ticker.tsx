import { motion, useInView, useReducedMotion } from "motion/react";
import { useEffect, useMemo, useRef, useState } from "react";
import { EASE_OUT } from "@/lib/ease";
import { cn } from "@/lib/utils";

export interface NumberTickerProps {
  /** Where the count lands. */
  value: number;
  /** Where it starts. Default 0. */
  from?: number;
  /** Seconds for the whole count. Default 1.6. */
  duration?: number;
  /** Seconds to wait before starting, for staggering a row. Default 0. */
  delay?: number;
  decimals?: number;
  prefix?: string;
  suffix?: string;
  /** 12400 reads as 12.4K. */
  compact?: boolean;
  /** Counts only once it scrolls into view. Default true. */
  startOnView?: boolean;
  /** Rises and sharpens as it counts. Off leaves the figure still while the digits run. Default true. */
  entrance?: boolean;
  className?: string;
}

/** Fast out of the gate and a long settle, which reads as a number arriving rather than a progress bar filling. */
function easeOutExpo(t: number): number {
  return t === 1 ? 1 : 1 - 2 ** (-10 * t);
}

/** A number that counts up to its value, once, when it is first seen. */
export function NumberTicker({
  value,
  from = 0,
  duration = 1.6,
  delay = 0,
  decimals = 0,
  prefix,
  suffix,
  compact = false,
  startOnView = true,
  entrance = true,
  className,
}: NumberTickerProps) {
  const ref = useRef<HTMLSpanElement>(null);
  const reduce = useReducedMotion();
  const inView = useInView(ref, { once: true, amount: 0.4 });
  // The count is written straight to the DOM, so React does not re-render every frame.
  const text = useRef<HTMLSpanElement>(null);
  const [initial] = useState(() => (reduce ? value : from));

  const format = useMemo(
    () =>
      new Intl.NumberFormat("en-US", {
        minimumFractionDigits: decimals,
        maximumFractionDigits: decimals,
        ...(compact ? { notation: "compact" as const, maximumFractionDigits: 1 } : {}),
      }),
    [decimals, compact],
  );

  const started = startOnView ? inView : true;

  useEffect(() => {
    const write = (n: number) => {
      if (text.current) text.current.textContent = `${prefix ?? ""}${format.format(n)}${suffix ?? ""}`;
    };
    if (reduce) {
      write(value);
      return;
    }
    if (!started) return;

    let frame = 0;
    let startedAt: number | null = null;
    const span = (duration + delay) * 1000;
    const wait = delay * 1000;

    const tick = (now: number) => {
      startedAt ??= now;
      const elapsed = now - startedAt;

      if (elapsed < wait) {
        frame = requestAnimationFrame(tick);
        return;
      }

      const progress = Math.min((elapsed - wait) / (span - wait), 1);
      write(from + (value - from) * easeOutExpo(progress));
      if (progress < 1) frame = requestAnimationFrame(tick);
    };

    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [started, reduce, value, from, duration, delay, format, prefix, suffix]);

  const final = `${prefix ?? ""}${format.format(value)}${suffix ?? ""}`;

  const rise = entrance && !reduce;

  return (
    <span ref={ref} className={cn("inline-block tabular-nums", className)}>
      {/* The final value, read once. The counting text beside it is hidden, since a number read out forty times is noise. */}
      <span className="sr-only">{final}</span>
      <motion.span
        aria-hidden="true"
        className="inline-block"
        initial={rise ? { opacity: 0, y: "0.25em", filter: "blur(6px)" } : false}
        animate={
          started && rise ? { opacity: 1, y: "0em", filter: "blur(0px)" } : rise ? {} : { opacity: 1 }
        }
        transition={{ duration: 0.5, ease: EASE_OUT, delay }}
      >
        <span ref={text}>{`${prefix ?? ""}${format.format(initial)}${suffix ?? ""}`}</span>
      </motion.span>
    </span>
  );
}
