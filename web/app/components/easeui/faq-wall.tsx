import { CornerDownLeft, Plus, Search } from "lucide-react";
import { type ReactNode, useId, useMemo, useState } from "react";
import { cn } from "@/lib/utils";

export interface FaqItem {
  question: string;
  answer: ReactNode;
  /** Extra words this answer should match on, for the words people actually type. */
  keywords?: string[];
}

/** Words carried by every question, so matching on them is matching on nothing. */
const STOPWORDS = new Set([
  "a", "an", "and", "are", "as", "at", "be", "by", "can", "do", "does", "for", "from", "how", "i", "if", "in", "is",
  "it", "my", "of", "on", "or", "our", "the", "to", "we", "what", "when", "where", "who", "why", "will", "with", "you",
  "your",
]);

function terms(text: string): string[] {
  return text
    .toLowerCase()
    .replace(/[^a-z0-9\s]/g, " ")
    .split(/\s+/)
    .filter((word) => word.length > 1 && !STOPWORDS.has(word));
}

/** How well one entry answers what was typed. */
function score(query: string, item: FaqItem): number {
  const asked = terms(query);
  if (asked.length === 0) return 0;

  const inQuestion = terms(item.question);
  const inKeywords = (item.keywords ?? []).map((word) => word.toLowerCase());

  let total = 0;
  for (const word of asked) {
    if (inKeywords.some((keyword) => keyword.includes(word))) total += 3;
    else if (inQuestion.includes(word)) total += 2;
    else if (inQuestion.some((other) => other.startsWith(word) || word.startsWith(other))) total += 1;
  }
  return total / asked.length;
}

export interface FaqWallProps {
  title?: string;
  description?: string;
  items: FaqItem[];
  /** Indexes open on first render. Default none. */
  defaultOpen?: number[];
  /** Lets several answers stay open at once. Default false, which closes the last one. */
  multiple?: boolean;
  /** The card under the list, for the question nobody wrote down. */
  contact?: { title: string; description?: string; actionLabel: string; actionHref: string };
  /** Adds a row at the end of the list that answers a typed question from the list itself. */
  ask?: {
    label?: string;
    placeholder?: string;
    /** Searched instead of `items`, for a bank larger than the list on screen. */
    bank?: FaqItem[];
    /** Shown when nothing scores. Falls back to the contact card's own words. */
    noMatch?: ReactNode;
  };
  className?: string;
}

/**
 * The questions that always follow a price, as a list of disclosures with an optional way out at the bottom.
 */
