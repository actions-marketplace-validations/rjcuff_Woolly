import { githubUrl } from '@/lib/shared';
import { cn } from '@/lib/utils';
import { ArrowLink, H2, LEAD, WRAP } from './parts';

// Every number here comes from bench/RESULTS.md. Update both together.
const NUMBERS: { value: string; label: string; sub: string }[] = [
  { value: '11 of 11', label: 'stale line flags were right', sub: '20 open source repos' },
  { value: '4 of 20', label: 'repos named code that’s gone', sub: '2,990 lines checked' },
  { value: '93%', label: 'of stale facts caught', sub: '28 of 30, none flagged wrongly' },
  { value: '0.7 s', label: 'for a woolly check', sub: 'median, on a laptop' },
];

/** The benchmark numbers, each on a light blue card. */
export function Benchmarks() {
  return (
    <section aria-label="Benchmarks" className="bg-[var(--mist)] py-20 sm:py-28">
      <div className={WRAP}>
        <div className="flex flex-wrap items-end justify-between gap-x-10 gap-y-5">
          <div className="max-w-[580px]">
            <h2 className={H2}>Measured on real repositories</h2>
            <p className={cn(LEAD, 'mt-5')}>
              We ran Woolly on the agent files of 20 open source projects, using git alone. Every result is published with its evidence, and you can run it yourself.
            </p>
          </div>
          <ArrowLink href={`${githubUrl}/blob/main/bench/RESULTS.md`}>See every result</ArrowLink>
        </div>
        <dl className="mt-10 grid grid-cols-2 gap-3 sm:mt-12 sm:gap-4 lg:grid-cols-4">
          {NUMBERS.map((n) => (
            <div key={n.label} className="flex flex-col rounded-3xl bg-[var(--tint-blue)] p-6 sm:p-7">
              <dt className="order-2 mt-4 text-[15px] font-semibold text-[var(--ink)]">{n.label}</dt>
              <dd className="serif order-1 text-[38px] leading-none tracking-[-0.02em] text-[var(--ink)] sm:text-[52px]">{n.value}</dd>
              <dd className="order-3 mt-1 text-[13.5px] text-[var(--muted)]">{n.sub}</dd>
            </div>
          ))}
        </dl>
      </div>
    </section>
  );
}
