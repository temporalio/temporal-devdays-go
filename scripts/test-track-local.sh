#!/bin/bash
# ABOUTME: Runs a Workshop's Track in its local Sandbox Image, the way
# `instruqt track test` would, before anything is pushed.
#
#   scripts/test-track-local.sh workshops/<slug> <image>
#
# Build the image first, natively, from workshops/:
#   docker buildx build --platform linux/arm64 -f _sandbox/Dockerfile \
#     -t temporal-devdays-go-sandbox:local --load .
#
# Boots the image, fakes Instruqt's bootstrap marker, runs track setup, confirms
# the editor (8080) and Temporal UI (8233) answer, then for every Challenge in
# order: the Check must FAIL on the untouched state, the solve script runs, and
# the Check must PASS. Exits non-zero if any Challenge is broken.
set -uo pipefail

WS="$(cd "${1:?usage: test-track-local.sh workshops/<slug> <image>}" && pwd)"
IMAGE="${2:?usage: test-track-local.sh workshops/<slug> <image>}"
NAME="track-test-$$"
OUT="$(mktemp "${TMPDIR:-/tmp}/track-test.XXXXXX")"
SOLVE="$(mktemp "${TMPDIR:-/tmp}/track-solve.XXXXXX")"
FAILS=0

# The image's own platform: arm64 for a local build on Apple Silicon, amd64 in CI.
docker run -d --rm --name "$NAME" \
    -v "$WS/instruqt:/opt/track:ro" "$IMAGE" sleep infinity >/dev/null \
    || { echo "[FAIL] could not start $IMAGE"; exit 1; }
trap 'docker rm -f "$NAME" >/dev/null 2>&1; rm -f "$OUT" "$SOLVE"' EXIT
x() { docker exec "$NAME" bash -lc "$1"; }

x 'mkdir -p /opt/instruqt/bootstrap && touch /opt/instruqt/bootstrap/host-bootstrap-completed'
echo "==> track setup"
x 'bash /opt/track/track_scripts/setup-workshop' | tail -3

for port in 8080 8233; do
    code="$(x "curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:$port/")"
    [ "$code" = "200" ] || [ "$code" = "302" ] && echo "[ OK ] port $port answers ($code)" \
        || { echo "[FAIL] port $port answers $code"; FAILS=$((FAILS + 1)); }
done

for dir in "$WS"/instruqt/[0-9][0-9]-*/; do
    c="$(basename "$dir")"
    if x "bash /opt/track/$c/check-workshop" >"$OUT" 2>&1; then
        echo "[FAIL] $c: Check passed before the work was done"; FAILS=$((FAILS + 1))
    else
        echo "[ OK ] $c: Check fails first: $(tail -1 "$OUT" | sed 's/^FAIL: //')"
    fi
    if ! x "bash /opt/track/$c/solve-workshop" >"$SOLVE" 2>&1; then
        echo "[FAIL] $c: solve script failed"; tail -5 "$SOLVE"; FAILS=$((FAILS + 1)); continue
    fi
    if START=$(date +%s) && x "bash /opt/track/$c/check-workshop" >"$OUT" 2>&1; then
        echo "[ OK ] $c: Check passes after solve ($(( $(date +%s) - START ))s)"
    else
        echo "[FAIL] $c: Check still fails after solve: $(tail -1 "$OUT")"; FAILS=$((FAILS + 1))
    fi
done

[ "$FAILS" -eq 0 ] && echo "OK: the Track runs end to end." || echo "$FAILS problem(s)."
exit $((FAILS > 0))
