import { expect, test } from '@playwright/test';

function decodeConfig(href: string): unknown {
  const config = new URL(href).searchParams.get('config');
  if (!config) throw new Error(`no config in ${href}`);
  return JSON.parse(Buffer.from(config, 'base64').toString('utf8'));
}

for (const path of ['/download', '/docs/connect']) {
  test(`${path} has one-click install links for Cursor and LM Studio`, async ({ page }) => {
    await page.goto(path);
    const want = { command: 'woolly', args: ['mcp'] };

    const cursor = page.getByRole('link', { name: 'Add to Cursor' });
    const cursorHref = await cursor.getAttribute('href');
    expect(cursorHref).toMatch(/^cursor:\/\/anysphere\.cursor-deeplink\/mcp\/install\?name=woolly&config=/);
    expect(decodeConfig(cursorHref!)).toEqual(want);

    const lm = page.getByRole('link', { name: 'Add to LM Studio' });
    const lmHref = await lm.getAttribute('href');
    expect(lmHref).toMatch(/^lmstudio:\/\/add_mcp\?name=woolly&config=/);
    expect(decodeConfig(lmHref!)).toEqual(want);
  });
}
