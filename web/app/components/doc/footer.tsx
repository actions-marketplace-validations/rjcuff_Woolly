import { Link } from 'react-router';
import { Wordmark } from '@/components/brand';
import { Label, SmartLink } from '@/components/doc/parts';
import { blueskyUrl, githubUrl, linkedinUrl } from '@/lib/shared';

const FOOTER: { title: string; links: { label: string; href: string }[] }[] = [
  {
    title: 'Product',
    links: [
      { label: 'Download', href: '/download' },
      { label: 'Project memory', href: '/docs/project-memory' },
      { label: 'Benchmark', href: '/docs/benchmark' },
      { label: 'Changelog', href: `${githubUrl}/blob/main/CHANGELOG.md` },
    ],
  },
  {
    title: 'Developers',
    links: [
      { label: 'Quickstart', href: '/docs/quickstart' },
      { label: 'Connect your apps', href: '/docs/connect' },
      { label: 'JavaScript', href: '/docs/javascript' },
      { label: 'Python', href: '/docs/python' },
      { label: 'Contributing', href: `${githubUrl}/blob/main/CONTRIBUTING.md` },
    ],
  },
  {
    title: 'Connect',
    links: [
      { label: 'GitHub', href: githubUrl },
      { label: 'Bluesky', href: blueskyUrl },
      { label: 'LinkedIn', href: linkedinUrl },
      { label: 'Report an issue', href: `${githubUrl}/issues` },
    ],
  },
];

/** The bottom of every page except the docs. */
export function DocFooter() {
  return (
    <footer className="border-t border-fd-border pt-12 pb-10">
      <nav aria-label="Footer">
      <div className="grid gap-10 sm:grid-cols-[1.3fr_1fr_1fr_1fr]">
        <div>
          <Wordmark className="text-[22px]" />
          <p className="mt-3 max-w-[220px] text-[14px] leading-relaxed text-fd-muted-foreground">
            Your agents share one brain. Local, private and MIT licensed.
          </p>
        </div>
        {FOOTER.map((col) => (
          <div key={col.title}>
            <Label className="text-fd-foreground">{col.title}</Label>
            <ul className="mt-4 flex flex-col gap-2.5">
              {col.links.map((l) => (
                <li key={l.label}>
                  <SmartLink href={l.href} className="group text-[14px] text-fd-muted-foreground transition-colors hover:text-fd-foreground">
                    <span className="link-draw">{l.label}</span>
                  </SmartLink>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <div className="mt-12 flex flex-wrap items-center justify-between gap-4 border-t border-fd-border pt-6 text-[13px] text-fd-muted-foreground">
        <div className="flex flex-wrap items-center gap-x-5 gap-y-2">
          <span>© {new Date().getFullYear()} Woolly™</span>
          <Link to="/privacy" className="hover:text-fd-foreground">
            Privacy
          </Link>
          <Link to="/terms" className="hover:text-fd-foreground">
            Terms
          </Link>
        </div>
        <div className="flex items-center gap-5">
          <a href="#top" className="group inline-flex items-center gap-1 hover:text-fd-foreground">
            Back to top
            <span aria-hidden="true" className="transition-transform duration-200 ease-ui group-hover:-translate-y-0.5 motion-reduce:transition-none">
              ↑
            </span>
          </a>
        </div>
      </div>
      </nav>
    </footer>
  );
}
