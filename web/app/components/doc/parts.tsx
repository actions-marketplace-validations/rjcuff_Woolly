import { ArrowRight, ArrowUpRight } from 'lucide-react';
import type { ReactNode } from 'react';
import { Link } from 'react-router';
import { cn } from '@/lib/utils';

/** Small uppercase mono text, for labels that sit above or beside the real words. */
export function Label({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <span className={cn('font-mono text-[11px] uppercase tracking-[0.14em] text-fd-muted-foreground', className)}>
      {children}
    </span>
  );
}

export function SmartLink({ href, className, children, ...rest }: { href: string; className?: string; children: ReactNode } & Omit<React.ComponentProps<'a'>, 'href'>) {
  return href.startsWith('/') && !href.startsWith('//') ? (
    <Link to={href} className={className} {...rest}>
      {children}
    </Link>
  ) : (
    <a href={href} className={className} {...rest}>
      {children}
    </a>
  );
}

/** A quiet link with an arrow, for "read more" at the end of a section. */
export function MoreLink({ href, children }: { href: string; children: ReactNode }) {
  return (
    <SmartLink href={href} className="group inline-flex items-center gap-1 text-[15px] font-medium text-fd-foreground underline-offset-4 hover:underline">
      {children}
      <ArrowUpRight aria-hidden="true" className="size-3.5 text-fd-muted-foreground transition-[transform,color] duration-200 ease-out group-hover:-translate-y-0.5 group-hover:translate-x-0.5 group-hover:text-brand" />
    </SmartLink>
  );
}

/** The main button: square, Woolly blue, with the mammoth in a white block at its start. */
export function BlockButton({ href, children, className }: { href: string; children: ReactNode; className?: string }) {
  return (
    <SmartLink
      href={href}
      className={cn(
        'btn-press group inline-flex h-11 items-stretch bg-brand p-[3px] text-brand-foreground outline-none transition-colors hover:bg-[color-mix(in_oklch,var(--color-brand)_88%,black)] focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-offset-2 focus-visible:ring-offset-fd-background',
        className,
      )}
    >
      {/* Two arrows: on hover the first slides out the right as the second slides in from the left, and back again off hover. */}
      <span aria-hidden="true" className="relative flex w-[38px] items-center justify-center overflow-hidden bg-white text-brand">
        <ArrowRight className="size-[18px] transition-transform duration-300 ease-ui group-hover:translate-x-[38px] motion-reduce:transition-none" />
        <ArrowRight className="absolute size-[18px] -translate-x-[38px] transition-transform duration-300 ease-ui group-hover:translate-x-0 motion-reduce:transition-none" />
      </span>
      <span className="flex flex-1 items-center justify-center px-4 font-mono text-[13px] font-medium uppercase tracking-[0.12em]">{children}</span>
    </SmartLink>
  );
}

/** Corner marks, like crop marks on a print. */
function Corners() {
  const mark = 'pointer-events-none absolute size-2 border-brand/60 transition-[translate] duration-200 ease-out';
  return (
    <span aria-hidden="true">
      <span className={`${mark} left-0 top-0 border-l border-t group-hover:-translate-x-0.5 group-hover:-translate-y-0.5`} />
      <span className={`${mark} right-0 top-0 border-r border-t group-hover:translate-x-0.5 group-hover:-translate-y-0.5`} />
      <span className={`${mark} bottom-0 left-0 border-b border-l group-hover:-translate-x-0.5 group-hover:translate-y-0.5`} />
      <span className={`${mark} bottom-0 right-0 border-b border-r group-hover:translate-x-0.5 group-hover:translate-y-0.5`} />
    </span>
  );
}

/** The second button: a pale blue block framed by corner marks. */
export function FrameButton({ href, children, className }: { href: string; children: ReactNode; className?: string }) {
  return (
    <SmartLink
      href={href}
      className={cn(
        'btn-press group relative inline-flex h-11 items-center justify-center gap-2 bg-brand/[0.07] px-4 font-mono text-[13px] font-medium uppercase tracking-[0.12em] text-fd-foreground outline-none transition-colors hover:bg-brand/[0.12] focus-visible:ring-2 focus-visible:ring-brand',
        className,
      )}
    >
      <Corners />
      {children}
      <ArrowUpRight aria-hidden="true" className="size-4 transition-transform duration-200 ease-out group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
    </SmartLink>
  );
}
