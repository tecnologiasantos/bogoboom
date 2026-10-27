#!/usr/bin/env python3
"""Create labels, milestones and issues from .github/seed/*.json using only the GitHub REST API (via `gh api`).
Idempotent: existing labels are updated, existing milestones/issues (by title) are skipped.
Usage: REPO=owner/name [DRY_RUN=true] python3 v2/scripts/seed_github.py   (run from the repo root; needs `gh` authenticated)"""
import json, os, subprocess, sys, time

REPO = os.environ.get("REPO") or sys.exit("set REPO=owner/name")
DRY = os.environ.get("DRY_RUN", "false").lower() == "true"
SEED = os.environ.get("SEED_DIR", ".github/seed")

def gh(method, path, payload=None, ok_codes=()):
    cmd = ["gh", "api", "-X", method, path]
    if payload is not None:
        cmd += ["--input", "-"]
    r = subprocess.run(cmd, input=json.dumps(payload) if payload is not None else None, capture_output=True, text=True)
    if r.returncode != 0:
        if any(c in r.stderr for c in ok_codes):
            return None
        raise RuntimeError(f"{method} {path}: {r.stderr.strip()[:300]}")
    return json.loads(r.stdout) if r.stdout.strip() else None

def paged(path):
    out, page = [], 1
    while True:
        chunk = gh("GET", f"{path}{'&' if '?' in path else '?'}per_page=100&page={page}") or []
        out += chunk
        if len(chunk) < 100:
            return out
        page += 1

load = lambda n: json.load(open(f"{SEED}/{n}.json"))

print("== labels")
for l in load("labels"):
    if DRY: print("  [dry]", l["name"]); continue
    created = gh("POST", f"/repos/{REPO}/labels", {"name": l["name"], "color": l["color"]}, ok_codes=("already_exists", "422"))
    if created is None:
        from urllib.parse import quote
        gh("PATCH", f"/repos/{REPO}/labels/{quote(l['name'])}", {"color": l["color"]})
    print("  ", l["name"])

print("== milestones")
ms = {} if DRY else {m["title"]: m["number"] for m in paged(f"/repos/{REPO}/milestones?state=all")}
for m in load("milestones"):
    if m["title"] in ms: continue
    if DRY: print("  [dry]", m["title"]); continue
    ms[m["title"]] = gh("POST", f"/repos/{REPO}/milestones", {"title": m["title"]})["number"]
    print("  ", m["title"])

print("== issues")
existing = set() if DRY else {i["title"] for i in paged(f"/repos/{REPO}/issues?state=all") if "pull_request" not in i}
created = skipped = 0
for i in load("issues"):
    if i["title"] in existing:
        skipped += 1; continue
    if DRY: print("  [dry]", i["title"]); continue
    gh("POST", f"/repos/{REPO}/issues", {"title": i["title"], "body": i["body"], "labels": i["labels"], "milestone": ms.get(i["milestone"])})
    created += 1
    print("  +", i["title"])
    time.sleep(1.2)  # stay clear of secondary rate limits
print(f"done: {created} created, {skipped} already existed")
