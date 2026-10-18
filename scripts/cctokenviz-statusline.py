#!/usr/bin/env python3
"""Minimal Claude Code status line that also feeds CCTokenViz.

Claude Code pipes a JSON blob to the status line command on every refresh. This script
saves that JSON to ~/.cctokenviz/status-<session_id>.json, which is where CCTokenViz
reads plan limits (5-hour / weekly) and the real context window size from, then prints a
short status line.

Already have a status line script? Copy `export_for_cctokenviz` into it and call it
with the parsed JSON; nothing else is needed.

Setup (~/.claude/settings.json):
  "statusLine": {"type": "command", "command": "python /path/to/cctokenviz-statusline.py"}
"""
import json
import os
import sys


def export_for_cctokenviz(d):
    """Write the raw status JSON where CCTokenViz can read it. Never raises."""
    try:
        sid = str(d.get("session_id") or "")
        if not sid or not all(c.isalnum() or c in "-_" for c in sid):
            return
        out = os.path.join(os.path.expanduser("~"), ".cctokenviz")
        os.makedirs(out, exist_ok=True)
        tmp = os.path.join(out, f"status-{sid}.json.tmp")
        with open(tmp, "w", encoding="utf-8") as f:
            json.dump(d, f)
        os.replace(tmp, os.path.join(out, f"status-{sid}.json"))
    except Exception:
        pass


def main():
    try:
        d = json.load(sys.stdin)
    except Exception:
        return  # never break the prompt on bad input

    export_for_cctokenviz(d)

    parts = [d.get("model", {}).get("display_name") or "Claude"]
    ctx = d.get("context_window", {}).get("used_percentage")
    if ctx is not None:
        parts.append(f"ctx {round(float(ctx))}%")
    limits = d.get("rate_limits", {})
    five = limits.get("five_hour", {}).get("used_percentage")
    week = limits.get("seven_day", {}).get("used_percentage")
    if five is not None:
        parts.append(f"5h {round(float(five))}%")
    if week is not None:
        parts.append(f"7d {round(float(week))}%")
    sys.stdout.write(" | ".join(parts))


if __name__ == "__main__":
    main()
