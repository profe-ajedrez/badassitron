#!/usr/bin/env python3
"""Run each example and compare stdout with its README expected output.

Requires the Go toolchain from the repository go.mod. Uses only the Python
standard library; any Go compile error, mismatch or example timeout fails.
"""
from pathlib import Path
import os
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
examples = sorted((ROOT / 'examples').glob('*/main.go'))
failed = []
for main in examples:
    directory = main.parent
    readme = (directory / 'README.md').read_text(encoding='utf-8')
    match = re.search(r'Expected output:\s*```text\n(.*?)\n```', readme, re.S)
    if match is None:
        failed.append(f'{directory.name}: no expected output in README')
        continue
    result = subprocess.run(
        ['go', 'run', f'./examples/{directory.name}'], cwd=ROOT,
        env={**os.environ, 'GOWORK': 'off'}, capture_output=True, text=True,
        timeout=90, check=False,
    )
    expected = match.group(1).strip('\n')
    actual = result.stdout.strip('\n')
    if result.returncode != 0 or actual != expected:
        failed.append(f'{directory.name}: exit={result.returncode}\n'
                      f'  expected: {expected!r}\n'
                      f'  actual:   {actual!r}\n'
                      f'  stderr:   {result.stderr[:500]!r}')
    else:
        print(f'PASS: {directory.name}')

if failed:
    print('\n'.join(failed), file=sys.stderr)
    sys.exit(1)
print(f'PASS: all {len(examples)} example outputs match their READMEs')
