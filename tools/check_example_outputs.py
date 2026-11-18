#!/usr/bin/env python3
"""Verify that every runnable example's documented output matches go run.

Run on the Go version declared in go.mod with the resolved module dependencies:
    GOWORK=off python3 tools/check_example_outputs.py
No files, dependencies, or module directives are modified by this tool.
"""
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
EXPECTED = re.compile(r'Expected output:\s*\n+```text\n(.*?)\n```', re.S)


def main() -> int:
    failures = []
    checked = 0
    for entry in sorted((ROOT / 'examples').iterdir()):
        if not (entry / 'main.go').exists():
            continue
        readme = entry / 'README.md'
        if not readme.exists():
            failures.append(f'{entry.name}: README missing')
            continue
        match = EXPECTED.search(readme.read_text(encoding='utf-8'))
        if not match:
            failures.append(f'{entry.name}: missing fenced Expected output block')
            continue
        want = match.group(1).rstrip() + '\n'
        result = subprocess.run(['go', 'run', f'./examples/{entry.name}'],
                                cwd=ROOT, capture_output=True, text=True,
                                timeout=120, check=False)
        checked += 1
        if result.returncode:
            failures.append(f'{entry.name}: go run failed ({result.returncode}): {result.stderr.strip()}')
        elif result.stdout != want:
            failures.append(f'{entry.name}: expected {want!r}, got {result.stdout!r}')
    if failures:
        print(f'FAIL: {len(failures)} example output problems ({checked} executed)')
        for failure in failures:
            print(' -', failure)
        return 1
    print(f'PASS: {checked} example programs match their README expected output blocks')
    return 0


if __name__ == '__main__':
    sys.exit(main())
