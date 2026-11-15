import { ChevronDown, Menu, X } from "lucide-react";
import { type ReactNode, useEffect, useRef, useState } from "react";
import { Link } from "react-router";
import { cn } from "@/lib/utils";

export interface NavItem {
  label: string;
  href: string;
  /** One line under the label inside a menu. Ignored on a plain link. */
  description?: string;
}

export interface NavLink {
  label: string;
  /** Where the link goes. Left off when `items` is set, since the label opens a menu instead. */
  href?: string;
  /** Turns the entry into a menu: a button with a chevron, and these inside it. */
  items?: NavItem[];
}

export interface SiteNavbarProps {
  /** Wordmark at the left. A string renders as text; pass a node for a real mark. */
  brand: { label: React.ReactNode; href: string };
  links: NavLink[];
  /** Quiet action, usually Sign in. */
  secondaryAction?: { label: string; href: string };
  /** The filled one. */
  primaryAction?: { label: string; href: string };
  /** Always on the bar, phone included, such as a theme switch. */
  extras?: ReactNode;
  /** Marks the current page, matched against `href`. */
  currentHref?: string;
  className?: string;
}

/** Site paths stay in the app; anything else is a plain link. */
function A({ href = "", ...props }: React.ComponentProps<"a">) {
  return href.startsWith("/") ? <Link to={href} {...props} /> : <a href={href} {...props} />;
}

