#!/usr/bin/env bash
# Creates labels, milestones and issues from .github/seed/*.json. Idempotent (skips existing titles).
# Usage: GH_TOKEN=<token with repo/issues scope> REPO=tecnologiasantos/bogoboom bash scripts/seed_github.sh
set -euo pipefail
: "${REPO:?set REPO=owner/name}"
DRY=${DRY_RUN:-false}
run() { if [ "$DRY" = "true" ]; then echo "[dry] $*"; else "$@"; fi; }
S=.github/seed

echo "== labels"
jq -c '.[]' $S/labels.json | while read -r l; do
  n=$(jq -r .name <<<"$l"); c=$(jq -r .color <<<"$l")
  run gh label create "$n" --color "$c" --repo "$REPO" --force >/dev/null && echo "  $n"
done

echo "== milestones"
existing_ms=$(gh api "repos/$REPO/milestones?state=all&per_page=100" --jq '.[].title' || true)
jq -r '.[].title' $S/milestones.json | while read -r t; do
  grep -qxF "$t" <<<"$existing_ms" && continue
  run gh api "repos/$REPO/milestones" -f title="$t" >/dev/null && echo "  $t"
done

echo "== issues"
existing=$(gh issue list --repo "$REPO" --state all --limit 500 --json title --jq '.[].title' || true)
n=$(jq length $S/issues.json)
for i in $(seq 0 $((n-1))); do
  t=$(jq -r ".[$i].title" $S/issues.json)
  grep -qxF "$t" <<<"$existing" && { echo "  skip: $t"; continue; }
  jq -r ".[$i].body" $S/issues.json > /tmp/body.md
  m=$(jq -r ".[$i].milestone" $S/issues.json)
  args=(--repo "$REPO" --title "$t" --body-file /tmp/body.md --milestone "$m")
  for lb in $(jq -r ".[$i].labels[]" $S/issues.json | tr ' ' '#'); do args+=(--label "${lb//#/ }"); done
  run gh issue create "${args[@]}" && sleep 1
done
echo done
