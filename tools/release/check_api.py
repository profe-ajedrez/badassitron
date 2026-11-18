#!/usr/bin/env python3
"""Compare the current AST API against frozen pre-migration source declarations.

Does not require network access or resolve Go dependencies. Behavior and
binary compatibility remain the responsibility of the full release gate.
"""
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[2]
FREEZE = ROOT / 'testdata/migration/api-surface'


def exported_key(entry):
    name = entry['name']
    if entry['kind'] == 'func' and '.' in name:
        receiver = name.rsplit('.', 1)[0].lstrip('*')
        if not receiver or not receiver[0].isupper():
            return None
    return (entry['package'], entry['kind'], name)


def check(historical, unsigned, current):
    old = {tuple(entry) for entry in historical}
    now = {k for entry in current if (k := exported_key(entry)) is not None}
    missing = sorted(old - now)
    if missing:
        raise ValueError(f'missing historical API identifiers: {missing}')
    old_unsigned = {(e['kind'], e['name']): e['signature'] for e in unsigned}
    new_unsigned = {(e['kind'], e['name']): e['signature'] for e in current if e['package'] == 'dec128/uint128'}
    gone = sorted(old_unsigned.keys() - new_unsigned.keys())
    added = sorted(new_unsigned.keys() - old_unsigned.keys())
    changed = sorted(key for key in old_unsigned.keys() & new_unsigned.keys() if old_unsigned[key] != new_unsigned[key])
    if gone or added or changed:
        raise ValueError(f'uint128 signatures diverged: missing={gone} added={added} changed={changed}')
    return len(old), len(old_unsigned), sorted(now - old)


def main():
    env = dict(os.environ, GO111MODULE='off')
    cmd = ['go', 'run', str(ROOT / 'tools/migration-api/main.go'), '-root', str(ROOT)]
    current = json.loads(subprocess.check_output(cmd, env=env, cwd=ROOT))
    historical = json.loads((FREEZE / 'historical-public-identifiers.json').read_text())
    unsigned = json.loads((FREEZE / 'historical-uint128-signatures.json').read_text())
    n, m, added = check(historical, unsigned, current)
    print(f'PASS: {n} historical public identifiers preserved; {len(added)} additions: {added}')
    print(f'PASS: all {m} unsigned public declarations retain exact source signatures')


if __name__ == '__main__':
    try:
        main()
    except (ValueError, subprocess.CalledProcessError) as exc:
        sys.exit(f'FAIL: {exc}')
