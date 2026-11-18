#!/usr/bin/env python3
"""Stage an M7-G parser candidate from M7-F without changing the baseline.

This is EXPERIMENTAL ONLY. A staged tree must never be treated as a release.
The script requires an absent --output directory and validates the single-file
production change, identical module metadata and no replace/vendor.
"""
import argparse
import hashlib
from pathlib import Path
import shutil
import sys

PARSER = Path('dec128/uint128/from_upstream_string.go')
FILES = {
    'm7e': 'm7e-branchless.go.txt',
    'g2': 'g2-success-fastpath.go.txt',
    'g3': 'g3-error-shortcuts.go.txt',
}

def hashfile(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def production(root):
    return {p.relative_to(root):hashfile(p) for p in root.rglob('*.go')
            if not p.name.endswith('_test.go') and 'tools' not in p.relative_to(root).parts}

def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--baseline', required=True, type=Path)
    ap.add_argument('--variant', choices=sorted(FILES), required=True)
    ap.add_argument('--output', required=True, type=Path)
    a = ap.parse_args()
    root = a.baseline.resolve()
    output = a.output.resolve()
    if not (root / 'go.mod').exists():
        ap.error('baseline must be a repository root containing go.mod')
    if output == root or output.is_relative_to(root) or root.is_relative_to(output):
        ap.error('output must be disjoint from baseline')
    if output.exists():
        ap.error('output already exists; never overwrite a reference')
    variant = root / 'tools/release/m7g-variants' / FILES[a.variant]
    if not variant.is_file():
        ap.error(f'missing variant: {variant}')
    baseline_manifest = production(root)
    if PARSER not in baseline_manifest:
        ap.error('missing baseline parser')
    try:
        shutil.copytree(root, output)
        (output/PARSER).write_bytes(variant.read_bytes())
        candidate_manifest = production(output)
        changed = [str(p) for p in sorted(baseline_manifest.keys() | candidate_manifest.keys())
                   if baseline_manifest.get(p) != candidate_manifest.get(p)]
        if changed != [str(PARSER)]:
            raise ValueError(f'illegal production differences: {changed}')
        for mod in ['go.mod', 'go.sum']:
            if hashfile(root/mod) != hashfile(output/mod):
                raise ValueError(f'changed module metadata: {mod}')
        if 'replace ' in (output/'go.mod').read_text() or (output/'vendor').exists():
            raise ValueError('replace or vendor forbidden')
        print(f'STAGED {a.variant}: only {PARSER} differs')
        print(f'baseline parser:  {baseline_manifest[PARSER]}')
        print(f'candidate parser: {candidate_manifest[PARSER]}')
        print(f'output: {output}')
    except Exception:
        shutil.rmtree(output, ignore_errors=True)
        raise

if __name__ == '__main__':
    try:
        main()
    except (OSError, ValueError) as err:
        sys.exit(f'FAIL: {err}')
