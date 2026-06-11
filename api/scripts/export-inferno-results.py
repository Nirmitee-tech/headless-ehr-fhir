#!/usr/bin/env python3
"""Export Inferno session results to JSON + Markdown for publishing.

Usage: python3 scripts/export-inferno-results.py <session_id> [--inferno http://localhost:4567]
Writes: inferno-results/<date>/results.json, results.md
"""
import json, sys, urllib.request, datetime, os, re

session = sys.argv[1]
base = sys.argv[3] if len(sys.argv) > 3 else "http://localhost:4567"

with urllib.request.urlopen(f"{base}/api/test_sessions/{session}/results", timeout=30) as r:
    results = json.load(r)

date = datetime.date.today().isoformat()
outdir = f"inferno-results/{date}"
os.makedirs(outdir, exist_ok=True)

rows, passes, fails, tls_fails, skips = [], 0, 0, 0, 0
for r in results:
    tid = r.get("test_id") or ""
    if not tid:
        continue  # group aggregates
    st = r.get("result", "")
    msg = (r.get("result_message") or "").strip()
    short = tid.split("-")[-1]
    is_tls = "TLS" in msg or "tls" in short.lower()
    if st == "pass":
        passes += 1
    elif st == "fail" and is_tls:
        tls_fails += 1
    elif st == "fail":
        fails += 1
    elif st == "skip":
        skips += 1
        continue
    else:
        continue
    rows.append((short, st if not (st == "fail" and is_tls) else "fail (TLS-only)", msg[:110]))

functional_total = passes + fails
with open(f"{outdir}/results.json", "w") as f:
    json.dump({"session": session, "date": date,
               "summary": {"pass": passes, "functional_fail": fails,
                            "tls_only_fail": tls_fails, "skip": skips,
                            "functional_total": functional_total},
               "results": results}, f, indent=2)

md = [f"# Inferno ONC (g)(10) Test Results — {date}",
      "",
      f"Suite: **SMART App Launch — Standalone Patient App** (ONC Certification (g)(10) Standardized API Test Kit)",
      f"Server: headless-ehr-fhir, built-in SMART on FHIR server (standalone mode)",
      "",
      f"## Summary: {passes}/{functional_total + tls_fails} passing"
      + (f" ({fails} functional failures, {tls_fails} TLS-only failures expected in HTTP dev mode)" if (fails or tls_fails) else " — all tests green"),
      "",
      "| Test | Result | Notes |",
      "|------|--------|-------|"]
for short, st, msg in rows:
    icon = {"pass": "✅", "fail": "❌"}.get(st, "⚠️")
    md.append(f"| {short} | {icon} {st} | {msg} |")
md += ["", f"Session ID: `{session}` — reproduce by running Inferno against this repo (see README).", ""]
with open(f"{outdir}/results.md", "w") as f:
    f.write("\n".join(md))

print(f"pass={passes} fail={fails} tls={tls_fails} skip={skips} → {outdir}/")
