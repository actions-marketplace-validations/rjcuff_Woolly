import { expect, test } from '@playwright/test';
import { parseChangelog } from '../app/lib/releases';

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
    '## 1.1.0',
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
