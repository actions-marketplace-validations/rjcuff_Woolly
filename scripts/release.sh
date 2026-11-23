#!/bin/sh
# Builds release archives into dist/ for every supported platform.
#   scripts/release.sh 0.1.0
set -eu

version="${1:?usage: scripts/release.sh <version>}"
cd "$(dirname "$0")/.."

# The Claude Code plugin updates for users only when its version changes.
for manifest in plugin/.claude-plugin/plugin.json plugin-directory/.claude-plugin/plugin.json; do
  plugin_version=$(sed -n 's/.*"version": *"\([^"]*\)".*/\1/p' "$manifest" | head -1)
  if [ "$plugin_version" != "$version" ]; then
    echo "$manifest has version $plugin_version, not $version. Bump it first." >&2
    exit 1
  fi
done
rm -rf dist
mkdir -p dist

for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 linux/arm windows/amd64 windows/arm64; do
  os=${target%/*}
  arch=${target#*/}
  name="woolly_${os}_${arch}"
  dir="dist/$name"
  bin="woolly"
  [ "$os" = windows ] && bin="woolly.exe"

  mkdir -p "$dir"
  # GOARM=7 runs on Raspberry Pi 2 and newer with a 32-bit OS.
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch GOARM=7 go build -trimpath \
    -ldflags "-s -w -X main.version=$version" -o "$dir/$bin" ./cmd/woolly
  cp LICENSE README.md "$dir/"

  if [ "$os" = windows ]; then
    scripts/winsign.sh "$dir/$bin"
    (cd "$dir" && zip -q "../$name.zip" "$bin" LICENSE README.md)
  else
    tar -czf "dist/$name.tar.gz" -C "$dir" "$bin" LICENSE README.md
  fi
  rm -rf "$dir"
done

(cd dist && shasum -a 256 woolly_* > checksums.txt)
scripts/mcpb.sh "$version"
scripts/windows-packages.sh "$version"
echo "Built woolly $version into dist/"
ls -1 dist
