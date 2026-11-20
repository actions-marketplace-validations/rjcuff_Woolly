#!/bin/sh
# Makes a small git repo for testing the GitHub Action, and prints the base
# commit a pull request would compare against.
#   scripts/action-fixture.sh <dir>
#
# CLAUDE.md has three lines: one that was already stale on the base commit
# and two that the commit after it makes stale. One of them tries to mention
# someone and add an image, which the report must quote, not render.
set -eu

dir=$1
mkdir -p "$dir"
cd "$dir"
export GIT_AUTHOR_NAME=fixture GIT_AUTHOR_EMAIL=fixture@example.com
export GIT_COMMITTER_NAME=fixture GIT_COMMITTER_EMAIL=fixture@example.com

git init -q
# The Action's own source may be checked out inside this folder in CI.
printf 'woolly-src/\n' > .gitignore
printf 'func oldHelper() {}\nfunc requireUser() {}\n' > a.go
git add .
git commit -q -m one

printf 'func requireUser() {}\n' > a.go
cat > CLAUDE.md <<'EOF'
- Use oldHelper for dates
- Handlers call requireUser first
- Ask @octocat about requireUser <img src=https://example.com/x.png>
EOF
git add .
git commit -q -m "drop oldHelper"
base=$(git rev-parse HEAD)

printf 'func requireSession() {}\n' > a.go
git commit -q -am "rename requireUser"
echo "$base"
