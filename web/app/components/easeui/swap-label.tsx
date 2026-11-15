import { motion, useReducedMotion } from "motion/react";
import { type ReactNode, useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";

const FADE = { duration: 0.22, ease: [0.23, 1, 0.32, 1] } as const;

/** One label, faded and softly blurred out when it is not the one showing. Never unmounted, so it keeps its size. */
function Face({ shown, children, className }: { shown: boolean; children: ReactNode; className?: string }) {
  const reduce = useReducedMotion();
  return (
    <motion.span
      aria-hidden={!shown}
      initial={false}
      animate={shown ? { opacity: 1, filter: "blur(0px)" } : { opacity: 0, filter: reduce ? "blur(0px)" : "blur(3px)" }}
      transition={reduce ? { duration: 0 } : FADE}
      className={cn("col-start-1 row-start-1 inline-flex items-center justify-center gap-2", className)}
    >
      {children}
    </motion.span>
  );
}

/**
 * Two labels stacked in one grid cell, so the button is always as wide as the longer one and never resizes
 * when they trade places.
 */
export function CrossfadeLabel({
  showing,
  labels,
  loading = false,
}: {
  showing: string;
  labels: Record<string, ReactNode>;
  loading?: boolean;
}) {
  return (
    <span className="relative inline-grid">
      {Object.entries(labels).map(([key, label]) => (
        <Face key={key} shown={!loading && key === showing}>
          {label}
        </Face>
      ))}
      <motion.span
        aria-hidden="true"
        initial={false}
        animate={{ opacity: loading ? 1 : 0 }}
        transition={FADE}
        className="pointer-events-none absolute inset-0 flex items-center justify-center"
      >
        <span className="h-3.5 w-3.5 animate-spin rounded-full border-[1.5px] border-current/30 border-t-current motion-reduce:animate-none" />
      </motion.span>
    </span>
  );
}

/** Two labels of different lengths that trade places inside a pill. */
export function GlideLabel({ showing, labels }: { showing: string; labels: Record<string, ReactNode> }) {
  const refs = useRef<Record<string, HTMLSpanElement | null>>({});
  const [widths, setWidths] = useState<Record<string, number>>({});
  const reduce = useReducedMotion();

  useEffect(() => {
    const observer = new ResizeObserver(() => {
      setWidths(
        Object.fromEntries(Object.entries(refs.current).map(([key, node]) => [key, node?.getBoundingClientRect().width ?? 0])),
      );
    });
    for (const node of Object.values(refs.current)) if (node) observer.observe(node);
    return () => observer.disconnect();
  }, []);

  const width = widths[showing];

  return (
    <motion.span
      initial={false}
      animate={width ? { width } : undefined}
      transition={reduce ? { duration: 0 } : { duration: 0.3, ease: [0.645, 0.045, 0.355, 1] }}
      className="relative inline-grid overflow-hidden"
    >
      {Object.entries(labels).map(([key, label]) => (
        <Face key={key} shown={key === showing} className="justify-start whitespace-nowrap">
          <span
            ref={(node) => {
              refs.current[key] = node;
            }}
            className="inline-flex items-center gap-2"
          >
            {label}
          </span>
        </Face>
      ))}
    </motion.span>
  );
}
