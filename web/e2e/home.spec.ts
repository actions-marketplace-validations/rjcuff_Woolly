import { expect, test, type Page } from '@playwright/test';

/** A link in the main nav. It floats at the left on a wide screen and sits on top otherwise, but it is always on screen. */
function navLink(page: Page, name: string) {
  return page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name, exact: true });
}

test('home leads with the headline and buttons', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveTitle('Woolly | One memory for your AI coding tools');
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('your agents share one brain');
  await expect(page.getByRole('main').getByRole('link', { name: 'Get started' }).first()).toBeInViewport();

  // Never sideways, on any screen.
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)).toBe(false);

  await page.getByRole('main').getByRole('link', { name: 'Get started' }).first().click();
  await expect(page).toHaveURL(/\/docs\/quickstart$/);
});

test('raw .md shows the page as Markdown and back', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'Raw .md' }).click();
  await expect(page.getByRole('main')).toContainText('# Woolly: your agents share one brain');
  await expect(page.getByRole('main')).toContainText('woolly connect');
  await page.getByRole('button', { name: 'Page' }).click();
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
});

test('nav links and theme toggle are visible and work', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name: /GitHub/ })).toHaveAttribute(
    'href',
    'https://github.com/rjcuff/woolly',
  );

  const toggle = page.getByRole('button', { name: /switch to (dark|light) mode/i });
  await expect(toggle).toBeInViewport();
  const before = await page.evaluate(() => document.documentElement.classList.contains('dark'));
  await toggle.click();
  await expect.poll(() => page.evaluate(() => document.documentElement.classList.contains('dark'))).toBe(!before);

  await expect(navLink(page, 'Docs')).toBeVisible();
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

test('works with logos and features are shown', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('link', { name: 'Ollama', exact: true })).toHaveAttribute('href', '/docs/ollama');
  await expect(page.getByRole('link', { name: 'Claude Code', exact: true })).toBeVisible();
  await expect(page.getByRole('img', { name: /brain drawn in small squares/ })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Memory you can read' })).toBeAttached();
});

test('download page offers the Mac and Windows apps and install commands', async ({ page }) => {
  await page.goto('/');
  await navLink(page, 'Download').click();
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

test.describe('on Windows', () => {
  test.use({
    userAgent:
      'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36',
  });

  test('home page offers the Windows app', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('link', { name: 'Download for Windows' }).first()).toHaveAttribute(
      'href',
      /Woolly-windows\.exe$/,
    );
  });
});

test.describe('on a Mac', () => {
  test.use({
    userAgent:
      'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15',
  });

  test('home page offers the Mac app', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('link', { name: 'Download for Mac' }).first()).toHaveAttribute(
      'href',
      /Woolly-mac\.dmg$/,
    );
  });
});

test('home shows what shipped lately', async ({ page }) => {
  await page.goto('/');
  const main = page.getByRole('main');
  await expect(page.getByRole('heading', { name: 'Shipping weekly' })).toBeVisible();
  const releases = main.getByRole('list').filter({ hasText: /^v\d/ }).first().getByRole('listitem');
  expect(await releases.count()).toBe(4);
  await expect(releases.first()).toContainText(/v\d+\.\d+\.\d+/);
});
