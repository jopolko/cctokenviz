#!/usr/bin/env python3
"""Draw the CCTokenViz icon and write build/appicon.png and build/windows/icon.ico.

Three context meters on a dark tile (green, amber, red: low, medium and nearly full), the
way the app's rows fill. Pure Python, no dependencies: shapes are rendered with 4x4
supersampling and written as PNG; the .ico embeds PNGs at 16 to 256 px.
Usage: python scripts/make-icon.py
"""
import os
import struct
import zlib

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

BG = (14, 22, 16)
RIM = (38, 70, 48)
TRACK = (36, 54, 42)
BARS = [  # (y centre, fill fraction, colour)
    (0.31, 0.46, (46, 230, 107)),
    (0.50, 0.72, (245, 217, 10)),
    (0.69, 0.93, (255, 77, 77)),
]
BAR_X0, BAR_X1, BAR_H = 0.17, 0.83, 0.115


def rrect(x, y, x0, y0, x1, y1, r):
    """True if (x, y) is inside the rounded rectangle."""
    if x < x0 or x > x1 or y < y0 or y > y1:
        return False
    cx = min(max(x, x0 + r), x1 - r)
    cy = min(max(y, y0 + r), y1 - r)
    return (x - cx) ** 2 + (y - cy) ** 2 <= r * r


def sample(x, y):
    """Colour (r, g, b, a) of the icon at unit coordinates, or None outside the tile."""
    if not rrect(x, y, 0.02, 0.02, 0.98, 0.98, 0.21):
        return None
    c = RIM if not rrect(x, y, 0.045, 0.045, 0.955, 0.955, 0.185) else BG
    for cy, frac, col in BARS:
        y0, y1 = cy - BAR_H / 2, cy + BAR_H / 2
        if rrect(x, y, BAR_X0, y0, BAR_X1, y1, BAR_H / 2):
            c = TRACK
            if rrect(x, y, BAR_X0, y0, BAR_X0 + (BAR_X1 - BAR_X0) * frac, y1, BAR_H / 2):
                c = col
    return c + (255,)


def render(size, ss=4):
    rows = []
    for py in range(size):
        row = bytearray()
        for px in range(size):
            r = g = b = a = 0
            for sy in range(ss):
                for sx in range(ss):
                    s = sample((px + (sx + 0.5) / ss) / size, (py + (sy + 0.5) / ss) / size)
                    if s:
                        r, g, b, a = r + s[0], g + s[1], b + s[2], a + 255
            n = ss * ss
            if a:
                cnt = a // 255
                row += bytes((r // cnt, g // cnt, b // cnt, a // n))
            else:
                row += b"\x00\x00\x00\x00"
        rows.append(bytes(row))
    return rows


def png(size):
    raw = b"".join(b"\x00" + r for r in render(size))

    def chunk(tag, data):
        body = tag + data
        return struct.pack(">I", len(data)) + body + struct.pack(">I", zlib.crc32(body) & 0xFFFFFFFF)

    return (b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">IIBBBBB", size, size, 8, 6, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(raw, 9)) + chunk(b"IEND", b""))


def ico(sizes):
    images = [png(s) for s in sizes]
    head = struct.pack("<HHH", 0, 1, len(sizes))
    offset = 6 + 16 * len(sizes)
    entries = b""
    for s, data in zip(sizes, images):
        entries += struct.pack("<BBBBHHII", s % 256, s % 256, 0, 0, 1, 32, len(data), offset)
        offset += len(data)
    return head + entries + b"".join(images)


def main():
    out_png = os.path.join(ROOT, "build", "appicon.png")
    out_ico = os.path.join(ROOT, "build", "windows", "icon.ico")
    os.makedirs(os.path.dirname(out_ico), exist_ok=True)
    open(out_png, "wb").write(png(512))
    open(out_ico, "wb").write(ico([16, 24, 32, 48, 64, 128, 256]))
    print("wrote", out_png)
    print("wrote", out_ico)


if __name__ == "__main__":
    main()
