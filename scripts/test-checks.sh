#!/bin/bash
# ABOUTME: Proves a Workshop's gated Checks (ADR 0003). For each coding Challenge,
# in a scratch copy of the Exercise: the Check must FAIL with that Challenge's TODOs
# undone (earlier ones done), and PASS once they are spliced in from the Solution.
#
#   scripts/test-checks.sh workshops/<slug>
#
# Reads workshops/<slug>/sandbox/challenges.tsv: one line per coding Challenge,
#   <challenge-number> <TAB> <file> <TAB> <TODO label>[,<TODO label>...]
# Needs Go, jq and the temporal CLI locally.
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WS="$(cd "${1:?usage: test-checks.sh workshops/<slug>}" && pwd)"
MAP="$WS/sandbox/challenges.tsv"
SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/test-checks-XXXXXX")"
trap 'rm -rf "$SCRATCH"' EXIT
mkdir -p "$SCRATCH/exercise" "$SCRATCH/bin"
(cd "$ROOT/workshops/_sandbox/workshop-solve" && go build -o "$SCRATCH/bin/workshop-solve" .) || exit 1
cp "$ROOT/workshops/_sandbox/bin/workshop-check" "$SCRATCH/bin/"
[ -n "$(ls -A "$WS/sandbox/bin" 2>/dev/null)" ] && cp "$WS/sandbox/bin/"* "$SCRATCH/bin/"
export PATH="$SCRATCH/bin:$PATH"
(cd "$WS/exercise" && tar --exclude '*.bak' -cf - .) | (cd "$SCRATCH/exercise" && tar -xf -)
FAILS=0

# Rows that share a Challenge number (one Challenge, TODOs in several files) are
# one group: check once before, splice every row, check once after.
ROWS=()
while IFS= read -r line; do
    [ -z "$line" ] || [ "${line:0:1}" = "#" ] && continue
    ROWS+=("$line")
done <"$MAP"

i=0
while [ "$i" -lt "${#ROWS[@]}" ]; do
    N="${ROWS[$i]%%$'\t'*}"
    GROUP=()
    while [ "$i" -lt "${#ROWS[@]}" ] && [ "${ROWS[$i]%%$'\t'*}" = "$N" ]; do
        GROUP+=("${ROWS[$i]}"); i=$((i + 1))
    done
    LABELS="$(for row in "${GROUP[@]}"; do echo "${row##*$'\t'}"; done | paste -sd',' -)"

    printf 'Challenge %s, before %s: ' "$N" "$LABELS"
    if MSG="$(workshop-check "$SCRATCH/exercise" "$N" 2>&1)"; then
        echo "PASSED, but must fail"; FAILS=$((FAILS + 1))
    else
        echo "fails as it should"; echo "    learner sees: ${MSG#FAIL: }"
    fi
    for row in "${GROUP[@]}"; do
        IFS=$'\t' read -r _ FILE TODOS <<<"$row"
        IFS=',' read -r -a LIST <<<"$TODOS"
        workshop-solve "$SCRATCH/exercise" "$WS/solution" "$FILE" "${LIST[@]}" >/dev/null || FAILS=$((FAILS + 1))
    done
    printf 'Challenge %s, after  %s: ' "$N" "$LABELS"
    if MSG="$(workshop-check "$SCRATCH/exercise" "$N" 2>&1)"; then
        echo "passes"
    else
        echo "FAILED, must pass"; echo "    ${MSG}"; FAILS=$((FAILS + 1))
    fi
done

# Every TODO spliced in, the scratch Exercise must equal the Solution. Both trees
# share one module path, so nothing is excluded but the backups.
if diff -rq -x '*.bak' "$SCRATCH/exercise" "$WS/solution" >/dev/null; then
    echo "All TODOs done: the Exercise is identical to the Solution."
else
    echo "All TODOs done, but the Exercise still differs from the Solution:"
    diff -r -x '*.bak' "$SCRATCH/exercise" "$WS/solution" | head -20
    FAILS=$((FAILS + 1))
fi

[ "$FAILS" -eq 0 ] && echo "OK: every Check gates its Challenge." || echo "$FAILS problem(s)."
exit $((FAILS > 0))
