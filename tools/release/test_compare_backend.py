import importlib.util
from pathlib import Path
import unittest
spec = importlib.util.spec_from_file_location('release_bench', Path(__file__).with_name('compare_backend.py'))
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)

TEXT = '''BenchmarkMigrationM3ThreeWay/Add/Facade-16 100 1.500 ns/op 0 B/op 0 allocs/op
BenchmarkMigrationM3ThreeWay/Add/UpstreamDirect-16 100 1.000 ns/op 0 B/op 0 allocs/op
BenchmarkMigrationM3FCodec/ParseLong/Facade-16 100 4.500 ns/op 0 B/op 0 allocs/op
BenchmarkMigrationM3FCodec/ParseLong/UpstreamDirect-16 100 4.000 ns/op 0 B/op 0 allocs/op
'''
class TestBenchmarkComparison(unittest.TestCase):
    def test_complete_pair(self):
        out=mod.render(TEXT,1)
        self.assertIn('`Add` | 1.500 | 1.000 | +50.0%',out)
        self.assertIn('`ParseLong`',out)
    def test_missing_pair(self):
        with self.assertRaisesRegex(ValueError,'expected'):
            mod.render(TEXT.splitlines()[0]+'\n',1)
    def test_sample_mismatch(self):
        with self.assertRaisesRegex(ValueError,'expected 2'):
            mod.render(TEXT,2)
    def test_empty_rejected(self):
        with self.assertRaisesRegex(ValueError,'no facade'):
            mod.render('not a benchmark',1)

if __name__=='__main__': unittest.main()
