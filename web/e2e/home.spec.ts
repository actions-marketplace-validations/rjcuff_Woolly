import { expect, test, type Page } from '@playwright/test';

/** A link in the main nav. On phones the links are in the menu drawer. */
async function navLink(page: Page, isMobile: boolean, name: string) {
  if (isMobile) {
    await page.getByRole('button', { name: 'Open menu' }).click();
    return page.getByRole('navigation', { name: 'Menu links' }).getByRole('link', { name, exact: true });
  }
  return page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name, exact: true });
}

test('home shows the hero, the sections and the benchmark numbers', async ({ page, isMobile }) => {
  await page.goto('/');
  await expect(page).toHaveTitle('Woolly | Project memory your coding agents can trust');
  const hero = page.getByRole('region', { name: 'Hero' });
  await expect(hero).toBeVisible();
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Memory that knows when it’s wrong');
  await expect(hero.getByRole('link', { name: 'Get started' })).toHaveAttribute('href', '/docs');
  // Phones can't run the install command, so the hero leaves it out there.
  await expect(hero.getByRole('button', { name: 'Copy install command' })).toBeVisible({ visible: !isMobile });
  for (const name of ['Works with', 'Built for', 'Benchmarks', 'Questions', 'Get started']) {
    await expect(page.getByRole('region', { name, exact: true })).toBeAttached();
  }
  // The numbers are the published benchmark's.
  await expect(page.getByText('11 of 11')).toBeVisible();

  // Never sideways, on any screen.
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)).toBe(false);
});

test('one question is open at a time', async ({ page }) => {
  await page.goto('/');
  const faq = page.getByRole('region', { name: 'Questions' });
  const first = faq.getByRole('button', { name: 'Does Woolly send my data anywhere?' });
  const second = faq.getByRole('button', { name: 'Which apps work with Woolly?' });
  await expect(first).toHaveAttribute('aria-expanded', 'true');
  await second.click();
  await expect(second).toHaveAttribute('aria-expanded', 'true');
  await expect(first).toHaveAttribute('aria-expanded', 'false');
});

test('nav links work and the site stays light', async ({ page, isMobile }) => {
  await page.emulateMedia({ colorScheme: 'dark' });
  await page.goto('/');
  await expect(await navLink(page, isMobile, 'Docs')).toBeVisible();
  const github = isMobile ? page.getByRole('navigation', { name: 'Menu links' }).getByRole('link', { name: /GitHub/ }) : page.getByRole('link', { name: /^GitHub/ }).first();
  await expect(github).toHaveAttribute('href', 'https://github.com/rjcuff/woolly');

  // Light only, even when the computer asks for dark, and no theme switch.
  for (const path of ['/', '/download', '/docs', '/team']) {
    await page.goto(path);
    expect(await page.evaluate(() => document.documentElement.classList.contains('dark')), path).toBe(false);
    await expect(page.getByRole('button', { name: /switch to (dark|light) mode|toggle theme/i })).toHaveCount(0);
  }
});

test('the phone menu opens and closes', async ({ page, isMobile }) => {
  test.skip(!isMobile, 'Phones only');
  await page.goto('/');
  await page.getByRole('button', { name: 'Open menu' }).click();
  const menu = page.getByRole('dialog', { name: 'Menu' });
  await expect(menu).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(menu).toBeHidden();
});

test('link previews point at the homepage picture', async ({ page, request }) => {
  await page.goto('/');
  const image = await page.locator('meta[property="og:image"]').getAttribute('content');
  expect(image).toBe('https://usewoolly.com/og/home.png');

  for (const path of ['/og/home.png', '/og/docs/image.png']) {
    const res = await request.get(path);
    expect(res.status()).toBe(200);
    expect(res.headers()['content-type']).toContain('image/png');
  }
});

test('sitemap, robots and install script are served', async ({ request }) => {
  for (const path of ['/sitemap.xml', '/robots.txt', '/install.sh', '/install.ps1', '/og/docs/profile.png', '/og/team.png']) {
    expect((await request.get(path)).status(), path).toBe(200);
  }
  const sitemap = await (await request.get('/sitemap.xml')).text();
  expect(sitemap).toContain('https://usewoolly.com/docs/quickstart');
  expect(sitemap).toContain('https://usewoolly.com/team');
});

test('download page offers the Mac and Windows apps and install commands', async ({ page, isMobile }) => {
  await page.goto('/');
  await (await navLink(page, isMobile, 'Download')).click();
  await expect(page).toHaveURL(/\/download$/);
  await page.getByRole('tab', { name: 'macOS' }).click();
  await expect(page.getByRole('link', { name: 'Download for Mac' }).first()).toHaveAttribute('href', /Woolly-mac\.dmg$/);
  await page.getByRole('tab', { name: 'Windows' }).click();
  await expect(page.getByRole('link', { name: 'Download for Windows' }).first()).toHaveAttribute('href', /Woolly-windows\.exe$/);
  await expect(page.getByText('irm https://usewoolly.com/install.ps1 | iex')).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)).toBe(false);
});

test('the Teams link opens the Woolly Team page', async ({ page, isMobile }) => {
  await page.goto('/');
  await (await navLink(page, isMobile, 'Teams')).click();
  await expect(page).toHaveURL(/\/team$/);
  await expect(page).toHaveTitle('Woolly Team | Accurate agent context across every repository');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Accurate agent context across every repository');
  const install = 'https://github.com/apps/woolly-team/installations/new';
  const hero = page.getByRole('region', { name: 'Hero' });
  const start = hero.getByRole('link', { name: 'Get started' });
  const signIn = hero.getByRole('link', { name: 'Sign in' });
  await expect(start).toHaveAttribute('href', install);
  await expect(signIn).toHaveAttribute('href', 'https://team.usewoolly.com/login');
  // Both hero actions sit on one row, phones included.
  const [a, b] = [await start.boundingBox(), await signIn.boundingBox()];
  expect(Math.abs(a!.y + a!.height / 2 - (b!.y + b!.height / 2))).toBeLessThan(2);
  const pricing = page.getByRole('region', { name: 'Pricing' });
  for (const price of ['$0', '$99', '$299', 'Custom']) await expect(pricing).toContainText(price);
  for (const link of await pricing.getByRole('link', { name: 'Get started' }).all()) await expect(link).toHaveAttribute('href', install);
  await expect(pricing.getByRole('link', { name: 'Talk to us' })).toHaveAttribute('href', /issues\/new\?title=Woolly\+Team\+Enterprise$/);
  await expect(page.getByRole('region', { name: 'Get started' }).getByRole('link', { name: 'Sign in' })).toHaveAttribute('href', 'https://team.usewoolly.com/login');
  await expect(page.locator('meta[property="og:image"]')).toHaveAttribute('content', 'https://usewoolly.com/og/team.png');
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)).toBe(false);
});
