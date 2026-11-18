import importlib.util
from pathlib import Path
import tempfile
import unittest

spec = importlib.util.spec_from_file_location('release_backend', Path(__file__).with_name('check_backend.py'))
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

class TestBackendGuard(unittest.TestCase):
    def test_check_live_tree(self):
        self.assertEqual(list(mod.violations(mod.ROOT)), [])
    def test_reintroduced_engine_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root/'go.mod').write_text('module x\n\ngo 1.27\n\nrequire github.com/jokruger/dec128 v1.4.0\n')
            (root/'dec128/uint128').mkdir(parents=True)
            (root/'dec128/arithmetic.go').write_text('package dec128\n')
            errors = list(mod.violations(root))
            self.assertTrue(any('historical engine was reintroduced' in x for x in errors))
    def test_replace_rejected(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root/'go.mod').write_text('module x\n\ngo 1.27\n\nrequire github.com/jokruger/dec128 v1.4.0\n\nreplace x => ../x\n')
            (root/'dec128/uint128').mkdir(parents=True)
            errors = list(mod.violations(root))
            self.assertTrue(any('replace' in x for x in errors))

if __name__ == '__main__': unittest.main()
