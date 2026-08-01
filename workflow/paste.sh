#!/bin/sh
set -u

text=${1-}
printf '%s' "$text" | /usr/bin/pbcopy

# Clipboard success is the fallback when Accessibility permission is absent.
/usr/bin/osascript -e 'tell application "System Events" to keystroke "v" using command down' >/dev/null 2>&1 || true
