import { Check } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button, H2, LEAD, WRAP } from './parts';

const STEPS: { title: string; body: string }[] = [
  { title: 'Install the GitHub App', body: 'Choose the repositories to watch. The app can read Actions only.' },
  { title: 'Add the workflow', body: 'Copy one workflow file into each repository. It runs the open source Woolly check.' },
  { title: 'Open a pull request', body: 'Findings appear on your dashboard and as a comment on the pull request.' },
];

/** Three numbered cards on the misty band. */
export function TeamSteps() {
  return (
    <section id="how" aria-label="How it works" className="scroll-mt-6 bg-[var(--mist)] py-20 sm:py-28">
      <div className={WRAP}>
        <h2 className={H2}>Up and running in five minutes</h2>
        <p className={cn(LEAD, 'mt-5')}>Three steps, no code to change.</p>
        <ol className="mt-10 grid gap-4 sm:mt-12 md:grid-cols-3">
          {STEPS.map((s, i) => (
            <li key={s.title} className="rounded-3xl bg-[var(--tint-blue)] p-7 sm:p-8">
              <span className="serif block text-[52px] leading-none text-[var(--ink)]">{i + 1}</span>
              <h3 className="mt-5 text-[17px] font-bold text-[var(--ink)]">{s.title}</h3>
              <p className="mt-2 text-[15.5px] leading-[1.6] text-pretty text-[var(--body)]">{s.body}</p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}

type Plan = { name: string; price: string; per: string; blurb: string; perks: string[]; cta: { label: string; href: string }; featured?: boolean };

function PlanCard({ plan }: { plan: Plan }) {
  const f = plan.featured;
  return (
    <div className={cn('flex flex-col items-start gap-4 rounded-3xl p-8', f ? 'bg-[var(--action)] text-white shadow-[0_24px_48px_-24px_rgb(30_39_102/0.6)]' : 'bg-[var(--mist)] ring-1 ring-[var(--line)]')}>
      <h3 className="serif text-[26px]">{plan.name}</h3>
      <p className="flex flex-wrap items-baseline gap-2">
        <strong className="serif text-[48px] leading-none font-medium">{plan.price}</strong>
        <span className={cn('text-[15px]', f ? 'text-[#c3cdff]' : 'text-[var(--muted)]')}>{plan.per}</span>
      </p>
      <p className={cn('text-[15.5px] leading-[1.6]', f ? 'text-[#c3cdff]' : 'text-[var(--body)]')}>{plan.blurb}</p>
      <ul className="mb-2 flex flex-col gap-2.5 text-[15.5px]">
        {plan.perks.map((p) => (
          <li key={p} className="flex gap-2.5">
            <span className={cn('mt-0.5 flex size-[18px] shrink-0 items-center justify-center rounded-full', f ? 'bg-white/15' : 'bg-[var(--tint-blue)]')}>
              <Check className="size-3" strokeWidth={3} aria-hidden="true" />
            </span>
            {p}
          </li>
        ))}
      </ul>
      <Button href={plan.cta.href} tone="white" className={cn('mt-auto', !f && 'bg-[var(--tint-blue)] ring-0 shadow-none hover:bg-[#dce2ff]')}>
        {plan.cta.label}
      </Button>
    </div>
  );
}

/** The free check next to the night blue Team plan. */
export function TeamPricing({ installUrl }: { installUrl: string }) {
  const plans: Plan[] = [
    {
      name: 'Woolly',
      price: 'Free',
      per: 'Open source',
      blurb: 'For individual developers and single repositories.',
      perks: ['Checks agent files and memory against your code', 'Pull request comments for stale lines', 'Runs locally and in your own CI'],
      cta: { label: 'Download Woolly', href: '/download' },
    },
    {
      name: 'Woolly Team',
      price: '$299',
      per: 'per month per organization',
      blurb: 'For engineering teams running agents across many repositories.',
      perks: ['Everything in Woolly', 'Organization-wide dashboard', 'Shared review queue and activity', 'Up to 25 repositories, unlimited members', 'Data export and deletion controls'],
      cta: { label: 'Get started', href: installUrl },
      featured: true,
    },
  ];
  return (
    <section id="pricing" aria-label="Pricing" className="scroll-mt-6 bg-white py-20 sm:py-28">
      <div className={WRAP}>
        <h2 className={H2}>Simple pricing for every team</h2>
        <p className={cn(LEAD, 'mt-5 max-w-[520px]')}>Start with the open source check. Add Woolly Team when you need to see across repositories.</p>
        <div className="mt-10 grid gap-4 sm:mt-12 md:grid-cols-2">
          {plans.map((p) => (
            <PlanCard key={p.name} plan={p} />
          ))}
        </div>
      </div>
    </section>
  );
}
