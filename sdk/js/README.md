# woolly-ai

JavaScript and TypeScript client for [Woolly](https://usewoolly.com), one local memory shared by all your AI tools. Everything stays on your computer.

```sh
npm install woolly-ai
```

You need Woolly running: [download the Mac app](https://usewoolly.com/download) or run `woolly serve`.

```ts
import { Woolly } from 'woolly-ai';

const w = new Woolly({ app: 'My app' });

await w.remember('Moved to Denver in October', { key: 'user.location' });
const found = await w.recall('where do I live?');

// Add what Woolly knows to any prompt
const system = `You are a helpful assistant.\n\n${await w.context(userMessage)}`;
```

## Vercel AI SDK

Wrap any model, and every call gets the user's relevant memories:

```ts
import { generateText, wrapLanguageModel } from 'ai';
import { woollyMemory } from 'woolly-ai/ai-sdk';

const model = wrapLanguageModel({ model: yourModel, middleware: woollyMemory() });
const { text } = await generateText({ model, prompt: 'Any snack ideas?' });
```

If Woolly isn't running, calls go through unchanged.

## Memory for the agents you build

The same API as mem0, on your own machine. `add` turns a conversation into memories for one user, and a local model decides whether each fact is new, updates an older one, or is no longer true. `search` finds them before your agent answers.

```ts
await w.add(
  [
    { role: 'user', content: 'I just moved to Austin' },
    { role: 'assistant', content: 'Nice!' },
  ],
  { userId: 'alice' },
);
// [{ memory: 'Lives in Austin', event: 'UPDATE', previous_memory: 'Lives in Denver', ... }]

const memories = await w.search('where does she live?', { userId: 'alice' });
```

Scope with `userId`, `agentId`, `runId` or `project`. Also `getAll`, `history` (every version of a memory), `deleteAll` and `dream`, which merges duplicates and settles facts that changed. Pass `infer: false` to `add` to save messages as they are, without a model. See [Memory API for agents](https://usewoolly.com/docs/memory-api).

## Learn from a conversation

With Ollama running, a local model picks out the facts worth keeping:

```ts
await w.learn([
  { role: 'user', content: 'I just moved to Denver. Keep answers short.' },
  { role: 'assistant', content: 'Got it!' },
]);
```

## Everything else

`list`, `get`, `update`, `forget`, `restore`, `trash`, `conflicts`, `resolve` and `profile`. Pass `project` to `remember`, `recall` and `context` to keep notes about one codebase separate.

`new Woolly()` finds the server at `WOOLLY_HOST` (default `127.0.0.1:11435`) and reads the token from `WOOLLY_HOME` (default `~/.woolly`). The `app` you pass is shown on memories your app saves and in `woolly log`.

Requires Node 18 or newer. MIT license.
