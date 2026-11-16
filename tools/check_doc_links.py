#!/usr/bin/env python3
"""Verify that local Markdown links resolve within the repository.

Uses only the Python standard library. Does not fetch external URLs or
validate HTML fragment anchors; it checks bundled paths in Markdown sources.
Run from any directory: python3 tools/check_doc_links.py
"""
from pathlib import Path
from urllib.parse import unquote, urlparse
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
LINK = re.compile(r'(?<!!)\[[^\]\n]+\]\(([^)]+)\)')
errors = []
links_checked = 0
files_checked = 0

for path in sorted(ROOT.rglob('*.md')):
    if any(p in ('.git', 'vendor') for p in path.relative_to(ROOT).parts):
        continue
    files_checked += 1
    in_fence = False
    for line_number, line in enumerate(path.read_text(encoding='utf-8').splitlines(), 1):
        if line.strip().startswith(('```', '~~~')):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        for match in LINK.finditer(line):
            target = match.group(1).strip().split()[0].strip('<>')
            parsed = urlparse(target)
            if parsed.scheme or target.startswith(('#', '//')):
                continue
            local = (path.parent / unquote(parsed.path)).resolve()
            links_checked += 1
            if not local.is_relative_to(ROOT.resolve()) or not local.exists():
                errors.append(f'{path.relative_to(ROOT)}:{line_number}: {target}')

if errors:
    print(f'FAIL: {len(errors)} broken local links in {files_checked} Markdown files')
    print('\n'.join(errors))
    sys.exit(1)
print(f'PASS: {links_checked} local links in {files_checked} Markdown files')
