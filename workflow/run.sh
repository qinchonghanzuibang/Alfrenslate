#!/bin/sh
set -eu

workflow_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
arch=$(uname -m)
case "$arch" in
  arm64) binary="$workflow_dir/bin/alfrenslate-arm64" ;;
  x86_64) binary="$workflow_dir/bin/alfrenslate-amd64" ;;
  *)
    printf '%s\n' "Alfrenslate does not support this Mac architecture: $arch" >&2
    exit 1
    ;;
esac

if [ ! -x "$binary" ]; then
  printf '%s\n' "Alfrenslate binary is missing or not executable: $binary" >&2
  exit 1
fi

exec "$binary" "$@"
