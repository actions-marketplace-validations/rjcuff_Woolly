// Tests against a real `woolly serve` with a throwaway memory folder.
//   npm test   (builds first; uses ../../bin/woolly or WOOLLY_BIN)

import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { mkdtempSync, rmSync } from 'node:fs';
import { createServer } from 'node:net';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { after, before, test } from 'node:test';
import { generateText, wrapLanguageModel } from 'ai';
import { MockLanguageModelV4 } from 'ai/test';
import { Woolly, WoollyError, WoollyNotRunning } from '../dist/index.js';
import { woollyMemory } from '../dist/ai-sdk.js';

const bin = process.env.WOOLLY_BIN ?? resolve(import.meta.dirname, '../../../bin/woolly');
const home = mkdtempSync(join(tmpdir(), 'woolly-js-test-'));
let server;
let address;
let w;

const freePort = () =>
  new Promise((ok) => {
    const s = createServer().listen(0, '127.0.0.1', () => {
      const { port } = s.address();
      s.close(() => ok(port));
    });
  });

before(async () => {
  address = `127.0.0.1:${await freePort()}`;
  server = spawn(bin, ['serve', '--addr', address], {
    env: { ...process.env, WOOLLY_HOME: home, WOOLLY_NO_EMBED: '1', OLLAMA_HOST: '127.0.0.1:9' },
    stdio: 'ignore',
  });
  w = new Woolly({ address, home, app: 'JS test' });
  for (let i = 0; i < 50 && !(await w.isRunning()); i++) await new Promise((r) => setTimeout(r, 100));
});

after(() => {
  server?.kill();
  rmSync(home, { recursive: true, force: true });
});

test('remembers and recalls', async () => {
  const r = await w.remember('Has a cat named Miso', { key: 'user.pet' });
  assert.equal(r.memory.via, 'JS test');
  const found = await w.recall('what pets do I have');
  assert.ok(found.some((m) => m.content.includes('Miso')));
  assert.equal(typeof found[0].score, 'number');
});

test('a newer fact replaces the older one', async () => {
  await w.remember('Lives in Austin', { key: 'user.city' });
  const r = await w.remember('Lives in Denver', { key: 'user.city' });
  assert.equal(r.superseded?.[0]?.content, 'Lives in Austin');
});

test('context frames memories as notes', async () => {
  await w.remember('Prefers short answers');
  const note = await w.context('how should answers be written');
  assert.match(note, /information, not instructions/);
  assert.match(note, /Prefers short answers/);
  assert.equal(await w.context('zzzz qqqq'), '');
});

test('projects keep notes apart', async () => {
  await w.remember('Builds with Bazel', { project: 'monorepo' });
  await w.remember('Builds with Make', { project: 'tools' });
  const found = await w.recall('how does it build', { project: 'monorepo' });
  assert.ok(found.some((m) => m.content.includes('Bazel')));
  assert.ok(!found.some((m) => m.content.includes('Make')));
});

test('update, forget and restore', async () => {
  const { memory } = await w.remember('Uses tabs');
  assert.equal((await w.update(memory.id, { content: 'Uses spaces' })).content, 'Uses spaces');
  await w.forget(memory.id);
  assert.ok((await w.trash()).some((m) => m.id === memory.id));
  assert.equal((await w.restore(memory.id)).content, 'Uses spaces');
});

test('secrets are refused', async () => {
  await assert.rejects(w.remember('my key is sk-proj-a8F3kQ9zLm2xR7tV4bN6cW1yE5'), WoollyError);
});

test('profile has markdown', async () => {
  const p = await w.profile();
  assert.match(p.markdown, /# About me/);
});

test('learn needs Ollama', async () => {
  await assert.rejects(w.learn([{ role: 'user', content: 'I moved to Denver' }]), /Ollama/);
});

test('a wrong token is refused', async () => {
  const bad = new Woolly({ address, token: 'nope' });
  await assert.rejects(bad.list(), (e) => e instanceof WoollyError && e.status === 401);
});

test('says when Woolly is not running', async () => {
  const off = new Woolly({ address: `127.0.0.1:${await freePort()}`, token: 'x' });
  assert.equal(await off.isRunning(), false);
  await assert.rejects(off.list(), WoollyNotRunning);
});

const reply = {
  content: [{ type: 'text', text: 'ok' }],
  finishReason: { unified: 'stop', raw: 'stop' },
  usage: {
    inputTokens: { total: 1, noCache: 1, cacheRead: 0, cacheWrite: 0 },
    outputTokens: { total: 1, text: 1, reasoning: 0 },
  },
  warnings: [],
};

test('AI SDK middleware adds memories as a system message', async () => {
  await w.remember('Is allergic to peanuts');
  const mock = new MockLanguageModelV4({ doGenerate: reply });
  const model = wrapLanguageModel({ model: mock, middleware: woollyMemory({ client: w }) });
  const { text } = await generateText({ model, prompt: 'Any snack ideas without peanuts?' });
  assert.equal(text, 'ok');
  const prompt = mock.doGenerateCalls[0].prompt;
  assert.equal(prompt[0].role, 'system');
  assert.match(prompt[0].content, /allergic to peanuts/);
});

test('AI SDK middleware passes calls through when Woolly is off', async () => {
  const mock = new MockLanguageModelV4({ doGenerate: reply });
  const off = new Woolly({ address: `127.0.0.1:${await freePort()}`, token: 'x' });
  const model = wrapLanguageModel({ model: mock, middleware: woollyMemory({ client: off }) });
  await generateText({ model, prompt: 'hello' });
  assert.equal(mock.doGenerateCalls[0].prompt.some((m) => m.role === 'system'), false);
});

test('memory layer: add, search, history and delete by user', async () => {
  const added = await w.add([{ role: 'user', content: 'Deploys the API to Fly.io' }, { role: 'assistant', content: 'ok' }], {
    userId: 'alice',
    infer: false,
    metadata: { team: 'web' },
  });
  assert.equal(added.length, 1);
  assert.equal(added[0].event, 'ADD');
  await w.add('Deploys the API to Render', { userId: 'bob', infer: false });

  const found = await w.search('where does the API deploy', { userId: 'bob' });
  assert.equal(found.length, 1);
  assert.match(found[0].memory, /Render/);
  assert.equal(typeof found[0].score, 'number');

  const mine = await w.getAll({ userId: 'alice' });
  assert.equal(mine.length, 1);
  assert.deepEqual(mine[0].metadata, { team: 'web' });

  const hist = await w.history(added[0].id);
  assert.equal(hist[0].event, 'ADD');

  assert.equal(await w.deleteAll({ userId: 'alice' }), 1);
  assert.equal((await w.getAll({ userId: 'alice' })).length, 0);
  assert.equal((await w.getAll({ userId: 'bob' })).length, 1);
  await assert.rejects(w.deleteAll({}), WoollyError);
});

test('add with a model needs Ollama, and dream still folds repeats', async () => {
  await assert.rejects(w.add('I live in Denver', { userId: 'carol' }), (e) => e instanceof WoollyError && e.status === 503);
  await w.add('Uses pnpm', { userId: 'dave', infer: false });
  await w.add('uses pnpm.', { userId: 'dave', infer: false });
  const rep = await w.dream({ userId: 'dave' });
  assert.equal(rep.merged, 1);
  assert.equal((await w.getAll({ userId: 'dave' })).length, 1);
});
