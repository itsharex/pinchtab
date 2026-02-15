#!/usr/bin/env python3
"""Side-by-side benchmark: Pinchtab (CDP) vs OpenClaw (Playwright)
Both on localhost, both headless, same page (HN)."""
import time, json, subprocess

PINCHTAB = "http://localhost:18801"
# OpenClaw browser control server (Playwright-based)
# We measure via internal HTTP since both are local
RUNS = 5

def timed_curl(args, binary=False):
    start = time.time()
    r = subprocess.run(["curl", "-s"] + args, capture_output=True, timeout=30)
    ms = (time.time() - start) * 1000
    body = r.stdout if binary else r.stdout.decode("utf-8", errors="replace")
    return ms, body

def bench(label, args, parse_fn=None, binary=False):
    times = []
    extra = ""
    for i in range(RUNS):
        ms, body = timed_curl(args, binary=binary)
        times.append(ms)
        if i == 0 and parse_fn:
            try: extra = f" | {parse_fn(body)}"
            except: extra = " | parse error"
    avg = sum(times)/len(times)
    print(f"  {label:16s} {avg:7.1f}ms avg  [{min(times):.0f}-{max(times):.0f}ms]{extra}")

print()
print("╔═══════════════════════════════════════════════════════╗")
print("║  PINCHTAB vs OPENCLAW BROWSER — Side-by-Side Bench   ║")
print("╠═══════════════════════════════════════════════════════╣")
print(f"║  Page: news.ycombinator.com  |  Runs: {RUNS}              ║")
print("╚═══════════════════════════════════════════════════════╝")

print()
print("  PINCHTAB (Go + raw CDP, headless Chrome)")
print("  " + "─" * 50)

bench("Snapshot (full)", [f"{PINCHTAB}/snapshot"],
    lambda b: f"{json.loads(b)['count']} nodes, {len(b):,} bytes")

bench("Snap (interact)", [f"{PINCHTAB}/snapshot?filter=interactive"],
    lambda b: f"{json.loads(b)['count']} nodes, {len(b):,} bytes")

bench("Snap (depth=3)", [f"{PINCHTAB}/snapshot?depth=3"],
    lambda b: f"{json.loads(b)['count']} nodes")

bench("Screenshot", [f"{PINCHTAB}/screenshot?raw=true"],
    lambda b: f"{len(b):,} bytes JPEG", binary=True)

bench("Text extract", [f"{PINCHTAB}/text"],
    lambda b: f"{len(json.loads(b)['text']):,} chars")

# Action: click ref (requires snapshot first for cache)
timed_curl([f"{PINCHTAB}/snapshot"])  # prime cache
bench("Click (by ref)", [
    "-X", "POST", f"{PINCHTAB}/action",
    "-H", "Content-Type: application/json",
    "-d", json.dumps({"kind": "click", "ref": "e4"})
])

# Navigate back for fair comparison
timed_curl(["-X", "POST", f"{PINCHTAB}/navigate",
    "-H", "Content-Type: application/json",
    "-d", json.dumps({"url": "https://news.ycombinator.com"})])
time.sleep(1)

print()
print("  OPENCLAW BROWSER (Node.js + Playwright, headless Chromium)")
print("  " + "─" * 50)
print("  (Measured via agent tool calls — includes IPC overhead)")
print("  Run the snapshot/screenshot tool calls and note the timing.")
print()
print("  ⚠️  OpenClaw timing includes: gateway→browser-server→Playwright→CDP")
print("      Pinchtab timing is: curl→Go HTTP→CDP (direct)")
print()

# Quick summary
snap_full = []
for i in range(RUNS):
    ms, _ = timed_curl([f"{PINCHTAB}/snapshot"])
    snap_full.append(ms)

print("  SUMMARY (Pinchtab only — OpenClaw via tool calls above)")
print(f"  Snapshot:    {sum(snap_full)/len(snap_full):.1f}ms avg")
print()
