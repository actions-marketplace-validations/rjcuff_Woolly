import { expect, test } from '@playwright/test';
import { parseChangelog, published } from '../app/lib/releases';

test('the changelog becomes a short plain summary per release', () => {
  const md = [
    '# Changelog',
    '',
    '## Unreleased',
    '',
    '- Not out yet.',
    '',
    '## 1.2.0',
    '',
    '- A new [`woolly init`](https://example.com) command. It does more too. And one more thing that is long enough.',
    '- Something else.',
    '',
    '## 1.1.0 (2026-11-01)',
    '',
    '- **Bold** start without a period',
  ].join('\n');
  expect(parseChangelog(md, 4)).toEqual([
    {
      version: '1.2.0',
      summary: 'A new woolly init command. It does more too. And one more thing that is long enough.',
      href: 'https://github.com/rjcuff/woolly/releases/tag/v1.2.0',
    },
    {
      version: '1.1.0',
      summary: 'Bold start without a period',
      href: 'https://github.com/rjcuff/woolly/releases/tag/v1.1.0',
    },
  ]);
  expect(parseChangelog(md, 1)).toHaveLength(1);
});

test('a file name with a dot does not end the sentence', () => {
  const md = ['## 2.0.0', '', '- A Windows app. [Download Woolly-windows.exe](https://example.com), run it, and Woolly starts.'].join('\n');
  expect(parseChangelog(md, 1)[0].summary).toBe(
    'A Windows app. Download Woolly-windows.exe, run it, and Woolly starts.',
  );
});

test('a version GitHub has not released yet is left out', () => {
  const md = ['## 2.0.0', '', '- Not out yet.', '', '## 1.0.0', '', '- Out.'].join('\n');
  const parsed = parseChangelog(md, 5);
  const dates = new Map([['1.0.0', '2026-11-01T00:00:00Z']]);
  expect(published(parsed, dates, 4).map((r) => r.version)).toEqual(['1.0.0']);
  // When GitHub can't be reached, nothing is hidden.
  expect(published(parsed, new Map(), 4).map((r) => r.version)).toEqual(['2.0.0', '1.0.0']);
});
