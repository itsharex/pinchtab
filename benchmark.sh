#!/bin/bash
# Benchmark: Pinchtab vs OpenClaw Browser
# Tests: navigate, snapshot, screenshot, text extraction
# Runs each 5 times, reports avg

PINCHTAB="http://localhost:18800"
TEST_URL="https://news.ycombinator.com"
RUNS=5

echo "═══════════════════════════════════════════════"
echo "  Pinchtab vs OpenClaw Browser — Benchmark"
echo "═══════════════════════════════════════════════"
echo ""
echo "Target: $TEST_URL"
echo "Runs per test: $RUNS"
echo ""

# ── Pinchtab Tests ──────────────────────────────

echo "┌─────────────────────────────────────────────┐"
echo "│  PINCHTAB (HTTP API, headless Chrome/CDP)   │"
echo "└─────────────────────────────────────────────┘"

# Navigate
echo -n "  Navigate:   "
total=0
for i in $(seq 1 $RUNS); do
  ms=$(curl -s -o /dev/null -w '%{time_total}' -X POST "$PINCHTAB/navigate" \
    -H 'Content-Type: application/json' \
    -d "{\"url\":\"$TEST_URL\"}")
  total=$(echo "$total + $ms" | bc)
done
avg=$(echo "scale=3; $total / $RUNS * 1000" | bc)
echo "${avg}ms avg"

# Snapshot
echo -n "  Snapshot:   "
total=0
sizes=0
for i in $(seq 1 $RUNS); do
  tmpfile=$(mktemp)
  ms=$(curl -s -o "$tmpfile" -w '%{time_total}' "$PINCHTAB/snapshot")
  total=$(echo "$total + $ms" | bc)
  sz=$(wc -c < "$tmpfile")
  sizes=$(echo "$sizes + $sz" | bc)
  if [ $i -eq 1 ]; then
    count=$(cat "$tmpfile" | python3 -c "import sys,json; print(json.load(sys.stdin)['count'])" 2>/dev/null)
  fi
  rm "$tmpfile"
done
avg=$(echo "scale=3; $total / $RUNS * 1000" | bc)
avg_sz=$(echo "scale=0; $sizes / $RUNS" | bc)
echo "${avg}ms avg | ${count} nodes | ~${avg_sz} bytes"

# Snapshot (interactive only)
echo -n "  Snap+filter:"
total=0
for i in $(seq 1 $RUNS); do
  tmpfile=$(mktemp)
  ms=$(curl -s -o "$tmpfile" -w '%{time_total}' "$PINCHTAB/snapshot?filter=interactive")
  total=$(echo "$total + $ms" | bc)
  if [ $i -eq 1 ]; then
    icount=$(cat "$tmpfile" | python3 -c "import sys,json; print(json.load(sys.stdin)['count'])" 2>/dev/null)
  fi
  rm "$tmpfile"
done
avg=$(echo "scale=3; $total / $RUNS * 1000" | bc)
echo " ${avg}ms avg | ${icount} nodes"

# Screenshot
echo -n "  Screenshot: "
total=0
for i in $(seq 1 $RUNS); do
  ms=$(curl -s -o /dev/null -w '%{time_total}' "$PINCHTAB/screenshot?raw=true")
  total=$(echo "$total + $ms" | bc)
done
avg=$(echo "scale=3; $total / $RUNS * 1000" | bc)
echo "${avg}ms avg"

# Text
echo -n "  Text:       "
total=0
for i in $(seq 1 $RUNS); do
  tmpfile=$(mktemp)
  ms=$(curl -s -o "$tmpfile" -w '%{time_total}' "$PINCHTAB/text")
  total=$(echo "$total + $ms" | bc)
  if [ $i -eq 1 ]; then
    tlen=$(cat "$tmpfile" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['text']))" 2>/dev/null)
  fi
  rm "$tmpfile"
done
avg=$(echo "scale=3; $total / $RUNS * 1000" | bc)
echo "${avg}ms avg | ${tlen} chars"

echo ""
echo "Done. OpenClaw browser test runs separately via tool calls."
