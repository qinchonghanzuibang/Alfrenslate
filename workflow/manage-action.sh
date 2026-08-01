#!/bin/sh
set -u

workflow_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
output=$("$workflow_dir/run.sh" action "${1-}" 2>&1)
status=$?
if [ -z "$output" ]; then
  if [ "$status" -eq 0 ]; then
    output="Action completed."
  else
    output="Action failed. Run alfrenslate diagnostics for details."
  fi
fi
printf '%s\n' "$output"
exit 0
