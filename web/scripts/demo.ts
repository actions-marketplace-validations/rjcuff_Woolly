// Records the README demo GIF from real Woolly output.
//
//   make demo
//
// It runs each command with the woolly binary in a throwaway folder, draws a
// terminal frame by frame in a headless browser, and hands the frames to
// scripts/gifenc to build docs/demo.gif.
import { execFileSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { chromium } from '@playwright/test';

const root = resolve(import.meta.dirname, '../..');
const bin = join(root, 'bin/woolly');
const home = mkdtempSync(join(tmpdir(), 'woolly-demo-'));
const frames = mkdtempSync(join(tmpdir(), 'woolly-frames-'));

// A step runs one woolly command. `shown` is the command line drawn on
// screen; leave it out for something an app does on its own, and describe
// that with `note` instead. Every line of output is real.
type Step = {
  shown?: string;
  note?: string;
  args: string[];
  input?: string;
  cwd?: string;
  pause: number;
  clear?: boolean;
  tidy?: (out: string) => string;
};

// A drone firmware repo whose CLAUDE.md was right when it was written. Then a
// refactor renamed a function, moved a file and renamed a script, and
// nobody touched CLAUDE.md.
const repo = join(home, 'code', 'drone');
mkdirSync(join(repo, 'main'), { recursive: true });
mkdirSync(join(repo, 'src', 'control'), { recursive: true });
const git = (...args: string[]) =>
  execFileSync('git', args, {
    cwd: repo,
    env: { ...process.env, GIT_AUTHOR_NAME: 'demo', GIT_AUTHOR_EMAIL: 'demo@example.com', GIT_COMMITTER_NAME: 'demo', GIT_COMMITTER_EMAIL: 'demo@example.com' },
    stdio: 'pipe',
  });
const put = (name: string, text: string) => writeFileSync(join(repo, name), text);
git('init', '-q', '-b', 'main');
put('main/imu.c', 'int imu_init(void) { return 0; }\nint imu_read_accel(int16_t *out) { return read_raw(out); }\n');
put('src/flight.ts', 'export const mixer = 1;\n');
put('package.json', '{"scripts": {"lint": "eslint .", "dev": "vite"}}\n');
put(
  'CLAUDE.md',
  [
    '# Drone firmware',
    '',
    '- Call imu_init before anything else',
    '- imu_read_accel returns raw counts; divide by 16384 for g',
    '- Flight control lives in src/flight.ts',
    '- Run `npm run lint` before you push',
    '- Unplug the battery before flashing',
    '',
  ].join('\n'),
);
git('add', '.');
git('commit', '-q', '-m', 'first flight');
put('main/imu.c', 'int imu_init(void) { return 0; }\nint imu_read_accel_g(float *out) { return read_scaled(out); }\n');
put('package.json', '{"scripts": {"check": "eslint .", "dev": "vite"}}\n');
git('mv', 'src/flight.ts', 'src/control/flight.ts');
git('add', '.');
git('commit', '-q', '-m', 'refactor imu and tooling');

// Commit hashes and dates make lines long; drop the dates so the demo reads at a glance.
const short = (out: string) => out.replace(/ \(\d{4}-\d{2}-\d{2}\)/g, '').replace(/^.*woolly init --hooks.*\n?/m, '');

const steps: Step[] = [
  { note: '# A refactor renamed a function, moved a file and renamed a script.', args: ['version'], pause: 60, tidy: () => '' },
  { shown: 'woolly check', args: ['check'], cwd: repo, pause: 420, tidy: short },
  { shown: 'woolly check --fix', args: ['check', '--fix'], cwd: repo, pause: 450, clear: true, tidy: short },
];

function run(step: Step): string {
  try {
    return execFileSync(bin, step.args, {
      env: { ...process.env, HOME: home, WOOLLY_HOME: join(home, '.woolly'), WOOLLY_NO_EMBED: '1', NO_COLOR: '1' },
      encoding: 'utf8',
      cwd: step.cwd ?? home,
      input: step.input ?? '',
      stdio: ['pipe', 'pipe', 'pipe'],
    });
  } catch (e) {
    const err = e as { stdout?: string; stderr?: string };
    return (err.stdout ?? '') + (err.stderr ?? '');
  }
}

const escape = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

function colorize(line: string): string {
  const l = escape(line);
  if (l.startsWith('fixed ')) return `<span class="green">fixed</span>${l.slice('fixed'.length)}`;
  if (l.startsWith('Checked ')) return `<span class="bold">Checked</span>${l.slice('Checked'.length)}`;
  if (/names? code that's gone/.test(l)) return `<span class="red">${l}</span>`;
  if (l.startsWith('Every function')) return `<span class="green">${l}</span>`;
  const where = l.match(/^(  )(\S+:\d+)(\s+)(.*)$/);
  if (where) return `${where[1]}<span class="red">${where[2]}</span>${where[3]}${where[4]}`;
  if (/^\s{4,}\S/.test(l)) return `<span class="muted">${l}</span>`;
  if (l.startsWith('woolly: not saved')) return `<span class="red">${l}</span>`;
  if (l.startsWith('Replaced older memory')) return `<span class="amber">${l}</span>`;
  if (l.startsWith('Remembered:')) return `<span class="green">Remembered:</span>${l.slice('Remembered:'.length)}`;
  if (l.startsWith('#')) return `<span class="bold">${l}</span>`;
  const app = l.match(/^(Claude Code|Cursor|command line)(\s+)(.*)$/);
  if (app) return `<span class="green">${app[1]}</span>${app[2]}${app[3]}`;
  return l;
}

const page = (body: string) => `<!doctype html><html><head><style>
  body { margin: 0; background: #0e0e0e; }
  .win { width: 1000px; height: 560px; background: #0e0e0e; color: #f2f2f2; box-sizing: border-box;
         font: 17px/1.55 "SF Mono", Menlo, monospace; padding: 0; overflow: hidden; }
  .bar { height: 36px; background: #2a2a2a; display: flex; align-items: center; gap: 8px; padding: 0 14px; }
  .dot { width: 12px; height: 12px; border-radius: 50%; background: #8a8a8a; }
  .term { padding: 18px 24px; white-space: pre-wrap; }
  .p { color: #6cc28f; } .green { color: #6cc28f; } .red { color: #f07167; } .amber { color: #e0b05a; }
  .muted { color: #8a8a8a; } .bold { font-weight: 700; }
  .cursor { display: inline-block; width: 10px; height: 20px; background: #f2f2f2; vertical-align: -4px; }
</style></head><body><div class="win"><div class="bar"><span class="dot"></span><span class="dot"></span><span class="dot"></span></div><div class="term">${body}</div></div></body></html>`;

async function main() {
  const browser = await chromium.launch();
  const tab = await browser.newPage({ viewport: { width: 1000, height: 560 }, deviceScaleFactor: 1 });
  const durations: number[] = [];
  let n = 0;
  let history: string[] = [];

  const shot = async (body: string, cs: number) => {
    await tab.setContent(page(body));
    await tab.locator('.win').screenshot({ path: join(frames, `${String(n++).padStart(4, '0')}.png`) });
    durations.push(cs);
  };
  // Keep only the lines that fit on screen.
  const view = (lines: string[]) => lines.slice(-17).join('\n');

  await shot(view(['<span class="p">$</span> <span class="cursor"></span>']), 80);
  for (const step of steps) {
    if (step.clear) history = [];
    const output = (step.tidy ?? ((o) => o))(run(step)).trimEnd();
    if (step.note) {
      history.push(`<span class="muted">${escape(step.note)}</span>`, '');
      await shot(view([...history]), 90);
    }
    if (step.shown) {
      for (let i = 2; i <= step.shown.length + 1; i += 2) {
        const typed = escape(step.shown.slice(0, i));
        await shot(view([...history, `<span class="p">$</span> ${typed}<span class="cursor"></span>`]), 4);
      }
      history.push(`<span class="p">$</span> ${escape(step.shown)}`);
    }
    if (output) history.push(...output.split('\n').map(colorize), '');
    await shot(view([...history, '<span class="p">$</span> <span class="cursor"></span>']), step.pause);
  }

  await browser.close();
  writeFileSync(join(frames, 'durations.json'), JSON.stringify(durations));
  const out = join(root, 'docs/demo.gif');
  mkdirSync(join(root, 'docs'), { recursive: true });
  execFileSync('go', ['run', './scripts/gifenc', frames, out], { cwd: root, stdio: 'inherit' });
  rmSync(home, { recursive: true, force: true });
  // DEMO_KEEP_FRAMES=1 keeps the frames to look at.
  if (process.env.DEMO_KEEP_FRAMES) console.log('frames in', frames);
  else rmSync(frames, { recursive: true, force: true });
}

await main();
