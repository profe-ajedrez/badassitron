#!/usr/bin/env python3
"""Check that the documentation homepages link to all user-facing guides.

This is a navigation coverage check, not an HTML-anchor validator.
"""
from pathlib import Path
import re
from urllib.parse import unquote, urlparse

ROOT = Path(__file__).resolve().parents[1]
LINK = re.compile(r'(?<!!)\[[^\]\n]+\]\(([^)]+)\)')


def linked_files(page: Path) -> set[Path]:
    result = set()
    for match in LINK.finditer(page.read_text(encoding='utf-8')):
        uri = urlparse(match.group(1))
        if uri.scheme or not uri.path:
            continue
        result.add((page.parent / unquote(uri.path)).resolve())
    return result


def main() -> int:
    root = linked_files(ROOT / 'README.md')
    docs = linked_files(ROOT / 'docs/README.md')
    examples = linked_files(ROOT / 'examples/README.md')
    entry = ['docs/README.md', 'docs/GETTING_STARTED.md',
             'docs/RECIPES.md', 'examples/README.md']
    guides = ['docs/OPTIMUS.md', 'docs/DECIMAL_BEHAVIOR.md',
              'docs/DECIMAL_TYPES.md', 'docs/ROUNDING_AND_ERRORS.md',
              'docs/DATA_IO.md', 'docs/UNSIGNED_INTEGERS.md',
              'docs/ARCHITECTURE.md', 'docs/GODOC_AND_EXAMPLES.md',
              'docs/CALCULATION_PIPELINE.md']
    missing = [p for p in entry if (ROOT / p).resolve() not in root]
    missing += [p for p in guides if (ROOT / p).resolve() not in docs]
    example_files = sorted((ROOT / 'examples').glob('*/main.go'))
    missing += [str(p.parent.relative_to(ROOT) / 'README.md')
                for p in example_files
                if (p.parent / 'README.md').resolve() not in examples]
    if missing:
        print('FAIL: missing navigation links:', ', '.join(missing))
        return 1
    print(f'PASS: entry points, {len(guides)} guides, and '
          f'{len(example_files)} runnable examples indexed')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
