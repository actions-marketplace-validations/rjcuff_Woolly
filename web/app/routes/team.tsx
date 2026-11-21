import type { Route } from './+types/team';
import { FileText } from 'lucide-react';
import { GitHubIcon } from '@/components/brand';
import { BuiltFor, type Case } from '@/components/sky/built-for';
import { SkyFoot, SkyNav } from '@/components/sky/chrome';
import { Closing } from '@/components/sky/closing';
import { Faq } from '@/components/sky/faq';
import { SkyHero } from '@/components/sky/hero';
import { ArrowLink, Button } from '@/components/sky/parts';
import { TeamPricing, TeamSteps } from '@/components/sky/team';
import { LogoIcon, WorksWith, type WorksWithTile } from '@/components/sky/works-with';
import { getStars } from '@/lib/github';
import { integrations } from '@/lib/integrations';
import { seo } from '@/lib/seo';
import { teamInstallUrl, teamLoginUrl } from '@/lib/shared';

export function meta(_: Route.MetaArgs) {
  return seo({
    title: 'Woolly Team | Accurate agent context across every repository',
    description:
      'Woolly Team flags the CLAUDE.md, AGENTS.md and project memory lines each pull request makes stale, across every repository. It never reads your code.',
    path: '/team',
    image: '/og/team.png',
    imageAlt: 'A woolly mammoth floating on a balloon over a snowy ice cave and glacier',
  });
}

export async function loader() {
  return { stars: await getStars() };
}

const logo = (name: string) => <LogoIcon path={integrations.find((i) => i.name === name)?.path ?? ''} />;
const DOCS = '/docs/memory-files';

// The agent files Woolly Team checks on every pull request.
const FILES: WorksWithTile[] = [
  { name: 'CLAUDE.md', href: DOCS, icon: logo('Claude') },
  { name: 'AGENTS.md', href: DOCS, icon: <FileText className="size-7 sm:size-8" aria-hidden="true" /> },
  { name: 'GEMINI.md', href: DOCS, icon: logo('Gemini CLI') },
  { name: 'Cursor rules', href: DOCS, icon: logo('Cursor') },
  { name: 'Windsurf rules', href: DOCS, icon: logo('Windsurf') },
  { name: 'Copilot', href: DOCS, icon: <GitHubIcon className="size-7 sm:size-8" /> },
];

const FEATURES: Case[] = [
  { title: 'Catch stale instructions', body: 'Every pull request is checked, and lines that name removed or renamed code are flagged.' },
  { title: 'Every repository in one place', body: 'Coverage, open findings and recent checks across your organization, busiest first.' },
  { title: 'Evidence attached', body: 'Each finding links to the exact line, the commit that changed the code and its pull request.' },
  { title: 'Clear findings as a team', body: 'Mark findings fixed, dismissed or false positive, with a record of who decided and when.' },
  { title: 'Your code never leaves GitHub', body: 'Checks run in your own GitHub Actions. Woolly Team receives only counts and file pointers.' },
];

const QUESTIONS = [
  {
    question: 'Does Woolly Team read our source code?',
    answer: 'No. Checks run in your GitHub Actions, and the app can read Actions only. It receives counts and file pointers.',
  },
  {
    question: 'Does it verify that our code is correct?',
    answer: 'No. It finds notes that name code a commit removed or renamed, with the evidence. Your team decides what to change.',
  },
  {
    question: 'Which agents does it support?',
    answer: 'Any agent that reads CLAUDE.md, AGENTS.md, GEMINI.md, Cursor or Windsurf rules, or Copilot instructions.',
  },
  {
    question: 'How do we remove our data?',
    answer: 'Export everything from your dashboard at any time. Uninstalling the app deletes all of it.',
  },
];

/** Woolly Team, built from the same pieces as the homepage. */
export default function Team({ loaderData }: Route.ComponentProps) {
  return (
    <div className="home-sky relative min-h-svh bg-white">
      <SkyNav stars={loaderData.stars} className="absolute inset-x-0 top-0" />
      <SkyHero
        title="Accurate agent context across every repository"
        lead="Woolly Team flags the lines in CLAUDE.md, AGENTS.md and project memory that each pull request makes stale, so your team fixes them before agents act on them."
        actions={
          <>
            <Button href={teamInstallUrl}>Get started</Button>
            <Button href={teamLoginUrl} tone="white">
              Sign in to your dashboard
            </Button>
          </>
        }
      />
      <WorksWith
        title="Checks the files your agents already read"
        lead="Woolly Team runs the open source Woolly check on every pull request, so whatever your team uses, its instructions stay true to the code."
        link={{ label: 'See every supported file', href: DOCS }}
        tiles={FILES}
        more={null}
      />
      <BuiltFor title="Everything your team needs to trust its agents" cases={FEATURES} />
      <TeamSteps />
      <TeamPricing installUrl={teamInstallUrl} />
      <Faq items={QUESTIONS} lead="Anything else is in the docs." link={{ label: 'Read the docs', href: '/docs/pull-requests' }} />
      <Closing
        title="Give every agent on your team accurate context"
        line="One dashboard for every repository, with the evidence attached."
        actions={
          <>
            <Button href={teamInstallUrl}>Get started</Button>
            <ArrowLink href={teamLoginUrl} className="text-[var(--ink)] hover:text-[var(--accent)]">
              Sign in
            </ArrowLink>
          </>
        }
      />
      <SkyFoot />
    </div>
  );
}
