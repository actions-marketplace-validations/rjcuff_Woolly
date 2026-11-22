# Changelog

## Unreleased

A new look for usewoolly.com, and a page for Woolly Team.

- The website is an ice age picture book now, with a sky hero full of woolly mammoths, a soft serif for headings, light blue cards and one night blue for every action. The homepage has the apps Woolly works with, who it's for, the benchmark numbers, questions and a closing call to action. Download, Privacy, Terms, the docs and the not found page match it.
- Phones get a menu button that opens the links in a drawer.
- Woolly Team has its own page at usewoolly.com/team, with how it works, pricing and questions. team.usewoolly.com sends people who aren't signed in there.
- The favicon, app icons, link previews and brand images are night blue.
- The Product Hunt banner and the Bluesky links are gone.

Safer, steadier memory for the agents you build.

- Memories can be a `preference`, a `procedure` or an `episode`, as well as the project kinds. The MCP `remember` tool takes them for personal memory too, and `recall` takes `type` to get only some kinds. In the memory API, add messages with `memory_type` and search with `filters.type`; results include `type`. The Python and JavaScript SDKs take `type` in `recall` and `search` and `memory_type`/`memoryType` in `add`.
- Woolly learns which memories help. Each recall and session brief counts as a use, and AI apps can rate a memory with the new MCP `rate_memory` tool or `POST /v1/memories/{id}/feedback` (`feedback` in the SDKs). Helpful memories rank a little higher, unhelpful ones a little lower; nothing is hidden or deleted. Counts stay in `usage.json` on your device. Older `usage.json` files still load; going back to an older Woolly resets the counts and votes.
- Search ranks what you said above what an AI guessed, using each memory's confidence, skips near copies of a better result so the top few say different things, and drops weak meaning-only matches. `woolly recall --explain` shows how each result was scored, and API results include `score_parts`.
- The MCP `remember` tool doesn't save a paraphrase of something Woolly already knows, like the same fact with different punctuation. It adds any new tags to the existing memory instead. Word order counts, so "tabs over spaces" never hides "spaces over tabs". Facts with a key still replace older ones as before.
- Search drops a memory that shares only one incidental word with a longer question, and the duplicate checks treat "uses X" and "doesn't use X" as different facts.
- Link related memories, like a decision and the bug behind it, with the MCP `link_memories` tool, `POST /v1/memories/{id}/links` or `link` in the SDKs. Searching for one brings the other along at half its score. Links are a `links:` line in the Markdown header. A new version of a memory keeps its links, forgetting a memory removes links to it, and restoring it puts them back. Only memories with the same user, agent and run can be linked.
- `woolly review --old` lists guesses an AI made that Woolly isn't sure of and nothing has used in 90 days, so you can keep, fix or forget them. `woolly review --keep <id>` keeps one. `woolly doctor` says how many there are, and `GET /api/review-queue` lists them. Nothing expires on its own.
- `woolly mcp --user <id> --agent <name>` makes the MCP server read and write one user's or agent's memories, the way `user_id` and `agent_id` do in the memory API, so an agent you build gets its own memory without the SDK. It can't reach your memories by ID and doesn't read or save handoffs. The default server also can't forget, rate, link or resolve an app user's memory by ID.
- `woolly dream` rolls each week of episodes an AI saved, once they're two weeks old, into one summary, keeping the originals as history. Episodes you saved keep your words. This needs a local chat model.
- `GET /v1/conflicts` lists the conflicts in a scope, each with the memory it disagrees with, and `POST /v1/conflicts/{id}/resolve` keeps one. An app can only keep a version at least as trusted as the ones it replaces; it gets 403 otherwise, and you settle those on the review page. The SDKs have `scoped_conflicts`/`scopedConflicts` and `resolve_scoped`/`resolveScoped`.
- `make eval` runs a fast recall quality check and a table of memory lifecycle cases, with no Ollama needed.

