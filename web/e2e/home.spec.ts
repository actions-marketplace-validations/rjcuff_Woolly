import { expect, test } from '@playwright/test';

test('home leads with the headline, setup and button', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveTitle('Woolly | Local memory for your AI tools');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('One memory for every AI tool you use');
  await expect(page.getByRole('main').getByRole('link', { name: 'Get started' }).first()).toBeInViewport();

  // Never sideways, on any screen.
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)).toBe(false);

  await page.getByRole('tab', { name: 'Your team' }).click();
  await expect(page.locator('#setup-panel')).toContainText('woolly init');

  await page.getByRole('main').getByRole('link', { name: 'Get started' }).first().click();
  await expect(page).toHaveURL(/\/docs\/quickstart$/);
});

test('nav links and theme toggle are visible and work', async ({ page }) => {
  await page.goto('/');
  const nav = page.getByRole('navigation', { name: 'Main' });
  await expect(nav.getByRole('link', { name: 'GitHub' })).toHaveAttribute('href', 'https://github.com/rjcuff/woolly');
  await expect(nav.getByRole('link', { name: 'Docs' })).toBeVisible();

  const toggle = nav.getByRole('button', { name: /switch to (dark|light) mode/i });
  await expect(toggle).toBeInViewport();
  const before = await page.evaluate(() => document.documentElement.classList.contains('dark'));
  await toggle.click();
  await expect
    .poll(() => page.evaluate(() => document.documentElement.classList.contains('dark')))
    .toBe(!before);
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

test('integrations and feature cards are shown', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByText('One command adds Woolly to the AI tools you already use')).toBeVisible();
  await expect(page.getByRole('link', { name: 'Ollama', exact: true })).toHaveAttribute('href', '/docs/ollama');
  await expect(page.getByRole('heading', { name: 'Readable' })).toBeVisible();
});

test('download page offers the Mac and Windows apps and install commands', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('link', { name: 'Download', exact: true }).click();
  await expect(page).toHaveURL(/\/download$/);
  await page.getByRole('tab', { name: 'macOS' }).click();
  await expect(page.getByRole('link', { name: 'Download for Mac' }).first()).toHaveAttribute('href', /Woolly-mac\.dmg$/);
  await page.getByRole('tab', { name: 'Windows' }).click();
  await expect(page.getByRole('link', { name: 'Download for Windows' }).first()).toHaveAttribute('href', /Woolly-windows\.exe$/);
  await expect(page.getByText('irm https://usewoolly.com/install.ps1 | iex')).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)).toBe(false);
});

test.describe('on Windows', () => {
  test.use({ userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36' });

  test('home page offers the Windows app', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('link', { name: 'Download for Windows' }).first()).toHaveAttribute('href', /Woolly-windows\.exe$/);
  });
});

test.describe('on a Mac', () => {
  test.use({ userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15' });

  test('home page offers the Mac app', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('link', { name: 'Download for Mac' }).first()).toHaveAttribute('href', /Woolly-mac\.dmg$/);
  });
});
