import { createGetUrl } from 'fumadocs-core/source';

export const appName = 'Woolly';
export const tagline = 'Project memory your coding agents can trust';
export const siteDescription =
  'Woolly flags lines in CLAUDE.md, AGENTS.md and agent memory that a commit made wrong, with the commit that did it. Local, no model, free and open source.';
// The public URL of the site. Update this once the domain is live.
export const siteUrl = 'https://usewoolly.com';
export const docsRoute = '/docs';
export const docsImageRoute = '/og/docs';
export const docsContentRoute = '/llms.mdx/docs';

export const gitConfig = {
  user: 'rjcuff',
  repo: 'woolly',
  branch: 'main',
};
export const githubUrl = `https://github.com/${gitConfig.user}/${gitConfig.repo}`;
// Woolly Team, the paid service for teams with many repos.
export const teamUrl = 'https://team.usewoolly.com';
export const blueskyUrl = 'https://bsky.app/profile/usewoolly.com';
export const linkedinUrl = 'https://www.linkedin.com/company/usewoolly';
export const productHuntUrl =
  'https://www.producthunt.com/products/woolly?embed=true&utm_source=badge-featured&utm_medium=badge&utm_campaign=badge-woolly';
export const productHuntBadge = (theme: 'light' | 'dark') =>
  `https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=1315543&theme=${theme}`;

const getContentUrl = createGetUrl(docsContentRoute);

export function getPageMarkdownUrl(page: { slugs: string[]; locale?: string }) {
  const segments = [...page.slugs, 'content.md'];

  return { segments, url: getContentUrl(segments, page.locale) };
}

const getImageUrl = createGetUrl(docsImageRoute);

export function getPageImageUrl(page: { slugs: string[]; locale?: string }) {
  const segments = [...page.slugs, 'image.png'];

  return { segments, url: getImageUrl(segments, page.locale) };
}
