# Contributing to Woolly

Thanks for helping out. Bug reports, fixes, docs and new ideas are all welcome.

## Before you start

For anything bigger than a small fix, open an issue first so we can agree on the approach. It saves you from writing code that doesn't get merged.

## Repo layout

| Folder | Contents |
| --- | --- |
| `cmd/woolly`, `internal` | The Woolly program, written in Go |
| `docs` | Documentation in MDX |
| `ui` | The Woolly app's screens, built into the binary |
| `app` | The Mac and Windows app (Tauri), which wraps the same screens |
| `sdk/python` | The `woolly-ai` Python package |
| `sdk/js` | The `woolly-ai` JavaScript and TypeScript package |
| `bench/agentfiles` | The stale agent file benchmark on 20 public repos |
| `bench/locomo` | The LOCOMO benchmark |
| `action.yml` | The GitHub Action that runs `woolly check` on pull requests |
| `plugin`, `.claude-plugin` | The Claude Code plugin and its marketplace entry |
| `web` | Website and docs site |
| `scripts` | Install, release, demo and Homebrew scripts |
| `packaging/homebrew` | The Homebrew formula, kept in step with the tap |
| `packaging/mcpb` | The Claude Desktop bundle |

## Setup

For Go changes you only need Go. Use Go 1.26 on macOS, since recent macOS versions refuse binaries from older Go linkers. Go 1.22 or newer works elsewhere.

```sh
git clone https://github.com/rjcuff/woolly
cd woolly
make build      # builds bin/woolly
make test       # go vet and Go tests, enough for Go-only changes
```

`make check` runs everything, including the website, the review page and both SDKs. It also needs Node 22 and python3, plus a one-time setup.

```sh
(cd web && npm install && npx playwright install chromium)
(cd ui && npm install)
make check      # run before a pull request that touches more than Go
```

Try your build without touching your real memories:

```sh
WOOLLY_HOME=/tmp/woolly-dev ./bin/woolly remember "test memory"
```

### Review page

The page `woolly ui` opens lives in `ui/` (React and TypeScript). Its build goes into `internal/server/ui` and is embedded in the binary, so commit the build output.

```sh
cd ui
npm install
npm run dev     # proxies API calls to a running woolly serve
make ui         # from the repo root, rebuild before committing
```

### Python package

The `woolly-ai` package lives in `sdk/python`. It has no dependencies. Its tests start a real `woolly serve`:

```sh
make py-test
```

To publish a new version to PyPI, bump the version in `sdk/python/pyproject.toml` and `sdk/python/woolly/__init__.py`, then run `make pypi`. It needs `pip install build twine` and a PyPI token.

### Website and docs

```sh
cd web
npm install
npm run dev     # http://localhost:5173
```

Docs live in the top-level `docs` folder as MDX. Add a page there and list it in `docs/meta.json`.

The site builds to static files in `web/build/client`. It deploys on Vercel with the project root set to `web`, using the settings in `web/vercel.json`.

## Guidelines

- Keep Woolly a single binary with no outside Go dependencies. If you need one, explain why in the issue.
- Memories must stay plain Markdown files that people can read and edit by hand.
- Nothing may send data off the user's machine.
- Anything that changes how data is stored, deleted or exposed needs a test, and an update to `docs/security.mdx` if it changes what Woolly protects against.
- Add a line to CHANGELOG.md for anything users will notice.
- Add tests for new behavior. Run `make check` before you commit.
- Run `gofmt` on Go code.
- Write docs in plain American English. Short sentences, no jargon.

## Pull requests

1. Fork the repo and create a branch from `main`.
2. Make your change with tests.
3. Make sure `make check` passes.
4. Open a pull request that says what changed and why.

## Demo GIF

`docs/demo.gif` at the top of the README is recorded from real output. After changing what commands print, re-record it:

```sh
make demo
```

`docs/init-demo.gif`, the shared project memory demo, comes from a [VHS](https://github.com/charmbracelet/vhs) tape:

```sh
make build && PATH="$PWD/bin:$PATH" vhs scripts/demo/init.tape
```

## Releases

```sh
make publish VERSION=0.1.1
```

This runs every test, builds archives for macOS, Linux and Windows into `dist/`, tags the commit, pushes, and creates the GitHub release with the archives attached. It uses the GitHub login git already has saved.

To build the archives without publishing, run `make release VERSION=0.1.1`.

## Reporting bugs

Open an issue with:

- What you ran
- What you expected
- What happened instead
- The output of `woolly doctor`

## Code of Conduct

Everyone working on Woolly follows the [Code of Conduct](CODE_OF_CONDUCT.md).

## Security

Please don't report security problems in public issues. See [SECURITY.md](SECURITY.md).

## License

By contributing, you agree that your work is released under the MIT license.
