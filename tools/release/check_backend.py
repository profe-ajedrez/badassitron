#!/usr/bin/env python3
"""Fail closed if old arithmetic, vendoring or module rewrites enter release source."""
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DECLARED_GO = 'go 1.27'
MODULE = 'github.com/jokruger/dec128 v1.4.0'
ABSENT = [
    *(f'dec128/{name}.go' for name in ('arithmetic','const','dec128','from','internal','rescale_round','round_ops','rounding','sql_scan','sqrt_round','state','tools')),
    *(f'dec128/uint128/{name}.go' for name in ('arithmetic','bitwise','tools','from_legacy_string','to_legacy_string')),
]
ACTIVE = [*(f'dec128/{name}.go' for name in ('m2_candidate','m2_arithmetic','m2_missing','m2_parse','m2_remainder','m2_round')),
          'dec128/uint128/m3_adapter.go', 'dec128/uint128/from_upstream_string.go',
          'dec128/uint128/to_upstream_string.go']


def violations(root):
    mod = (root/'go.mod').read_text()
    if not re.search(r'^go 1\.27(?:\.\d+)?$', mod, re.M):
        yield 'go.mod must declare Go 1.27'
    if MODULE not in mod:
        yield 'go.mod must require Jokruger v1.4.0'
    if re.search(r'^\s*replace(?:\s|\()', mod, re.M):
        yield 'release go.mod cannot contain replace directives'
    for marker in ('vendor', 'go.work', 'go.work.sum'):
        if (root/marker).exists():
            yield f'prohibited release workspace/vendor path: {marker}'
    for p in root.rglob('*.go'):
        if '/third_party_licenses/' in str(p) or '/vendor/' in str(p):
            yield f'embedded upstream source: {p.relative_to(root)}'
        if p.is_file() and not p.name.endswith('_test.go') and p.relative_to(root).as_posix().startswith('dec128/') and p.name != 'legacy_build_tags_unsupported.go':
            if re.search(r'^//go:build\s+.*badassitron_(?:legacy|uint128_legacy)', p.read_text(), re.M):
                yield f'legacy selector remains active: {p.relative_to(root)}'
    for rel in ABSENT:
        if (root/rel).exists():
            yield f'historical engine was reintroduced: {rel}'
    for rel in ACTIVE:
        path = root/rel
        if not path.is_file():
            yield f'production source missing: {rel}'
        elif path.read_text().startswith('//go:build'):
            yield f'production source unexpectedly build-tagged: {rel}'
    if (root/'dec128/uint128/m3_adapter.go').is_file() and 'github.com/jokruger/dec128/uint128' not in (root/'dec128/uint128/m3_adapter.go').read_text():
        yield 'public uint128 no longer imports Jokruger arithmetic'
    if (root/'dec128/m2_candidate.go').is_file() and 'github.com/jokruger/dec128' not in (root/'dec128/m2_candidate.go').read_text():
        yield 'public dec128 no longer imports Jokruger'


def main():
    errors = list(violations(ROOT))
    if errors:
        sys.exit('FAIL:\n' + '\n'.join(errors))
    print('PASS: Go 1.27 + imported Jokruger v1.4.0; no replace/vendor/workspace')
    print('PASS: one production 128-bit engine; rollback sources absent; unsupported tags fail closed')

if __name__ == '__main__': main()
