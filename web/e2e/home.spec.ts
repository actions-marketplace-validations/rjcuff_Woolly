import { expect, test, type Page } from '@playwright/test';

/** The nav is always on screen now; there's no drawer to open. */
async function openNav(_page: Page, _isMobile: boolean) {}

/** A link in the main nav. */
async function navLink(page: Page, _isMobile: boolean, name: string) {
  return page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name, exact: true });
}

test('home splits on large screens, then flows into full-width sections', async ({ page, isMobile }) => {
  await page.goto('/');
  await expect(page).toHaveTitle('Woolly | Project memory your coding agents can trust');
  const hero = page.getByRole('region', { name: 'Hero' });
  await expect(hero).toBeVisible();
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Memory that knows when it’s wrong.');
  await expect(hero.getByRole('link', { name: 'Get started' })).toBeVisible();
  for (const step of ['Verify every note your agents read', 'Proven on real repositories', 'Precise, evidence-based detection', 'Local and private by default']) {
    await expect(page.getByRole('heading', { name: step })).toBeAttached();
  }
  await expect(page.getByRole('region', { name: 'Scrolling panels' })).toBeAttached();
  await expect(page.getByRole('region', { name: 'Measured on real repositories' })).toBeAttached();
  // The numbers are the published benchmark's.
  await expect(page.getByText('11 of 11')).toBeVisible();
  // The newest feature is the first thing in the hero.
  await expect(hero.getByRole('link', { name: /woolly check reads CLAUDE\.md/ })).toHaveAttribute('href', '/docs/memory-files');

  // Never sideways, on any screen.
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)).toBe(false);

  if (isMobile) return;
  // The hero stays under the nav while the panels scroll past it...
  await page.evaluate(() => window.scrollTo(0, 1200));
  await expect.poll(async () => Math.round((await hero.boundingBox())?.y ?? -1)).toBe(56);
  // ...and scrolls away with everything else once they run out.
  await page.getByRole('region', { name: 'Who it’s for' }).scrollIntoViewIfNeeded();
  await expect.poll(async () => (await hero.boundingBox())?.y ?? 0).toBeLessThan(0);
});

test('nav links work and the site stays light', async ({ page, isMobile }) => {
  await page.emulateMedia({ colorScheme: 'dark' });
  await page.goto('/');
  await expect(page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name: /GitHub/ })).toHaveAttribute(
    'href',
    'https://github.com/rjcuff/woolly',
  );
  await expect(await navLink(page, isMobile, 'Docs')).toBeVisible();

  // Light only, even when the computer asks for dark, and no theme switch.
  for (const path of ['/', '/download', '/docs']) {
    await page.goto(path);
    expect(await page.evaluate(() => document.documentElement.classList.contains('dark')), path).toBe(false);
    await expect(page.getByRole('button', { name: /switch to (dark|light) mode|toggle theme/i })).toHaveCount(0);
  }
});

test('link previews point at the Woolly image', async ({ page, request }) => {
  await page.goto('/');
  const image = await page.locator('meta[property="og:image"]').getAttribute('content');
  expect(image).toBe('https://usewoolly.com/og/docs/image.png');

  const res = await request.get('/og/docs/image.png');
  expect(res.status()).toBe(200);
  expect(res.headers()['content-type']).toContain('image/png');
});

test('sitemap, robots and install script are served', async ({ request }) => {
  for (const path of ['/sitemap.xml', '/robots.txt', '/install.sh', '/install.ps1', '/og/docs/profile.png']) {
    expect((await request.get(path)).status(), path).toBe(200);
  }
  expect(await (await request.get('/sitemap.xml')).text()).toContain('https://usewoolly.com/docs/quickstart');
});

test('download page offers the Mac and Windows apps and install commands', async ({ page, isMobile }) => {
  await page.goto('/');
  await (await navLink(page, isMobile, 'Download')).click();
  await expect(page).toHaveURL(/\/download$/);
  await page.getByRole('tab', { name: 'macOS' }).click();
  await expect(page.getByRole('link', { name: 'Download for Mac' }).first()).toHaveAttribute(
    'href',
    /Woolly-mac\.dmg$/,
  );
  await page.getByRole('tab', { name: 'Windows' }).click();
  await expect(page.getByRole('link', { name: 'Download for Windows' }).first()).toHaveAttribute(
    'href',
    /Woolly-windows\.exe$/,
  );
  await expect(page.getByText('irm https://usewoolly.com/install.ps1 | iex')).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)).toBe(false);
});


test('Teams link and the Woolly Team section point at team.usewoolly.com', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name: 'Teams', exact: true })).toHaveAttribute('href', 'https://team.usewoolly.com');
  const team = page.getByRole('region', { name: 'Woolly Team' });
  await expect(team.getByRole('heading', { name: /Visibility across every repository/ })).toBeVisible();
  await expect(team.getByRole('link', { name: 'Explore Woolly Team' })).toHaveAttribute('href', 'https://team.usewoolly.com');
  await expect(team.getByRole('link', { name: 'View pricing' })).toHaveAttribute('href', 'https://team.usewoolly.com/#pricing');
  await expect(team).toContainText('$299');
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
  expect(overflow).toBeLessThanOrEqual(0);
});