- AI apps can't skip the check for planted instructions by saying a memory came from you. The `remember` tool, `/api/remember` from scripts and the SDKs, and edits through `PUT /v1/memories/{id}` and `PATCH /api/memories/{id}` now refuse text that reads like instructions to an AI, whatever source they give. What you type on the command line or the review page is saved as before.
- When an AI app or a script says a fact came from you, it's saved with source `import` and the tag `user-stated`. It still beats an AI's guess, but it can't replace a fact you saved yourself; it waits as a conflict. Before, an AI app could replace any fact you saved this way.
- An AI app or a script can no longer resolve a conflict in favor of the less trusted memory, or bring back a version that was replaced or retired. `resolve_conflict` and `/api/resolve` return an error; the review page and `woolly resolve` still let you pick either. The Python and JavaScript `resolve` methods raise that error too.
- A fact you saved that went stale still wins over a newer guess from an AI with the same key. The guess waits as a conflict for you to review.
- Editing a fact keeps the code it points at, its type, tags, commit and confidence, unless the edit sets them. Merging facts, including when `woolly dream` merges two, keeps the code and tags of every fact merged, and the type, commit and confidence of the first. Before, the new version lost its sources, so Woolly could no longer flag it when the code changed.
- Memories an app saves for its users with `user_id`, `agent_id` or `run_id` stay out of your own AI apps and your profile: the session brief and its stale facts, the MCP `recall`, `get_profile`, `list_memories` and `why` tools, the proxy, `/api/recall` and `/api/profile`. They are never written to a repo's shared `.woolly/woolly.md`. Before, what an app kept about `alice` could show up in Claude Code or be committed with the team file.
- `GET`, `PUT`, `DELETE /v1/memories/{id}` and `/history` return 404 for a memory outside the `user_id`, `agent_id` or `run_id` in the request.
- Moving a fact's sources to a renamed file updates its time, so sync carries the change to your other devices.
- Forgetting a memory settles any conflict with it, so the other memory is current again instead of waiting forever.
- Restoring a memory from the trash never leaves two current facts for one key. If a newer fact took the key while it was in the trash, it comes back as a conflict.
- Woolly notices every change to the memory folder, including a sync tool that swaps in a file with an older time.
- `woolly doctor` and the new `GET /api/problems` list memory files Woolly couldn't read, and statuses it didn't recognize, instead of skipping them quietly.
- The session brief keeps room for your profile when a project has many notes, and logs only the memories it actually showed.
- The proxy adds at most 2000 characters of memories to a request. The MCP `recall` tool takes `max_chars`, and `list_memories` lists 50 memories unless you pass `limit`. `get_profile` caps its project section.
- `woolly dream` decides which of two disagreeing facts is newer by when each was last stated or confirmed, not when it was first saved.
- Usage and embedding cache files are written through temp files of their own, so the MCP server and `woolly serve` saving at the same time never corrupt them. One may still overwrite the other's latest counts.

## 0.18.0

Pull request checks you can pin, trust with outside text, and read from other tools.

- The pull request comment quotes each stale line and fact as code, so a line in CLAUDE.md can't mention people, load images, add links or render HTML in Woolly's comment. Lines that look like they hold an API key, token or password aren't quoted at all. File names in the comment can't trigger mentions either, like a path under `packages/@acme`.
- The comment says plainly that each finding is a line for a person to review, and that Woolly doesn't check whether the code itself is right.
- `woolly check --json` has a `summary` with facts, files and lines checked, how many went stale in this change and how many already were, and each agent file finding has a `state` of `new` or `before`. Existing fields are unchanged.
- `woolly check --status-json` prints the check as counts and file:line pointers with the commit that changed the code, and no text from agent files, facts or code. It leaves out Claude Code, Codex and claude-mem memory and files git doesn't track, like an ignored CLAUDE.local.md. `--status-file <path>` writes it next to `--markdown`.
- The GitHub Action has a `version` input to pin the Woolly release it runs. It downloads straight from GitHub releases and checks the published checksum, instead of running the install script from the website. `preinstalled` uses a woolly already on the runner.
- The Action has a `base` input, and new outputs: `new-stale`, `already-stale`, `lines-checked` and `status-file`. `lines-checked` fixes `checked` being the only count, which was 0 for repos with agent files but no project facts. The counts come from the status file instead of being read out of the Markdown report.
- The docs pin the Action, list the permissions it needs and why, say where the report goes, and warn against `pull_request_target`.
- CI on GitHub Actions runs the Go tests and runs the Action itself on Linux and macOS against a repo with known stale lines.