export function FaqWall({
  title = "Questions",
  description,
  items,
  defaultOpen = [],
  multiple = false,
  contact,
  ask,
  className,
}: FaqWallProps) {
  const [open, setOpen] = useState<Set<number>>(() => new Set(defaultOpen));
  const [query, setQuery] = useState("");
  const id = useId();

  const bank = ask?.bank ?? items;

  // Every entry that scores at all, best first, so the near misses are still offered rather than thrown away.
  const matches = useMemo(() => {
    if (!ask || query.trim().length < 2) return [];
    return bank
      .map((item) => ({ item, value: score(query, item) }))
      .filter((entry) => entry.value > 0)
      .sort((a, b) => b.value - a.value)
      .slice(0, 3);
  }, [ask, bank, query]);

  const best = matches[0]?.item ?? null;

  const toggle = (index: number) =>
    setOpen((current) => {
      const next = multiple ? new Set(current) : new Set<number>();
      if (current.has(index)) next.delete(index);
      else next.add(index);
      return next;
    });

  return (
    <section className={cn("w-full px-4 py-16 sm:px-6 sm:py-20", className)}>
      <div className="mx-auto max-w-3xl">
        <div className="flex flex-col gap-3 text-center">
          <h2 className="text-balance text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">{title}</h2>
          {description ? (
            <p className="mx-auto max-w-xl text-pretty text-sm leading-6 text-muted-foreground sm:text-base">
              {description}
            </p>
          ) : null}
        </div>

        <ul className="mt-10 flex flex-col">
          {items.map((item, index) => {
            const isOpen = open.has(index);
            const panelId = `${id}-answer-${index}`;

            return (
              <li key={item.question} className="border-t border-border last:border-b">
                <h3>
                  <button
                    type="button"
                    onClick={() => toggle(index)}
                    aria-expanded={isOpen}
                    aria-controls={panelId}
                    className="flex w-full items-center gap-4 py-5 text-left outline-none focus-visible:ring-2 focus-visible:ring-foreground/40"
                  >
                    <span className="flex-1 text-pretty text-sm font-medium text-foreground sm:text-base">
                      {item.question}
                    </span>
                    {/* One mark that turns, rather than two swapped in and out. */}
                    <Plus
                      aria-hidden="true"
                      className={cn(
                        "h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200 ease-out motion-reduce:transition-none",
                        isOpen && "rotate-45",
                      )}
                    />
                  </button>
                </h3>

                <div
                  id={panelId}
                  className={cn(
                    "grid transition-[grid-template-rows] duration-200 ease-out motion-reduce:transition-none",
                    isOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
                  )}
                >
                  {/* min-h-0 is what lets the track reach 0fr: a grid item refuses to shrink under its own content. */}
                  <div className="min-h-0 overflow-hidden">
                    <div
                      inert={!isOpen}
                      className={cn(
                        "pb-5 pr-8 text-sm leading-6 text-muted-foreground transition-opacity duration-200 ease-out motion-reduce:transition-none",
                        isOpen ? "opacity-100" : "opacity-0",
                      )}
                    >
                      {item.answer}
                    </div>
                  </div>
                </div>
              </li>
            );
          })}
        </ul>

        {ask ? (
          <div className="mt-8 rounded-2xl bg-card p-5 shadow-[0_0_0_1px_var(--border)]">
            <label htmlFor={`${id}-ask`} className="text-sm font-medium text-foreground">
              {ask.label ?? "Ask your own"}
            </label>

            <div className="mt-3 flex min-h-11 items-center gap-2 rounded-xl px-3 shadow-[0_0_0_1px_var(--border)] focus-within:shadow-[0_0_0_1px_var(--border-strong)]">
              <Search aria-hidden="true" className="h-4 w-4 shrink-0 text-muted-foreground" />
              <input
                id={`${id}-ask`}
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder={ask.placeholder ?? "Type a question"}
                className="min-w-0 flex-1 bg-transparent py-2 text-sm text-foreground outline-none placeholder:text-muted-foreground/70"
              />
              <CornerDownLeft aria-hidden="true" className="h-3.5 w-3.5 shrink-0 text-muted-foreground/70" />
            </div>

            {/* Polite, so an answer arriving does not interrupt whatever is being read. */}
            <div aria-live="polite" className="mt-4 empty:mt-0">
              {query.trim().length >= 2 ? (
                best ? (
                  <div className="flex flex-col gap-2">
                    <p className="text-sm font-medium text-foreground">{best.question}</p>
                    <div className="text-sm leading-6 text-muted-foreground">{best.answer}</div>

                    {matches.length > 1 ? (
                      <div className="mt-2 flex flex-wrap gap-1.5">
                        {matches.slice(1).map((entry) => (
                          <button
                            key={entry.item.question}
                            type="button"
                            onClick={() => setQuery(entry.item.question)}
                            className="rounded-md px-2.5 py-1 text-xs text-muted-foreground shadow-[0_0_0_1px_var(--border)] outline-none transition-colors duration-150 hover:text-foreground focus-visible:ring-2 focus-visible:ring-foreground/40"
                          >
                            {entry.item.question}
                          </button>
                        ))}
                      </div>
                    ) : null}
                  </div>
                ) : (
                  <p className="text-sm leading-6 text-muted-foreground">
                    {ask.noMatch ?? "Nothing here covers that one yet."}
                  </p>
                )
              ) : null}
            </div>
          </div>
        ) : null}

        {contact ? (
          <div className="mt-6 flex flex-col items-start gap-4 rounded-2xl bg-card p-6 shadow-[0_0_0_1px_var(--border)] sm:flex-row sm:items-center sm:justify-between">
            <div>
              <p className="text-sm font-medium text-foreground">{contact.title}</p>
              {contact.description ? (
                <p className="mt-1 text-sm text-muted-foreground">{contact.description}</p>
              ) : null}
            </div>
            <a
              href={contact.actionHref}
              className="btn-press inline-flex min-h-10 shrink-0 items-center rounded-md px-4 text-sm font-medium text-foreground shadow-[0_0_0_1px_var(--border-strong)] outline-none transition-colors duration-150 hover:bg-foreground/5 focus-visible:ring-2 focus-visible:ring-foreground/40"
            >
              {contact.actionLabel}
            </a>
          </div>
        ) : null}
      </div>
    </section>
  );
}
