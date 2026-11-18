#!/usr/bin/env python3
"""Check packaging hygiene and release-state documentation without network access."""
from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parents[2]


def violations(root: Path):
    mod = (root / 'go.mod').read_text(encoding='utf-8')
    if not re.search(r'^go 1\.27\s*$', mod, flags=re.M):
        yield 'go.mod must retain go 1.27'
    if not re.search(r'^require github\.com/jokruger/dec128 v1\.4\.0\s*$', mod, flags=re.M):
        yield 'go.mod must retain the pinned Jokruger v1.4.0 requirement'
    if re.search(r'^\s*replace(?:\s|\()', mod, flags=re.M):
        yield 'temporary replace directive found'
    for name in ('vendor', 'go.work', 'go.work.sum'):
        if (root / name).exists():
            yield f'unexpected local workspace or vendored path: {name}'
    for path in root.rglob('*'):
        if not path.is_file() and path.name != '__pycache__':
            continue
        if path.name == '__pycache__' or path.suffix in {'.pyc', '.pyo'}:
            yield f'Python cache in release source: {path.relative_to(root)}'
        if path.is_file() and path.suffix == '.zip':
            yield f'nested experiment archive in release source: {path.relative_to(root)}'
    for name in ('README.md', 'NOTICE.md', 'THIRD_PARTY_NOTICES.md',
                 'docs/RELEASE_STATUS.md', 'docs/MIGRATION_ARCHIVE.md',
                 'docs/PROVENANCE_REVIEW.md'):
        if not (root / name).is_file():
            yield f'missing current disclosure: {name}'
    status = (root / 'docs/RELEASE_STATUS.md').read_text(encoding='utf-8') if (root / 'docs/RELEASE_STATUS.md').exists() else ''
    if 'NOT APPROVED FOR' not in status or 'provenance' not in status.lower():
        yield 'current release document must disclose publication/provenance blocker'


def main():
    errors = list(violations(ROOT))
    if errors:
        print('FAIL: repository hygiene')
        print('\n'.join(errors))
        return 1
    print('PASS: no temporary workspace, replace, vendor, Python cache or embedded experiment ZIP')
    print('PASS: current release/provenance disclosures are present')
    return 0


if __name__ == '__main__':
    sys.exit(main())
