#!/usr/bin/env python3
"""Benchmark Pinchtab HTTP API performance"""
import time, json, subprocess, sys

PINCHTAB = "http://localhost:18800"
TEST_URL = "https://news.ycombinator.com"
RUNS = 5

def curl_timed(args, binary=False):
    """Run curl and return (elapsed_ms, response_body)"""
    start = time.time()
    result = subprocess.run(
        ["curl", "-s"] + args,
        capture_output=True, text=not binary, timeout=30
    )
    elapsed = (time.time() - start) * 1000
    out = result.stdout if not binary else result.stdout
    return elapsed, out

def bench(label, curl_args, parse_fn=None):
    times = []
    extra = None
    for i in range(RUNS):
        ms, body = curl_timed(curl_args)
        times.append(ms)
        if i == 0 and parse_fn:
            try:
                extra = parse_fn(body)
            except:
                extra = "parse error"
    avg = sum(times) / len(times)
    mn = min(times)
    mx = max(times)
    info = f" | {extra}" if extra else ""
    print(f"  {label:14s} {avg:7.1f}ms avg  {mn:7.1f}ms min  {mx:7.1f}ms max{info}")

print("═" * 55)
print("  PINCHTAB BENCHMARK")
print("═" * 55)
print(f"  Target: {TEST_URL}  |  Runs: {RUNS}")
print()

# Navigate first
bench("Navigate", [
    "-X", "POST", f"{PINCHTAB}/navigate",
    "-H", "Content-Type: application/json",
    "-d", json.dumps({"url": TEST_URL})
])

# Snapshot
bench("Snapshot", [f"{PINCHTAB}/snapshot"],
    lambda b: f"{json.loads(b)['count']} nodes, {len(b)} bytes")

# Snapshot interactive
bench("Snap+filter", [f"{PINCHTAB}/snapshot?filter=interactive"],
    lambda b: f"{json.loads(b)['count']} interactive nodes")

# Screenshot (binary, measure time only)
print("  Screenshot     ", end="", flush=True)
stimes = []
ssize = 0
for i in range(RUNS):
    start = time.time()
    r = subprocess.run(["curl", "-s", f"{PINCHTAB}/screenshot?raw=true"],
        capture_output=True, timeout=30)
    stimes.append((time.time() - start) * 1000)
    if i == 0: ssize = len(r.stdout)
avg = sum(stimes)/len(stimes)
print(f"  {avg:7.1f}ms avg  {min(stimes):7.1f}ms min  {max(stimes):7.1f}ms max | {ssize} bytes JPEG")

# Text
bench("Text", [f"{PINCHTAB}/text"],
    lambda b: f"{len(json.loads(b)['text'])} chars")

print()
