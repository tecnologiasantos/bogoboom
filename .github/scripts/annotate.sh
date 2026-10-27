#!/usr/bin/env bash
# Runs a command; if it fails, repeats its last output lines as GitHub error annotations so the
# failure is visible on the check run (and through the API) without opening the raw log.
# Usage: annotate.sh "<title>" <command...>
title="$1"; shift
out="$(mktemp)"
set +e
"$@" 2>&1 | tee "$out"
code=${PIPESTATUS[0]}
set -e
if [ "$code" -ne 0 ]; then
  ( grep -E "FAIL|panic|_test\.go|Error|expected|actual|want|got" "$out" | grep -v "Auth0\] Unauthorized" | head -n 40; tail -n 5 "$out" ) | while IFS= read -r line; do
    esc="${line//'%'/'%25'}"; esc="${esc//$'\r'/'%0D'}"
    echo "::error title=${title}::${esc:0:400}"
  done
fi
exit "$code"
