import { expect, test } from '@playwright/test';

const productHunt = 'https://www.producthunt.com/products/woolly?utm_source=other&utm_medium=social';

for (const path of ['/', '/docs', '/download']) {
  test(`launch banner is at the top of ${path}`, async ({ page }) => {
    await page.goto(path);
    const banner = page.getByRole('link', { name: /Woolly is live on Product Hunt/ });
    await expect(banner).toBeVisible();
    await expect(banner).toHaveAttribute('href', productHunt);
    await expect(banner).toHaveAttribute('target', '_blank');
    await expect(banner).toHaveAttribute('rel', 'noopener noreferrer');
    const box = await banner.boundingBox();
    expect(box?.y).toBe(0);
  });
}

test('launch banner fits a phone without scrolling sideways', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 800 });
  await page.goto('/');
  await expect(page.getByRole('link', { name: /Woolly is live on Product Hunt/ })).toBeVisible();
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
  expect(overflow).toBeLessThanOrEqual(0);
});
