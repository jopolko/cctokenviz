#!/usr/bin/env python3
"""Regenerate the README screenshots from the real frontend and demo data.

Renders frontend/ in headless Edge or Chrome with docs/preview/demo-stub.js standing in for
the app backend, so no real project names appear. Usage: python scripts/screenshots.py
"""
import os, shutil, subprocess, sys, tempfile, time

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BROWSERS = [
    r"C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe",
    r"C:\Program Files\Google\Chrome\Application\chrome.exe",
    "/usr/bin/google-chrome", "/usr/bin/chromium",
]
# (file, theme, spend panel open, window height)
SHOTS = [
    ("screenshot.png", "linux", True, 840),
    ("screenshot-sweet.png", "sweet", True, 840),
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
    for name, theme, spend, height in SHOTS:
        # Headless browsers won't go narrower than ~500px, so the app sits in a 240px frame (its real width)
        inner = "file:///" + index.replace("\\", "/") + f"?theme={theme}&spend={1 if spend else 0}"
        wrapper = os.path.join(tmp, f"shot-{name}.html")
        open(wrapper, "w", encoding="utf-8").write(
            f'<body style="margin:0;background:#000"><iframe src="{inner}" style="border:0;width:240px;height:{height}px;display:block"></iframe>')
        url = "file:///" + wrapper.replace("\\", "/")
        out = os.path.join(ROOT, "docs", name)
        started = time.time() - 1
        subprocess.run([browser, "--headless", "--disable-gpu", "--hide-scrollbars", "--force-device-scale-factor=2",
                        f"--window-size=240,{height}", "--virtual-time-budget=4000",
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
