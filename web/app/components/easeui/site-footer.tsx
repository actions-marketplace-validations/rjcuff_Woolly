import type { ReactNode } from "react";
import { Link } from "react-router";
import { cn } from "@/lib/utils";

export interface FooterLink {
  label: string;
  href: string;
  /** Opens in a new tab and carries the safe rel. */
  external?: boolean;
}

export interface FooterColumn {
  title: string;
  links: FooterLink[];
}

export interface SiteFooterProps {
  brand: { label: ReactNode; href: string };
  /** One line under the wordmark. */
  tagline?: string;
  columns: FooterColumn[];
  /** Small marks on the bottom row, usually social. Each needs a label for the reader. */
  social?: { label: string; href: string; icon: ReactNode }[];
  /** Bottom row links, such as Privacy and Terms. */
  legal?: FooterLink[];
  /** Overrides the generated copyright line. */
  copyright?: string;
  className?: string;
}

/** Site paths stay in the app; anything else is a plain link. */
function A({ href = "", ...props }: React.ComponentProps<"a">) {
  return href.startsWith("/") ? <Link to={href} {...props} /> : <a href={href} {...props} />;
}

/**
 * The bottom of a page: the wordmark and what the product is, the link columns, and a bottom row with the
 * legal links.
 */
export function SiteFooter({
  brand,
  tagline,
  columns,
  social,
  legal,
  copyright,
  className,
}: SiteFooterProps) {
  const year = new Date().getFullYear();

  const linkClass =
    "rounded text-sm text-muted-foreground outline-none transition-colors duration-150 hover:text-foreground focus-visible:ring-2 focus-visible:ring-foreground/40";

  return (
    <footer className={cn("w-full border-t border-border", className)}>
      <div className="mx-auto max-w-[1200px] px-[max(20px,env(safe-area-inset-left))] py-12 sm:py-16">
        <div className="flex flex-col gap-10 md:flex-row md:justify-between md:gap-16">
          <div className="max-w-xs">
            <A
              href={brand.href}
              className="rounded-lg text-sm font-semibold tracking-tight text-foreground outline-none focus-visible:ring-2 focus-visible:ring-foreground/40"
            >
              {brand.label}
            </A>
            {tagline ? <p className="mt-3 text-sm leading-6 text-muted-foreground">{tagline}</p> : null}
          </div>

          {/* Two columns on a phone, all of them in a row once there is width. */}
          <div className="grid grid-cols-2 gap-8 sm:grid-cols-3 md:gap-16">
            {columns.map((column) => (
              <div key={column.title}>
                <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground/70">{column.title}</h2>
                <ul className="mt-3 flex flex-col gap-2.5">
                  {column.links.map((link) => (
                    <li key={link.href}>
                      <A
                        href={link.href}
                        className={linkClass}
                        {...(link.external ? { target: "_blank", rel: "noreferrer noopener" } : {})}
                      >
                        {link.label}
                      </A>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-12 flex flex-col gap-4 border-t border-border pt-6 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-xs text-muted-foreground">{copyright ?? `© ${year} ${brand.label}. All rights reserved.`}</p>

          <nav aria-label="Footer" className="flex items-center gap-4">
            {legal?.map((link) => (
              <A key={link.href} href={link.href} className={cn(linkClass, "text-xs")}>
                {link.label}
              </A>
            ))}

            {social && social.length > 0 ? (
              <ul className="flex items-center gap-1">
                {social.map((mark) => (
                  <li key={mark.href}>
                    <A
                      href={mark.href}
                      target="_blank"
                      rel="noreferrer noopener"
                      aria-label={mark.label}
                      className="flex h-9 w-9 items-center justify-center rounded-full text-muted-foreground outline-none transition-colors duration-150 hover:bg-foreground/[0.06] hover:text-foreground focus-visible:ring-2 focus-visible:ring-foreground/40"
                    >
                      {mark.icon}
                    </A>
                  </li>
                ))}
              </ul>
            ) : null}
          </nav>
        </div>
      </div>
    </footer>
  );
}