## 0.17.0

Woolly checks the memory your agents already have. No setup, no moving your notes.

- `woolly check` now reads CLAUDE.md, AGENTS.md, GEMINI.md, Cursor and Windsurf rules, Copilot instructions, Claude Code's auto memory for the repo, Codex memories that mention it and claude-mem's observations, and flags every line that names a function, file, package script or make target the code no longer has. Git history decides, so only things that were in the code and are gone get flagged, each with the commit that did it. A rename says what the new name is, and a moved file says where it went. It works in any git repo, before `woolly init`. See [Check your agent files](https://usewoolly.com/docs/memory-files).
- `woolly check --fix` writes the new name or path into lines whose function was renamed or file was moved.
- `woolly check` warns when AGENTS.md is over the 32 KB Codex reads, when Claude Code's MEMORY.md is over the 200 lines it loads, and when CLAUDE.md doesn't pull in AGENTS.md.
- Session briefs list out of date lines in those files, so the agent doesn't follow them. The post-commit hook refreshes the list in the background.
- A new MCP tool, `check_memory_files`, so agents can check before trusting a note.
- The GitHub Action reports lines in the repo's agent files that a pull request makes stale, apart from lines that were already out of date, and counts them in its `stale` output.
- The Woolly section in CLAUDE.md and AGENTS.md leaves stale facts out and stays under 16 KB, so every agent reads it whole.
- A new benchmark runs `woolly check` on the agent files of well-known open source repos: `python3 bench/agentfiles/run.py`.
- The installer ends with the pixel mammoth and the three commands to start with. `woolly` on its own shows a header, and the terminal output has color when it's a terminal. Set `NO_COLOR` to turn it off.

## 0.16.0

Woolly shows its work. Ask it why code is the way it is, see what it caught, and let pull requests say which facts they break.

- `woolly why src/auth.ts:42` shows the decisions, conventions and gotchas behind a file, a line or a function. Decisions come first, with the reason and what was rejected, and every fact that points at code is checked against the code as it is now. It also works with `#name` for a function or type. Agents get the same answers from a new MCP `why` tool, and are told to ask before changing code they didn't write. See [Why is this code like this](https://usewoolly.com/docs/why).
- Decisions can point at code. `woolly decide --source-ref src/auth.ts#requireUser` and the `sources` field of the MCP `decide` tool tie a decision to what it's about, so `woolly why` finds it and `woolly check` flags it when that code changes.
- A GitHub Action, `rjcuff/woolly`, comments on a pull request when it makes a project fact stale, with a table of what changed. It updates its own comment on later pushes, writes the report to the job summary, and can fail the job with `fail-on-stale`. Nobody on the team needs Woolly installed. See [Pull request checks](https://usewoolly.com/docs/pull-requests).
- `woolly check --markdown` prints that report, and `woolly check --base <commit>` compares facts that don't record a commit, like ones read from `woolly.md` on a fresh machine, against that commit.
- Caught it. When a check flags a fact, Woolly says how many it has kept away from your agents in that repo. `woolly caught` lists them, and `woolly caught --badge` prints a README badge with the count.
- The session brief now names the facts it kept out and why, instead of only counting them, so the agent knows what not to rely on.
- The next session after a dream opens with what Woolly did while you were away: repeats folded together, facts settled and lessons learned for that project. It shows once per project. `woolly recap` shows it in the terminal, with the mammoth.
- The first `woolly init` brings in what CLAUDE.md, AGENTS.md, GEMINI.md, Cursor and Windsurf rules and Copilot instructions already say, one fact per list item under the heading it fits, then points those facts at the code they mention. It skips code blocks, anything personal, secrets and lines that read like prompt injection, and doesn't change your files. `--no-import` skips it.
- A new website: one blue on ice white, a pixel mammoth, and docs to match.

## 0.15.0

Memory that knows when it's wrong. Project facts now point at the code they're about, and Woolly flags the ones the code no longer backs up.

- Facts can point at the code they're about. A project fact can have a `type` (decision, convention, gotcha, bug or hardware), `sources` (a file, a function or constant in it, or a range of lines, each with a fingerprint of the code) and a `confidence`. Set them with `woolly remember --type hardware --source-ref main/pins.h:12-18`, the MCP `remember` tool, the API or the SDKs. Older memory files load as they are.
- A new `stale` status for facts whose code changed or is gone. Stale facts stay on disk with the reason and the commit, but leave search and session briefs until you review them.
- `.woolly/woolly.md` lines show `src=` for the code a fact points at and `stale=` with the commit that made it stale, so teammates see both in the diff. New Bugs and Hardware headings show up when they have facts; existing files don't change.
- When each memory was last used is kept in `~/.woolly/usage.json`, not in the memory files.
- `woolly check` checks every project fact that points at code against the code as it is now. A renamed or removed function, a deleted file, or a value the fact states that's no longer there marks the fact stale, with the reason and the commit that did it. Code that moved to another file, a renamed file and lines that shifted just move the fact's sources along. Comments and spacing never count. It exits with status 1 when a fact went stale, so it works in CI. Woolly finds definitions in Go, C, C++, Python, JavaScript, TypeScript and Rust, and checks that names still appear in other files. See [Stale facts](https://usewoolly.com/docs/stale-facts).
- The first `woolly check` in a repo adds sources to existing project facts from the files and code names they mention.
- `woolly review` goes through stale facts one at a time, so you can confirm, fix or retire each one. Nothing is deleted.
- `woolly init --hooks` adds a git post-commit hook that checks project memory after every commit, in the background, so the commit returns right away. Existing hooks are kept.
- The session brief leaves stale facts out and tells the agent how many went stale since the last session. It runs a quick check first.
- Facts that agents save with sources are checked against the repo, so a source that doesn't exist is dropped.
- The Python package (`woolly-ai` 0.5.0) and the JavaScript package (`woolly-ai` 0.3.0) can save `type`, `sources` and `confidence` with a fact, and return them along with any stale reason and commit.
- A new site and docs: one quiet column, easier to read, leading with stale detection.
- A stale-fact benchmark, `bench/stale/run.py`. On 50 scripted changes to a small ESP32 firmware repo, Woolly flagged 28 of the 30 facts that went wrong and none of the 20 that didn't, in a median of 35 ms. Every number Woolly publishes is in `bench/RESULTS.md`.

## 0.14.1

- Search by meaning is more accurate. Woolly now gives `nomic-embed-text` the search and document prefixes it was trained with (and the query prefix `mxbai-embed-large` expects). On LOCOMO, the share of questions where Woolly's top 20 results held the answer rose from 75.9% to 83.2% on the first three conversations. Saved embeddings are refreshed once, the first time you search.
- `bench/locomo/run.py --retrieval-only` measures Woolly's search alone, with no answer model: how often the answer is in the top 5, 10 and 20 results, and how often every message it needs is.
- The Mac and Windows apps show the right version. The 0.14.0 apps ran Woolly 0.14.0 but called themselves 0.13.1.
- The favicon is the mammoth again.

## 0.14.0

- A memory layer API for the agents you build, with mem0's routes and field names: `POST /v1/memories` turns a conversation into memories, and `POST /v1/search` finds them. For each fact a local model decides whether it's new, updates an older memory, says one is no longer true, or is already known, and reports `ADD`, `UPDATE`, `DELETE` or `NONE`. See [Memory API for agents](https://usewoolly.com/docs/memory-api).
- Memories can belong to a `user_id`, `agent_id` or `run_id`, so one user's memories are never shown to or changed by another.
- Nothing is overwritten. A replaced memory keeps its text with an `invalid_at` time, a fact that's no longer true is marked `expired`, and `GET /v1/memories/{id}/history` shows every version.
- Your version still wins: a model's change to something you said yourself is saved as a conflict for you to review, not applied.
- The API also accepts the token as `Authorization: Token ...`, the way mem0's clients send it.
- Dreaming: `woolly dream` tidies memory with your local model. It folds repeats together, merges facts that say the same thing, and settles facts that contradict each other, keeping the old ones as history. With `--project` it reads that project's handoffs and decisions and writes up to five lessons for the next session, each pointing at the notes it came from. While Woolly runs, it dreams on its own at most once a day when memory has changed. `WOOLLY_DREAM=off` turns that off. Also `POST /v1/dream`. See [Dreaming](https://usewoolly.com/docs/dreaming).
- `woolly resolve` now settles conflicts that have no key.
- The JavaScript package (`woolly-ai` 0.2.0) and the Python package (`woolly-ai` 0.4.0) have the memory layer too: `add`, `search`, `getAll` / `get_all`, `history`, `deleteAll` / `delete_all` and `dream`, scoped by user, agent or run.

## 0.13.1

- A Windows app. [Download Woolly-windows.exe](https://github.com/rjcuff/woolly/releases/latest/download/Woolly-windows.exe), run it, and Woolly starts in the background with an icon in the system tray. It has the same screens as the Mac app and installs for your account without admin rights. It isn't signed yet, so SmartScreen may ask you to choose More info, then Run anyway.
- The website offers Download for Mac or Download for Windows, depending on your computer.
- Fixed `woolly project edit`, which listed the project memory instead of opening your editor. On Windows it opens Notepad when `EDITOR` isn't set.
- Handoff notes are more accurate: finished work no longer shows up as a decision, and a next step or task the conversation doesn't back up is left out.
- `woolly doctor` shows the repo's project memory and whether decisions will be saved when a session ends.
- `--help` lists options the way the docs write them, like `--key`, and no longer ends with an error. A mistyped command suggests the right one.
- The install script points you to `woolly connect` next.

## 0.13.0

- Project memory in git: `woolly init`. A repo's memory lives in `.woolly/woolly.md`, one fact per line under Overview, Stack, Conventions, Decisions, Gotchas and Commands, so every teammate's agent reads and updates the same notes. Edit the file by hand and Woolly picks it up. Facts with a key replace older ones, lines someone deleted are forgotten, and `.gitattributes` gets `merge=union` so adding facts at the same time doesn't conflict. Personal facts never go in the repo, and lines with secrets or instructions to an AI aren't loaded. See [Project memory](https://usewoolly.com/docs/project-memory).
- Decisions are saved when a Claude Code session ends, in repos with `woolly.md`, using your local model. Turn it off with `woolly init --no-capture`.
- `woolly init --export claude,agents` keeps a Project memory section in CLAUDE.md and AGENTS.md, between Woolly's markers, for tools without MCP.
- `woolly project` shows the project's memory by heading, and `woolly project edit` opens it in your editor.
- The MCP `remember` and `recall` tools take a `scope`: `personal`, `project` or `auto`. `remember` also takes a `section`. Auto picks personal when unsure, and Woolly refuses to put a fact that reads like it's about you in shared project memory.
- Repos shared with `woolly share` move to `woolly.md` the first time they sync. `woolly share` still works and does the same as `woolly init`.

- JavaScript and TypeScript package: `npm install woolly-ai`. Save, recall, learn and add memory to prompts from Node apps, plus `woollyMemory()` middleware for the Vercel AI SDK that adds the user's relevant memories to every call. See [JavaScript](https://usewoolly.com/docs/javascript).
- A LOCOMO benchmark you can run on your own computer with Ollama: `python3 bench/locomo/run.py`. It reports answer accuracy and, separately, how often Woolly's search found the message with the answer. See [Benchmark](https://usewoolly.com/docs/benchmark).

## 0.12.0

- Import is smarter with Ollama. When the learning model is installed, Woolly has it read your recent messages about yourself and combines what it finds with the simple rules. In testing it found every fact in a sample export, up from about half.

- Import from ChatGPT and Claude in the app. Drop your data export into Memories → Import, review the facts Woolly found, uncheck what you don't want, and save. It works out which app the export came from, marks what Woolly already knows, and labels imported memories with where they came from. Also in the API: `POST /api/import/preview` and `/api/import/save`.
- Exports saved without a `.zip` name are now read as zips.

- Decisions. `woolly decide` and the new `decide` MCP tool save a choice with the reason and the options that were turned down. Sessions start with an Already decided list, and agents are told not to suggest a rejected option again. The Memories tab can filter to decisions.
- Notes that know when they're out of date. Project notes saved in a git repo record the commit, and the startup context marks a note when files it mentions changed since.

## 0.11.0

- The Mac app is signed and notarized by Apple, so it opens with a double-click. [Download it](https://github.com/rjcuff/woolly/releases/latest/download/Woolly-mac.dmg). One download works on Apple Silicon and Intel.

- Pick up where you left off. When Claude Code compacts a conversation or a session ends, Woolly saves a handoff note for the project: what you were working on, what got done, what was decided and what's next. The next session starts with it, in Claude Code or any connected app. A quick note is saved instantly, then a local model writes a better one in the background when Ollama is running. `woolly handoff` shows it, and MCP apps can save one with the new `handoff` tool. Run `woolly connect claude-code` again to add the new hooks.
- Share project memory with your team. `woolly share` keeps a repo's project notes in `.woolly/memories` as plain files to commit. Teammates' AI tools start with the same notes, and new notes are written back as they're saved. Only project notes are shared, never personal ones, and notes from a repo are checked for secrets and injected instructions. `--agents-md` also keeps a Project notes section in AGENTS.md for tools without MCP.

## 0.10.0

- The Woolly app. `woolly ui` opens a new app with Home (your memories at a glance, with a chart of what apps saved each day), Memories (add, search, edit in place, forget), Apps (connect Claude, Cursor, VS Code and more with one click) and Settings (Ollama, sync, conflicts, history, trash and activity).
- A Mac app in preview wraps the same screens, with a menu bar icon. Download `Woolly-mac.dmg` from the release. It isn't signed yet, so the first time you open it, go to System Settings → Privacy & Security and choose Open Anyway.
- `woolly learn` no longer saves anything from a message that only asks questions.
- New API endpoints for the app: `GET /api/apps`, `POST /api/apps/{id}/connect` and `/disconnect`, `GET /api/settings`, and `POST /api/sync/init`, `/join`, `/now`, `/leave`.

## 0.9.0

- Smart memory. `woolly learn` uses a model running in Ollama to pick out facts from a conversation or any text. A fact on the same subject as an older one replaces it, like a new city or a new editor, and repeats are skipped. Facts are saved as things an AI figured out, so they never overrule what you said yourself, and secrets and instructions are still refused. Also in the API (`POST /api/learn`) and Python (`w.learn(...)`). The default model is `qwen2.5:3b`; set `WOOLLY_LLM_MODEL` to change it. See [Smart memory](https://usewoolly.com/docs/learn).
- The proxy uses the model to save facts from your chats when it's available.
- Sync between your devices. `woolly sync init` and `woolly sync join` link your laptop, desktop or Raspberry Pi over your own network, with no cloud. Every message is encrypted and signed with a key only your devices have. The newest change wins, and memories deleted for good stay deleted everywhere. `woolly serve` and the background service sync every 5 minutes. See [Sync your devices](https://usewoolly.com/docs/sync).
- Builds for 32-bit ARM, for Raspberry Pi OS 32-bit. The install script picks it automatically.
- `woolly doctor` shows whether learning and sync are ready.
- Forgetting and restoring a memory now update its time, so sync knows it's the newest change.
- Python package 0.3.0 adds `learn()`.

## 0.8.0

- Claude Desktop can install Woolly with a double-click: download `woolly.mcpb` from the latest release. It holds Woolly for macOS, Windows and Linux.
- Woolly is listed on the official MCP Registry as `io.github.rjcuff/woolly`.
- Projects. Facts about a codebase can belong to that project, and facts about you stay global. Apps working in a git repo see your memories plus that project's, and nothing from other projects. Keys only clash within a project. Use `--project` on `remember`, `recall` and `list`, `project` in the API and Python package, and the project picker on the review page. See [Projects](https://usewoolly.com/docs/projects).
- The profile only lists facts about you. Project notes are shown separately.

## 0.7.0

- Memory that loads by itself. `woolly connect` adds a startup hook to Claude Code, so each session starts with your profile and the notes about the project you're in. `woolly context` prints what it loads. `woolly disconnect` removes the hook.
- Every memory records the app that saved it, like Claude Code, Cursor, the proxy or the command line. The review page shows it as "via Cursor".
- `woolly log` shows what each app saved, read and deleted. It's also on the new **Activity** tab of the review page and at `GET /api/activity`. The log stays in `~/.woolly/activity.log`.
- Apps using the local API can name themselves with the `X-Woolly-Client` header. The Python package sends "Python app" by default, or the `app` you pass to `Woolly()`.
- `woolly doctor` says whether Claude Code loads your profile at startup.

## 0.6.0

- `woolly connect` adds Woolly to Claude Code, Claude Desktop, Cursor, VS Code, Windsurf, Codex, Gemini CLI, LM Studio and opencode in one step. It keeps the rest of each config and saves a backup first. `woolly disconnect` undoes it.
- `woolly service install` runs Woolly in the background and starts it at login, like Ollama. It uses a LaunchAgent on macOS, a systemd user service on Linux and a startup entry on Windows, none of which need admin rights.
- `woolly serve --proxy` runs the API, review page and proxy in one process. `--log` writes output to a file.
- `woolly doctor` shows the background service and which apps are connected.
- `woolly update` updates Woolly to the latest version and verifies the download. It's the only command that contacts the internet, and only when you run it.
- The website shows the GitHub star count in the nav once it passes 100.

## 0.5.0

- New Python package: `pip install woolly-ai`. Save, recall and add memories to prompts from your own apps. See [Python](https://usewoolly.com/docs/python).
- New README and docs introduction.

## 0.4.0

- `woolly proxy` gives memory to Ollama and to apps that use the OpenAI or Anthropic API. It adds relevant memories to each chat and saves facts you state.
- Windows install script: `irm https://usewoolly.com/install.ps1 | iex`.
- Homebrew: `brew install rjcuff/tap/woolly`.

## 0.3.0

### Security

- Text from an AI or an import that reads like instructions to an AI is refused, so a malicious web page can't plant commands in your memory.
- Recalled memories are labeled as notes, not instructions, when sent to AI apps.
- The local API now requires a secret token stored in `~/.woolly/token`. `woolly ui` handles it for you.
- The local API rejects requests that browsers mark as coming from another site.
- The install script verifies downloads against the release checksums.
- Imports refuse files larger than 2 GB.

### New

- Forgetting a memory moves it to the trash. Use `woolly trash` to see it, `woolly restore <id>` to bring it back, and `woolly trash --empty` to delete for good.
- The review page has a Trash tab and an Undo button after you forget something.
- New Security page in the docs, and SECURITY.md.

### Changed

- Docs use a simpler layout with a top bar, a plain FAQ, and no numbered steps.
- `go install` builds now report the right version.
- The install script creates its install folder if it doesn't exist.

## 0.2.1

- The project moved to github.com/rjcuff/woolly and usewoolly.com. Old links redirect.

## 0.2.0

- `woolly ui` opens a review page to add, search, edit and forget memories, resolve conflicts and copy your profile.
- `woolly profile` and the `get_profile` tool give a short summary of what Woolly knows about you.
- Passwords, API keys and other secrets are refused.

## 0.1.0

- First release: remember, recall, list, forget, conflicts and resolve.
- `woolly import` for ChatGPT and Claude exports.
- `woolly mcp` for Claude Code, Claude Desktop, Cursor and other MCP apps.
- `woolly serve` for a local HTTP API.
- Optional search by meaning with Ollama.
