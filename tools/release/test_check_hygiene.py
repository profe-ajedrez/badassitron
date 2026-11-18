"""Unit checks for M7-H release hygiene."""
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest

import check_hygiene


class HygieneTests(unittest.TestCase):
    def make_root(self, path: Path):
        (path / 'go.mod').write_text(
            'module example.com/m\n\ngo 1.27\n\nrequire github.com/jokruger/dec128 v1.4.0\n')
        for filename in ('README.md', 'NOTICE.md', 'THIRD_PARTY_NOTICES.md',
                         'docs/RELEASE_STATUS.md', 'docs/MIGRATION_ARCHIVE.md',
                         'docs/PROVENANCE_REVIEW.md'):
            file = path / filename
            file.parent.mkdir(exist_ok=True, parents=True)
            file.write_text('NOT APPROVED FOR PUBLICATION; provenance pending\n')

    def test_clean_tree(self):
        with TemporaryDirectory() as temp:
            root = Path(temp)
            self.make_root(root)
            self.assertEqual(list(check_hygiene.violations(root)), [])

    def test_rejects_python_cache_and_nested_zip(self):
        with TemporaryDirectory() as temp:
            root = Path(temp)
            self.make_root(root)
            (root / 'scratch.pyc').write_bytes(b'')
            (root / 'model.zip').write_bytes(b'')
            errors = list(check_hygiene.violations(root))
            self.assertTrue(any('cache' in e for e in errors))
            self.assertTrue(any('nested' in e for e in errors))

    def test_rejects_replace_and_missing_disclosure(self):
        with TemporaryDirectory() as temp:
            root = Path(temp)
            self.make_root(root)
            with (root / 'go.mod').open('a') as f:
                f.write('\nreplace github.com/jokruger/dec128 => ../dec128\n')
            (root / 'docs/RELEASE_STATUS.md').write_text('No blockers\n')
            errors = list(check_hygiene.violations(root))
            self.assertTrue(any('replace' in e for e in errors))
            self.assertTrue(any('publication/provenance' in e for e in errors))


if __name__ == '__main__':
    unittest.main()
