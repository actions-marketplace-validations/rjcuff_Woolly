import { AnimatePresence, motion, useReducedMotion } from 'motion/react';
import { Plus } from 'lucide-react';
import { useId, useState } from 'react';
import { FAQ, type Faq as FaqItem } from '@/lib/faq';
import { cn } from '@/lib/utils';
import { ArrowLink, H2, LEAD, WRAP } from './parts';

// A quick start that settles gently, the same curve for opening and closing.
const EASE = [0.32, 0.72, 0, 1] as const;

/** One question. The answer grows open under it and folds away again. */
function Item({ question, answer, open, onToggle }: { question: string; answer: string; open: boolean; onToggle: () => void }) {
  const id = useId();
  const reduce = useReducedMotion();
  return (
    <li className={cn('rounded-2xl bg-[var(--tint-blue)] transition-colors duration-200', open && 'bg-[#dfe5ff]')}>
      <h3>
        <button
          type="button"
          id={`${id}-q`}
          aria-expanded={open}
          aria-controls={`${id}-a`}
          onClick={onToggle}
          className="no-press flex min-h-14 w-full items-center justify-between gap-4 rounded-2xl px-5 py-4 text-left text-[16.5px] font-semibold text-[var(--ink)] outline-none transition-colors duration-150 hover:bg-[#dce2ff] focus-visible:ring-2 focus-visible:ring-[var(--accent)] sm:px-6"
        >
          {question}
          <motion.span
            aria-hidden="true"
            animate={{ rotate: open ? 45 : 0 }}
            transition={reduce ? { duration: 0 } : { duration: 0.3, ease: EASE }}
            className="flex size-8 shrink-0 items-center justify-center rounded-full bg-white/70 text-[var(--ink)]"
          >
            <Plus className="size-4" />
          </motion.span>
        </button>
      </h3>
      <AnimatePresence initial={false}>
        {open ? (
          <motion.div
            key="answer"
            id={`${id}-a`}
            role="region"
            aria-labelledby={`${id}-q`}
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={reduce ? { duration: 0 } : { height: { duration: 0.35, ease: EASE }, opacity: { duration: 0.2, ease: 'easeOut' } }}
            className="overflow-hidden"
          >
            <motion.p
              initial={reduce ? false : { y: -6 }}
              animate={{ y: 0 }}
              exit={reduce ? undefined : { y: -6 }}
              transition={{ duration: 0.35, ease: EASE }}
              className="px-5 pb-5 text-[15.5px] leading-[1.65] text-pretty text-[var(--body)] sm:px-6"
            >
              {answer}
            </motion.p>
          </motion.div>
        ) : null}
      </AnimatePresence>
    </li>
  );
}

/** Common questions. One answer is open at a time. */
export function Faq({
  items = FAQ,
  lead = 'Anything else is in the docs, or ask on GitHub.',
  link = { label: 'Read the docs', href: '/docs' },
}: {
  items?: FaqItem[];
  lead?: string;
  link?: { label: string; href: string };
}) {
  const [open, setOpen] = useState<number | null>(0);
  return (
    <section aria-label="Questions" className="bg-white py-20 sm:py-28">
      <div className={cn(WRAP, 'grid gap-10 lg:grid-cols-[minmax(0,340px)_minmax(0,1fr)] lg:gap-16')}>
        <div className="flex flex-col items-start">
          <h2 className={H2}>Questions</h2>
          <p className={cn(LEAD, 'mt-5 max-w-[340px]')}>{lead}</p>
          <ArrowLink href={link.href} className="mt-4">
            {link.label}
          </ArrowLink>
        </div>
        <ul className="flex flex-col gap-3">
          {items.map((f, i) => (
            <Item key={f.question} question={f.question} answer={f.answer} open={open === i} onToggle={() => setOpen(open === i ? null : i)} />
          ))}
        </ul>
      </div>
    </section>
  );
}