/** The top of a page: a wordmark, the links, and up to two actions, collapsing to a panel on a phone. */
export function SiteNavbar({
  brand,
  links,
  secondaryAction,
  primaryAction,
  extras,
  currentHref,
  className,
}: SiteNavbarProps) {
  const [scrolled, setScrolled] = useState(false);
  const [open, setOpen] = useState(false);
  const [menu, setMenu] = useState<string | null>(null);
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set());
  const bar = useRef<HTMLElement>(null);

  // Passive, because this only ever reads scrollY and never cancels the scroll.
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  useEffect(() => {
    if (!open) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    // Locked while the panel is open, so the page behind it cannot scroll away.
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = previous;
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  useEffect(() => {
    if (!menu) return;
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") setMenu(null);
    };
    const onPointerDown = (event: PointerEvent) => {
      if (!bar.current?.contains(event.target as Node)) setMenu(null);
    };
    document.addEventListener("keydown", onKey);
    document.addEventListener("pointerdown", onPointerDown);
    return () => {
      document.removeEventListener("keydown", onKey);
      document.removeEventListener("pointerdown", onPointerDown);
    };
  }, [menu]);

  const linkClass = (href?: string) =>
    cn(
      "rounded-lg px-3 py-2 text-[15px] font-medium outline-none transition-colors duration-150 focus-visible:ring-2 focus-visible:ring-foreground/40",
      href && href === currentHref ? "text-foreground" : "text-muted-foreground hover:text-foreground",
    );

  return (
    <header
      ref={bar}
      className={cn(
        "sticky top-0 z-50 w-full transition-[background-color,box-shadow,backdrop-filter] duration-200",
        scrolled ? "bg-background/80 shadow-[0_1px_0_0_var(--border)] backdrop-blur-md" : "bg-transparent",
        className,
      )}
    >
      <nav aria-label="Main" className="mx-auto flex h-16 max-w-[1200px] items-center gap-3 px-[max(20px,env(safe-area-inset-left))]">
        <A
          href={brand.href}
          className="rounded-lg text-sm font-semibold tracking-tight text-foreground outline-none focus-visible:ring-2 focus-visible:ring-foreground/40"
        >
          {brand.label}
        </A>

        <ul className="ml-4 hidden items-center gap-1 md:flex">
          {links.map((link) => (
            <li key={link.label} className="relative">
              {link.items ? (
                <>
                  <button
                    type="button"
                    onClick={() => setMenu((current) => (current === link.label ? null : link.label))}
                    aria-expanded={menu === link.label}
                    className={cn(linkClass(), "inline-flex items-center gap-1", menu === link.label && "text-foreground")}
                  >
                    {link.label}
                    <ChevronDown
                      aria-hidden="true"
                      className={cn(
                        "h-3.5 w-3.5 transition-transform duration-200 ease-out motion-reduce:transition-none",
                        menu === link.label && "rotate-180",
                      )}
                    />
                  </button>

                  {menu === link.label ? (
                    <ul className="absolute left-0 top-full z-50 mt-1 w-64 rounded-xl bg-popover p-1 shadow-[0_0_0_1px_var(--border-strong),0_16px_40px_-16px_oklch(0%_0_0/0.4)]">
                      {link.items.map((item) => (
                        <li key={item.href}>
                          <A
                            href={item.href}
                            onClick={() => setMenu(null)}
                            className="flex flex-col gap-0.5 rounded-lg px-3 py-2 outline-none transition-colors duration-150 hover:bg-accent/10 focus-visible:ring-2 focus-visible:ring-foreground/40"
                          >
                            <span className="text-sm text-foreground">{item.label}</span>
                            {item.description ? (
                              <span className="text-xs text-muted-foreground">{item.description}</span>
                            ) : null}
                          </A>
                        </li>
                      ))}
                    </ul>
                  ) : null}
                </>
              ) : (
                <A
                  href={link.href}
                  aria-current={link.href === currentHref ? "page" : undefined}
                  className={linkClass(link.href)}
                >
                  {link.label}
                </A>
              )}
            </li>
          ))}
        </ul>

        <span className="flex-1" />

        {extras ? <div className="flex items-center gap-1">{extras}</div> : null}

        <div className="hidden items-center gap-2 md:flex">
          {secondaryAction ? (
            <A href={secondaryAction.href} className={linkClass(secondaryAction.href)}>
              {secondaryAction.label}
            </A>
          ) : null}
          {primaryAction ? (
            <A
              href={primaryAction.href}
              className="btn-brand btn-press inline-flex min-h-9 items-center gap-1.5 rounded-md px-3.5 text-sm font-medium outline-none focus-visible:ring-2 focus-visible:ring-foreground/40"
            >
              {primaryAction.label}
            </A>
          ) : null}
        </div>

        <button
          type="button"
          onClick={() => setOpen((current) => !current)}
          aria-expanded={open}
          aria-controls="site-navbar-panel"
          aria-label={open ? "Close menu" : "Open menu"}
          className="-mr-1 flex h-10 w-10 items-center justify-center rounded-lg text-foreground outline-none transition-colors duration-150 hover:bg-foreground/[0.06] focus-visible:ring-2 focus-visible:ring-foreground/40 md:hidden"
        >
          {open ? <X aria-hidden="true" className="h-5 w-5" /> : <Menu aria-hidden="true" className="h-5 w-5" />}
        </button>
      </nav>

      {/* A grid-template-rows interpolation, so the panel opens without the browser laying out on every frame. */}
      <div
        id="site-navbar-panel"
        className={cn(
          "grid overflow-hidden transition-[grid-template-rows] duration-200 ease-out motion-reduce:transition-none md:hidden",
          open ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
        )}
      >
        <div className="min-h-0">
          <div
            inert={!open}
            className="flex flex-col gap-1 border-t border-border bg-background/95 px-4 pb-4 pt-3 backdrop-blur-md sm:px-6"
          >
            {links.map((link) =>
              link.items ? (
                <div key={link.label} className="flex flex-col">
                  <button
                    type="button"
                    onClick={() =>
                      setExpanded((current) => {
                        const next = new Set(current);
                        if (next.has(link.label)) next.delete(link.label);
                        else next.add(link.label);
                        return next;
                      })
                    }
                    aria-expanded={expanded.has(link.label)}
                    className={cn(linkClass(), "flex min-h-11 items-center justify-between gap-2")}
                  >
                    {link.label}
                    <ChevronDown
                      aria-hidden="true"
                      className={cn(
                        "h-4 w-4 transition-transform duration-200 ease-out motion-reduce:transition-none",
                        expanded.has(link.label) && "rotate-180",
                      )}
                    />
                  </button>

                  <div
                    className={cn(
                      "grid transition-[grid-template-rows] duration-200 ease-out motion-reduce:transition-none",
                      expanded.has(link.label) ? "grid-rows-[1fr]" : "grid-rows-[0fr]",
                    )}
                  >
                    <div className="min-h-0 overflow-hidden">
                      {/* Indented against a rail, so a child reads as belonging to the row above it. */}
                      <div inert={!expanded.has(link.label)} className="ml-3 flex flex-col border-l border-border pl-2">
                        {link.items.map((item) => (
                          <A
                            key={item.href}
                            href={item.href}
                            onClick={() => setOpen(false)}
                            aria-current={item.href === currentHref ? "page" : undefined}
                            className={cn(linkClass(item.href), "flex min-h-11 items-center")}
                          >
                            {item.label}
                          </A>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <A
                  key={link.label}
                  href={link.href}
                  onClick={() => setOpen(false)}
                  aria-current={link.href === currentHref ? "page" : undefined}
                  className={cn(linkClass(link.href), "flex min-h-11 items-center")}
                >
                  {link.label}
                </A>
              ),
            )}

            <div className="mt-2 flex items-center gap-2">
              {secondaryAction ? (
                <A
                  href={secondaryAction.href}
                  onClick={() => setOpen(false)}
                  className="inline-flex min-h-11 flex-1 items-center justify-center rounded-md px-4 text-sm font-medium text-foreground shadow-[0_0_0_1px_var(--border-strong)] outline-none focus-visible:ring-2 focus-visible:ring-foreground/40"
                >
                  {secondaryAction.label}
                </A>
              ) : null}
              {primaryAction ? (
                <A
                  href={primaryAction.href}
                  onClick={() => setOpen(false)}
                  className="btn-brand inline-flex min-h-11 flex-1 items-center justify-center rounded-md px-4 text-sm font-medium outline-none focus-visible:ring-2 focus-visible:ring-foreground/40"
                >
                  {primaryAction.label}
                </A>
              ) : null}
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
