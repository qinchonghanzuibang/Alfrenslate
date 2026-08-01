#!/bin/sh
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
artifact="$root/Alfrenslate.alfredworkflow"
[ -f "$artifact" ] || { printf '%s\n' "Missing $artifact" >&2; exit 1; }

tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/alfrenslate-verify.XXXXXX")
trap 'rm -rf -- "$tmp_dir"' EXIT
/usr/bin/unzip -q "$artifact" -d "$tmp_dir"

/usr/bin/plutil -lint "$tmp_dir/info.plist" >/dev/null
/usr/bin/cmp -s "$root/workflow/info.plist" "$tmp_dir/info.plist" || {
  printf '%s\n' "Packaged info.plist differs from the verified source" >&2
  exit 1
}
[ "$(/usr/libexec/PlistBuddy -c 'Print :userconfigurationconfig' "$tmp_dir/info.plist" | /usr/bin/grep -c 'variable = ')" -eq 31 ]
[ "$(/usr/libexec/PlistBuddy -c 'Print :objects' "$tmp_dir/info.plist" | /usr/bin/grep -c 'alfred.workflow.trigger.universalaction')" -eq 1 ]
[ "$(/usr/libexec/PlistBuddy -c 'Print :version' "$tmp_dir/info.plist")" = "1.0.1" ]
[ "$(/usr/bin/grep -c '<key>keyword</key><string>ts</string>' "$tmp_dir/info.plist")" -eq 1 ]
if /usr/bin/grep -Eq '<key>keyword</key><string>(tx|zh|en)</string>' "$tmp_dir/info.plist"; then
  printf '%s\n' "Package contains an unsupported daily translation keyword" >&2
  exit 1
fi
for binary in alfrenslate-arm64 alfrenslate-amd64; do
  [ -x "$tmp_dir/bin/$binary" ] || { printf '%s\n' "Missing executable $binary" >&2; exit 1; }
done
/usr/bin/file "$tmp_dir/bin/alfrenslate-arm64" | /usr/bin/grep -q 'arm64'
/usr/bin/file "$tmp_dir/bin/alfrenslate-amd64" | /usr/bin/grep -q 'x86_64'
[ -x "$tmp_dir/run.sh" ] || { printf '%s\n' "run.sh is not executable" >&2; exit 1; }
[ -f "$tmp_dir/icon.png" ] || { printf '%s\n' "icon.png is missing" >&2; exit 1; }
/usr/bin/grep -q 'arm64)' "$tmp_dir/run.sh"
/usr/bin/grep -q 'x86_64)' "$tmp_dir/run.sh"
[ "$("$tmp_dir/run.sh" version)" = "1.0.1" ] || { printf '%s\n' "launcher version check failed" >&2; exit 1; }

if /usr/bin/find "$tmp_dir" -type f \( -name '.env' -o -name 'prefs.plist' -o -name '.DS_Store' \) | /usr/bin/grep -q .; then
  printf '%s\n' "Package contains a forbidden local file" >&2
  exit 1
fi
if /usr/bin/grep -RIE '(sk-[A-Za-z0-9_-]{20,}|Bearer [A-Za-z0-9._~+/=-]{16,}|app[_ -]?secret[=:][^[:space:]]+)' "$tmp_dir" >/dev/null 2>&1; then
  printf '%s\n' "Package may contain a credential" >&2
  exit 1
fi
printf '%s\n' "Workflow package verification passed."
