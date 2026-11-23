#!/bin/sh
# Writes the Scoop and winget manifests for a release into dist/, using the
# checksums in dist/. Submit them to a Scoop bucket and to
# microsoft/winget-pkgs by hand.
#   scripts/windows-packages.sh 0.20.0
set -eu

version="${1:?usage: scripts/windows-packages.sh <version>}"
cd "$(dirname "$0")/.."
sums=dist/checksums.txt
[ -f "$sums" ] || { echo "No $sums. Run: make release VERSION=$version" >&2; exit 1; }

sha() { grep " woolly_$1.zip\$" "$sums" | awk '{print $1}'; }
upper() { printf '%s' "$1" | tr '[:lower:]' '[:upper:]'; }
amd64=$(sha windows_amd64)
arm64=$(sha windows_arm64)
[ -n "$amd64" ] && [ -n "$arm64" ] || { echo "Missing Windows zips in $sums" >&2; exit 1; }
base="https://github.com/rjcuff/woolly/releases/download/v$version"

mkdir -p dist/scoop
cat > dist/scoop/woolly.json <<JSON
{
  "version": "$version",
  "description": "Local project memory for coding agents that flags stale facts",
  "homepage": "https://usewoolly.com",
  "license": "MIT",
  "architecture": {
    "64bit": {
      "url": "$base/woolly_windows_amd64.zip",
      "hash": "$amd64"
    },
    "arm64": {
      "url": "$base/woolly_windows_arm64.zip",
      "hash": "$arm64"
    }
  },
  "bin": "woolly.exe",
  "checkver": "github",
  "autoupdate": {
    "architecture": {
      "64bit": { "url": "https://github.com/rjcuff/woolly/releases/download/v\$version/woolly_windows_amd64.zip" },
      "arm64": { "url": "https://github.com/rjcuff/woolly/releases/download/v\$version/woolly_windows_arm64.zip" }
    }
  }
}
JSON

id=RyanCuff.Woolly
dir="dist/winget/manifests/r/RyanCuff/Woolly/$version"
mkdir -p "$dir"
cat > "$dir/$id.yaml" <<YAML
PackageIdentifier: $id
PackageVersion: $version
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.6.0
YAML
cat > "$dir/$id.installer.yaml" <<YAML
PackageIdentifier: $id
PackageVersion: $version
InstallerType: zip
NestedInstallerType: portable
NestedInstallerFiles:
  - RelativeFilePath: woolly.exe
    PortableCommandAlias: woolly
Installers:
  - Architecture: x64
    InstallerUrl: $base/woolly_windows_amd64.zip
    InstallerSha256: $(upper "$amd64")
  - Architecture: arm64
    InstallerUrl: $base/woolly_windows_arm64.zip
    InstallerSha256: $(upper "$arm64")
ManifestType: installer
ManifestVersion: 1.6.0
YAML
cat > "$dir/$id.locale.en-US.yaml" <<YAML
PackageIdentifier: $id
PackageVersion: $version
PackageLocale: en-US
Publisher: Ryan Cuff
PublisherUrl: https://usewoolly.com
PackageName: Woolly
PackageUrl: https://usewoolly.com
License: MIT
LicenseUrl: https://github.com/rjcuff/woolly/blob/main/LICENSE
ShortDescription: Local project memory for coding agents that flags stale facts
Moniker: woolly
Tags:
  - ai
  - memory
  - mcp
  - claude
ManifestType: defaultLocale
ManifestVersion: 1.6.0
YAML
echo "Wrote dist/scoop/woolly.json and $dir"
