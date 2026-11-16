import { gitConfig, githubUrl } from './shared';

export type Release = {
  version: string;
  /** One sentence about the headline change, as plain text. */
  summary: string;
  /** ISO date the release was published, when GitHub could tell us. */
  date: string | null;
  href: string;
};

/** Turns a Markdown bullet into plain text: links keep their words, code loses its backticks. */
function plain(text: string): string {
  return text
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1')
    .replace(/[`*_]/g, '')
    .trim();
}

// Short first sentences like "A Windows app." say too little, so keep adding
// sentences until there is enough to read.
const MIN_SUMMARY = 60;

/** The opening sentences of a bullet, at least MIN_SUMMARY characters when it has them. */
function opening(text: string): string {
  // A sentence ends at a stop followed by a space or the end, so a name like Woolly-windows.exe stays whole.
  const sentences = text.match(/.+?[.!?]+(?=\s|$)\s*|.+$/g) ?? [text];
  let out = '';
  for (const sentence of sentences) {
    out += sentence;
    if (out.trim().length >= MIN_SUMMARY) break;
  }
  return out.trim();
}

/** parseChangelog reads the newest releases from CHANGELOG.md. */
export function parseChangelog(markdown: string, limit: number): Omit<Release, 'date'>[] {
  // Only released versions; a heading like "Unreleased" isn't news yet.
  const sections = markdown
    .split(/^## /m)
    .slice(1)
    .filter((section) => /^\d+\.\d+\.\d+\s/.test(section));
  return sections.slice(0, limit).map((section) => {
    const [heading, ...rest] = section.split('\n');
    const version = heading.trim();
    const bullet = rest.find((line) => line.startsWith('- ')) ?? '';
    return {
      version,
      summary: opening(plain(bullet.slice(2))),
      href: `${githubUrl}/releases/tag/v${version}`,
    };
  });
}

/** Release dates from GitHub, by version. Read when the site is built, never by visitors. */
async function releaseDates(): Promise<Map<string, string>> {
  try {
    const res = await fetch(`https://api.github.com/repos/${gitConfig.user}/${gitConfig.repo}/releases?per_page=20`, {
      headers: { Accept: 'application/vnd.github+json' },
      signal: AbortSignal.timeout(5000),
    });
    if (!res.ok) return new Map();
    const data = (await res.json()) as { tag_name?: string; published_at?: string }[];
    return new Map(
      data
        .filter((r) => r.tag_name && r.published_at)
        .map((r) => [r.tag_name!.replace(/^v/, ''), r.published_at!] as const),
    );
  } catch {
    return new Map();
  }
}

/** getReleases returns the newest releases in a changelog, dated when GitHub answers. */
export async function getReleases(changelog: string, limit = 4): Promise<Release[]> {
  const dates = await releaseDates();
  return parseChangelog(changelog, limit).map((r) => ({ ...r, date: dates.get(r.version) ?? null }));
}
