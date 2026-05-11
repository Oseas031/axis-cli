#!/usr/bin/env bash
# harness-audit.sh — measure whether the Axis agent harness is helping or just
# adding weight. Run weekly; trend matters more than any single number.
#
# Tiers map to docs/architecture (harness metrics framework):
#   Tier 1 — safety gates (target = 0, binary)
#   Tier 2 — correctness signals (continuous, trend ↑ or stable)
#   Tier 3 — friction signals (need session logs; partial here)
#
# Read-only. No side effects. Safe to run from any branch.
set -u

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"
if [[ -z "${REPO_ROOT}" ]]; then
  echo "[ERR] not inside a git repo" >&2
  exit 2
fi
cd "${REPO_ROOT}"

WINDOW="${1:-30}"   # how many recent commits to inspect

# Conventional-commit scope tag OR explicit milestone/spec reference.
# Anything else on main is a §9 traceability miss.
TAG_RE='^(feat|fix|chore|docs|refactor|test|perf|build|ci|merge|wip\(red\))(\([^)]+\))?:|\bM[0-9]+\b|\bPhase[[:space:]]+[0-9]|\bmilestone[[:space:]]*[0-9]|\bT[0-9]+\b'

# Milestone reference required on feat/fix/merge commits.
# chore/docs/refactor/test/ci/build/perf are scope-tag-sufficient (see decision log).
MILESTONE_RE='\bM[0-9]+\b|\bPhase[[:space:]]+[0-9]|\bmilestone[[:space:]]*[0-9]'
MILESTONE_EXEMPT_RE='^(chore|docs|refactor|test|ci|build|perf|merge)(\([^)]+\))?:'

# Build-artifact paths that must never enter the index.
ARTIFACT_RE='\.(exe|test|out)$|^(axis-dev|coverage\.|dist/|\.cache/)'

echo "=========================================="
echo " Axis harness audit  (window: last ${WINDOW} commits)"
echo " HEAD: $(git rev-parse --short HEAD)  branch: $(git rev-parse --abbrev-ref HEAD)"
echo " Date: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "=========================================="

# ---------- Tier 1 — safety gates ----------
echo
echo "[Tier 1] safety gates  (target = 0)"

untagged=$(git log -n "${WINDOW}" --pretty=format:'%s' \
  | grep -vEc "${TAG_RE}")
# A commit fails the milestone gate iff it is NOT scope-exempt AND lacks a milestone ref.
no_milestone=$(git log -n "${WINDOW}" --pretty=format:'%s' \
  | grep -vE "${MILESTONE_EXEMPT_RE}" \
  | grep -vEc "${MILESTONE_RE}")

printf "  untagged commits          (no conv-type / no milestone) : %d / %d\n" \
  "${untagged}" "${WINDOW}"
printf "  no-milestone on feat/fix  (after chore|docs|refactor... exempt): %d / %d\n" \
  "${no_milestone}" "${WINDOW}"

artifact_hits=$(git log -n "${WINDOW}" --diff-filter=A --name-only --pretty=format: \
  | grep -Ec "${ARTIFACT_RE}" || true)
printf "  added build artifacts     (*.exe / *.out / dist/ / ...) : %d\n" \
  "${artifact_hits}"

# Harness handoff anchor is HANDOVER.md (session-state.md was retired 2026-05-11).
if [[ ! -f HANDOVER.md ]]; then
  echo "  HANDOVER.md staleness                                   : MISSING"
else
  last_hand=$(git log -1 --pretty=format:'%H' -- HANDOVER.md)
  if [[ -n "${last_hand}" ]]; then
    hand_stale=$(git rev-list --count "${last_hand}..HEAD")
    printf "  HANDOVER.md staleness     (commits behind HEAD)         : %s\n" "${hand_stale}"
  else
    echo "  HANDOVER.md staleness                                   : UNTRACKED"
  fi
fi

prohibition_hits=$(git log -n "${WINDOW}" -p \
  | grep -E '^\+' \
  | grep -Ec '"(github\.com/gin-gonic|github\.com/labstack/echo|github\.com/gofiber|react|vue)"' \
  || true)
printf "  §1 prohibition hits       (forbidden frameworks added)  : %d\n" \
  "${prohibition_hits}"

# ---------- Tier 2 — correctness signals ----------
echo
echo "[Tier 2] correctness signals  (trend matters, not absolute)"

reverts=$(git log -n "${WINDOW}" --pretty=format:'%s' \
  | grep -Ec '^(Revert |fixup! |amend! |wip(\(red\))?:)' || true)
printf "  rework markers            (revert / fixup / amend / wip): %d / %d\n" \
  "${reverts}" "${WINDOW}"

spec_refs=$(git log -n "${WINDOW}" --pretty=format:'%s%n%b' \
  | grep -Ec 'docs/specs/|Spec-RDT|requirements\.md|design\.md|tasks\.md' \
  || true)
printf "  Spec-RDT references       (in subject or body)          : %d / %d\n" \
  "${spec_refs}" "${WINDOW}"

test_touch=$(git log -n "${WINDOW}" --name-only --pretty=format: \
  | grep -Ec '_test\.go$' || true)
code_touch=$(git log -n "${WINDOW}" --name-only --pretty=format: \
  | grep -E '\.go$' | grep -Evc '_test\.go$' || true)
ratio="n/a"
if [[ "${code_touch}" -gt 0 ]]; then
  ratio=$(awk -v t="${test_touch}" -v c="${code_touch}" 'BEGIN{printf "%.2f", t/c}')
fi
printf "  test-vs-code touch ratio  (test files / non-test .go)   : %s  (%d / %d)\n" \
  "${ratio}" "${test_touch}" "${code_touch}"

# ---------- Tier 3 — friction (partial; full version needs session logs) ----------
echo
echo "[Tier 3] friction  (partial — full version needs .devin/harness-log.jsonl)"

if [[ -f .devin/harness-log.jsonl ]]; then
  entries=$(wc -l < .devin/harness-log.jsonl | tr -d ' ')
  printf "  harness-log entries                                     : %s\n" "${entries}"
else
  echo "  harness-log.jsonl                                       : not yet wired"
  echo "    (lands with Suggestion C — deny-commit hook + session log)"
fi

echo
echo "=========================================="
echo " Done. Trend > absolute. Re-run weekly; archive output to"
echo " reports/harness-audit/<YYYY-MM-DD>.txt to build a baseline."
echo "=========================================="
