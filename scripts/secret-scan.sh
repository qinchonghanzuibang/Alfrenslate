#!/bin/sh
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
if /usr/bin/grep -RIE --exclude-dir=.git --exclude='*.alfredworkflow' --exclude='*.sha256' '(sk-[A-Za-z0-9_-]{20,}|Bearer [A-Za-z0-9._~+/=-]{16,}|-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----)' "$root" >/dev/null 2>&1; then
  printf '%s\n' "Potential secret detected." >&2
  exit 1
fi
printf '%s\n' "Secret scan passed."
