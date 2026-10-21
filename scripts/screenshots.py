#!/usr/bin/env python3
"""Regenerate the README screenshots from the real frontend and demo data.

Renders frontend/ in headless Edge or Chrome with docs/preview/demo-stub.js standing in for
the app backend, so no real project names appear. Usage: python scripts/screenshots.py [name-substring]
"""
import os, shutil, subprocess, sys, tempfile, time

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BROWSERS = [
    r"C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe",
    r"C:\Program Files\Google\Chrome\Application\chrome.exe",
    "/usr/bin/google-chrome", "/usr/bin/chromium",
]
# (file, theme, spend panel open, window height, show only the API cost section, window width)
SHOTS = [
    ("screenshot-daylight.png", "light", False, 700, False, 240),
    ("screenshot-api-cost.png", "light", True, 190, True, 240),
    # Wider than the real window so each line of the breakdown reads as a full sentence
    ("screenshot-breakdown.png", "light", False, 385, False, 460),
]

def main():
    browser = next((b for b in BROWSERS if os.path.exists(b)), None)
    if not browser:
        sys.exit("No Edge or Chrome found")
    tmp = tempfile.mkdtemp()
    shutil.copytree(os.path.join(ROOT, "frontend"), os.path.join(tmp, "frontend"), ignore=shutil.ignore_patterns("wailsjs"))
    shutil.copy(os.path.join(ROOT, "docs", "preview", "demo-stub.js"), os.path.join(tmp, "frontend", "demo-stub.js"))
    index = os.path.join(tmp, "frontend", "index.html")
    html = open(index, encoding="utf-8").read()
    html = html.replace('<script src="themes.js">', '<script src="demo-stub.js"></script>\n    <script src="themes.js">')
    open(index, "w", encoding="utf-8").write(html)
    for name, theme, spend, height, only, width in SHOTS:
        if len(sys.argv) > 1 and sys.argv[1] not in name:
            continue
        # Headless browsers won't go narrower than ~500px, so the app sits in a frame of its own width
        inner = "file:///" + index.replace("\\", "/") + f"?theme={theme}&spend={1 if spend else 0}" + ("&only=usage" if only else "") + ("&breakdown=1" if name == "screenshot-breakdown.png" else "")
        wrapper = os.path.join(tmp, f"shot-{name}.html")
        open(wrapper, "w", encoding="utf-8").write(
            f'<body style="margin:0;background:#000"><iframe src="{inner}" style="border:0;width:{width}px;height:{height}px;display:block"></iframe>')
        url = "file:///" + wrapper.replace("\\", "/")
        out = os.path.join(ROOT, "docs", name)
        started = time.time() - 1
        subprocess.run([browser, "--headless", "--disable-gpu", "--hide-scrollbars", "--force-device-scale-factor=2",
                        f"--window-size={width},{height}", "--virtual-time-budget=4000",
                        f"--user-data-dir={os.path.join(tmp, 'profile-' + name)}", f"--screenshot={out}", url],
                       check=True, capture_output=True)
        for _ in range(60):  # the file can land a moment after the browser returns
            if os.path.exists(out) and os.path.getmtime(out) >= started:
                break
            time.sleep(0.5)
        else:
            sys.exit(f"{browser} did not write {out}")
        print("wrote", out)
    shutil.rmtree(tmp, ignore_errors=True)

if __name__ == "__main__":
    main()
