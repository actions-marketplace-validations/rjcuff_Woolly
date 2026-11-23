#!/bin/sh
# Signs Windows programs when a code signing certificate is set up, and does
# nothing otherwise, so unsigned builds keep working.
#   WINDOWS_CERT_PATH=cert.pfx WINDOWS_CERT_PASSWORD=... scripts/winsign.sh woolly.exe
# Uses signtool on Windows, or osslsigncode anywhere else.
set -eu

[ -n "${WINDOWS_CERT_PATH:-}" ] || exit 0
[ -f "$WINDOWS_CERT_PATH" ] || { echo "WINDOWS_CERT_PATH is set but $WINDOWS_CERT_PATH doesn't exist" >&2; exit 1; }
timestamp="${WINDOWS_TIMESTAMP_URL:-http://timestamp.digicert.com}"

for file in "$@"; do
  if command -v signtool >/dev/null 2>&1; then
    # Git Bash rewrites arguments starting with / into paths; leave these alone.
    MSYS2_ARG_CONV_EXCL='*' signtool sign /fd sha256 /f "$WINDOWS_CERT_PATH" /p "${WINDOWS_CERT_PASSWORD:-}" /tr "$timestamp" /td sha256 "$file"
  elif command -v osslsigncode >/dev/null 2>&1; then
    osslsigncode sign -pkcs12 "$WINDOWS_CERT_PATH" -pass "${WINDOWS_CERT_PASSWORD:-}" -h sha256 -ts "$timestamp" -in "$file" -out "$file.signed"
    mv "$file.signed" "$file"
  else
    echo "A certificate is set, but neither signtool nor osslsigncode is installed" >&2
    exit 1
  fi
  echo "Signed $file"
done
