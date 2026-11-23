---
name: woolly-check
description: Check this repo's agent memory (CLAUDE.md, AGENTS.md, rules files and Woolly facts) for lines that no longer match the code. Use when the user asks if their CLAUDE.md or memory is out of date, or after a big refactor.
---

Run `woolly check` from the repository root and read its output.

- If it reports nothing stale, say so in one line.
- If it reports stale lines, list each one with its file and line number, what changed in the code, and the commit that changed it when Woolly names one.
- If a line points at something that was renamed, offer to run `woolly check --fix`, which rewrites renamed names. Don't run it without asking.
- If the `woolly` command isn't found, tell the user to install it from https://usewoolly.com/docs/quickstart.
