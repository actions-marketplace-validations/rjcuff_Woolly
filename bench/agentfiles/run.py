"""Agent file benchmark for Woolly.

    python3 bench/agentfiles/run.py

Clones well-known open source repos that keep notes for coding agents
(CLAUDE.md, AGENTS.md, GEMINI.md, Cursor rules, Copilot instructions) and runs
`woolly check` on each. It counts the lines that name a function, file or
script the code no longer has, and saves every one of them so a person can
read through them and mark each right or wrong.

Each repo runs with its own empty Woolly folder and no access to your Claude
Code, Codex or claude-mem memory. Nothing touches your real memories, and
nothing needs Ollama or an API key.

Needs git, a network the first time, and Go (it builds woolly from this
checkout), or pass --woolly.
"""

import argparse
import json
import os
import statistics
import subprocess
import sys
import tempfile
import time
from datetime import date
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent


def build(out: Path) -> Path:
    exe = out / "woolly"
    subprocess.run(["go", "build", "-o", str(exe), "./cmd/woolly"], cwd=ROOT, check=True)
    return exe


def clone(name: str, cache: Path) -> Path | None:
    dest = cache / name.replace("/", "_")
    if not dest.exists():
        r = subprocess.run(
            # Agent files are from 2025 on, so older history isn't needed.
            ["git", "clone", "-q", "--single-branch", "--shallow-since=2024-10-01", f"https://github.com/{name}.git", str(dest)],
            capture_output=True,
        )
        if r.returncode != 0:
            return None
    return dest


def check(exe: Path, repo: Path, scratch: Path) -> tuple[dict | None, float]:
    env = dict(os.environ)
    for k in ("WOOLLY_HOME", "CLAUDE_CONFIG_DIR", "CODEX_HOME", "HOME"):
        d = scratch / k.lower()
        d.mkdir(parents=True, exist_ok=True)
        env[k] = str(d)
    env["WOOLLY_NO_EMBED"] = "1"
    start = time.perf_counter()
    r = subprocess.run(
        [str(exe), "check", "--json", "--no-backfill", "--dir", str(repo)],
        capture_output=True, text=True, env=env, timeout=600,
    )
    took = time.perf_counter() - start
    try:
        return json.loads(r.stdout), took
    except json.JSONDecodeError:
        print(r.stdout, r.stderr, file=sys.stderr)
        return None, took


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--woolly", help="a woolly binary to use instead of building one")
    ap.add_argument("--cache", default=str(Path(tempfile.gettempdir()) / "woolly-agentfiles"), help="where to keep the clones")
    ap.add_argument("--repos", default=str(HERE / "repos.txt"))
    ap.add_argument("--out", default=str(HERE / "results" / f"{date.today().isoformat()}.json"))
    args = ap.parse_args()

    cache = Path(args.cache)
    cache.mkdir(parents=True, exist_ok=True)
    names = [l.strip() for l in Path(args.repos).read_text().splitlines() if l.strip() and not l.startswith("#")]

    with tempfile.TemporaryDirectory() as tmp:
        exe = Path(args.woolly) if args.woolly else build(Path(tmp))
        rows, times = [], []
        for name in names:
            repo = clone(name, cache)
            if repo is None:
                print(f"  skip  {name} (clone failed)")
                continue
            head = subprocess.run(["git", "-C", str(repo), "rev-parse", "--short", "HEAD"], capture_output=True, text=True).stdout.strip()
            rep, took = check(exe, repo, Path(tmp) / name.replace("/", "_"))
            mem = (rep or {}).get("memory_files") or {}
            files = mem.get("files") or []
            if not files:
                print(f"  none  {name} (no agent files)")
                continue
            found = mem.get("findings") or []
            times.append(took)
            rows.append({
                "repo": name, "head": head, "seconds": round(took, 2),
                "files": [f["file"] for f in files], "lines": mem.get("lines", 0),
                "partial": mem.get("partial", False), "notes": mem.get("notes") or [],
                "findings": [{k: f.get(k) for k in ("file", "line", "text", "what", "kind", "reason", "commit", "date", "fix")} for f in found],
            })
            print(f"  {len(found):4d}  {name}  ({mem.get('lines', 0)} lines, {took:.1f}s)")

    with_files = len(rows)
    with_stale = sum(1 for r in rows if r["findings"])
    lines = sum(r["lines"] for r in rows)
    stale = sum(len(r["findings"]) for r in rows)
    summary = {
        "date": date.today().isoformat(),
        "repos_checked": with_files,
        "repos_with_stale_lines": with_stale,
        "lines_checked": lines,
        "stale_lines": stale,
        "repos_with_notes": sum(1 for r in rows if r["notes"]),
        "median_seconds": round(statistics.median(times), 2) if times else 0,
    }
    out = Path(args.out)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps({"summary": summary, "repos": rows}, indent=2) + "\n")
    print()
    print(f"{with_stale} of {with_files} repos have at least one line that names code that's gone.")
    print(f"{stale} of {lines} lines. Median check: {summary['median_seconds']}s. Saved {out}")


if __name__ == "__main__":
    main()
