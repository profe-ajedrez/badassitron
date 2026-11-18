import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('release_compare_iterations', Path(__file__).with_name('compare-iterations.py'))
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)


def raw(fa, upstream):
    out = []
    for x in (fa, fa, fa):
        out.append(f'BenchmarkMigrationM3ThreeWay/Add/Facade-24 500000 {x} ns/op 0 B/op 0 allocs/op')
    for x in (upstream, upstream, upstream):
        out.append(f'BenchmarkMigrationM3ThreeWay/Add/UpstreamDirect-24 500000 {x} ns/op 0 B/op 0 allocs/op')
    return '\n'.join(out)


class IterationComparisonTests(unittest.TestCase):
    def test_overhead_reduction_is_measured_in_points(self):
        text = mod.render(raw(1.2, 1), raw(1.05, 1), 3)
        self.assertIn('+20.0%', text)
        self.assertIn('+5.0%', text)
        self.assertIn('-15.0 |', text)
    def test_release_labels_are_configurable(self):
        output = mod.render(raw(1.2, 1), raw(1.05, 1), 3,
                            old_label='M7-C', new_label='M7-D')
        self.assertIn('M7-C overhead', output)
        self.assertIn('M7-D overhead', output)
        self.assertNotIn('M7-A overhead', output)

    def test_missing_benchmark_rejected(self):
        with self.assertRaisesRegex(ValueError, 'operation mismatch'):
            mod.render(raw(1.2, 1), raw(1.05, 1) + '\n' + raw(1.01, 1).replace('/Add/', '/MulExact/'), 3)
    def test_sample_mismatch_rejected(self):
        with self.assertRaisesRegex(ValueError, 'sample counts'):
            mod.render(raw(1.2, 1), raw(1.05, 1).splitlines()[0] + '\n', 3)


if __name__ == '__main__':
    unittest.main()
