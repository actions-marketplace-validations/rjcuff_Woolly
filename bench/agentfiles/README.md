# Agent file benchmark

How often do the notes real repos keep for their coding agents name code that's gone? And when Woolly says a line is out of date, is it right?

## Run it

```sh
python3 bench/agentfiles/run.py
```

It needs git, a network the first time, and Go (it builds `woolly` from this checkout, or pass `--woolly path/to/woolly`). It clones every repo in [repos.txt](repos.txt) with full history into a temporary folder (`--cache` to keep them somewhere else), runs `woolly check --json` on each with an empty Woolly folder and no access to your own Claude Code, Codex or claude-mem memory, and saves every finding to `results/<date>.json`. Big repos take a while to clone the first time.

## How it works

For each repo, `woolly check` reads its CLAUDE.md, AGENTS.md, GEMINI.md, Cursor and Windsurf rules and Copilot instructions, including AGENTS.md and CLAUDE.md files in subfolders. For every line, it looks for the code names, paths, package scripts and make targets it mentions, then asks git two questions:

1. Is it in the code now? Markdown doesn't count as code, and a file name counts for its name.
2. If not, which commit took it out? Only history since a month before the agent files were written is searched.

A line is flagged only when both say it's gone. See [Check your agent files](https://usewoolly.com/docs/memory-files) for the full rules.

## Reading the results

The script can't tell a real miss from a false alarm, so every finding is read against the repo at the commit it was run on, and the verdict goes in [verdicts.json](verdicts.json): `right` when the line names something the code really doesn't have anymore, `wrong` otherwise, with a note and who checked it. Precision is right / (right + wrong). It doesn't measure recall: lines that are wrong in ways Woolly doesn't look for, like a changed default, aren't counted.

The list of repos is well-known open source projects that keep agent files. It isn't a random sample, and these maintainers clearly keep their files up to date by hand, so expect more stale lines in a typical repo.
