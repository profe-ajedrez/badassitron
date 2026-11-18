import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('release_api', Path(__file__).with_name('check_api.py'))
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

class TestFrozenAPI(unittest.TestCase):
    def setUp(self):
        self.old = [['dec128', 'func', 'Dec128.Add'], ['dec128/uint128', 'func', 'Uint128.Add']]
        self.unsigned = [{'kind': 'func', 'name': 'Uint128.Add', 'signature': 'func(a Uint128) (Uint128, errors.Error)'}]
        self.current = [
            {'package': 'dec128', 'kind': 'func', 'name': 'Dec128.Add', 'signature': 'ignored'},
            {'package': 'dec128/uint128', **self.unsigned[0]},
        ]
    def test_exact_unsigned_signature(self):
        self.assertEqual(mod.check(self.old, self.unsigned, self.current)[:2], (2, 1))
    def test_missing_historical_declaration(self):
        with self.assertRaisesRegex(ValueError, 'missing historical'):
            mod.check(self.old, self.unsigned, self.current[1:])
    def test_changed_unsigned_signature(self):
        actual = [dict(e) for e in self.current]
        actual[1]['signature'] = 'func(a Uint128) Uint128'
        with self.assertRaisesRegex(ValueError, 'signatures diverged'):
            mod.check(self.old, self.unsigned, actual)
    def test_extra_unsigned_declaration(self):
        actual = self.current + [{'package':'dec128/uint128', 'kind':'func', 'name':'Surprise', 'signature':'func()'}]
        with self.assertRaisesRegex(ValueError, 'signatures diverged'):
            mod.check(self.old, self.unsigned, actual)
    def test_unexported_receiver_ignored(self):
        actual = self.current + [{'package':'dec128', 'kind':'func', 'name':'*sqlTarget.Scan', 'signature':'func()'}]
        self.assertEqual(mod.check(self.old, self.unsigned, actual)[2], [])

if __name__ == '__main__':
    unittest.main()
