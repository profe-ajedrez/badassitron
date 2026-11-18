#!/usr/bin/env python3
"""Local tests for the interleaved release benchmark harness."""
import importlib.util
from pathlib import Path
import tempfile
import unittest
import zipfile

PATH = Path(__file__).with_name('bench-interleaved.py')
SPEC = importlib.util.spec_from_file_location('bench_interleaved', PATH)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class InterleavedBenchmarkTests(unittest.TestCase):
    def test_go_benchmark_cpu_suffix_optional(self):
        for name in ('Facade', 'UpstreamDirect'):
            for suffix in ('', '-168'):
                line = f'BenchmarkMigrationM3FCodec/ParseShort/{name}{suffix}  123456  7.12 ns/op  0 B/op  0 allocs/op'
                match = MODULE.BENCH_LINE.match(line)
                self.assertIsNotNone(match)
                self.assertEqual(match.groups(), ('ParseShort', name, '7.12', '0', '0'))

    def test_parse_requires_all_expected_routes(self):
        rows = '\n'.join(
            f'BenchmarkA/{op}/{mode}-8  100  5.0 ns/op  0 B/op  0 allocs/op'
            for op in MODULE.OPERATIONS for mode in MODULE.MODES
        )
        parsed = MODULE.parse_bench(rows)
        self.assertEqual(len(parsed), len(MODULE.OPERATIONS) * 2)
        with self.assertRaises(ValueError):
            MODULE.parse_bench(rows.splitlines()[0])

    def test_blocks_and_decision(self):
        rows = []
        for block in range(1,5):
            for version in ('M7-D', 'M7-E'):
                for sequence in range(2):
                    for op in MODULE.OPERATIONS:
                        for mode in MODULE.MODES:
                            ns = 10.0
                            if op in ('ParseShort', 'ParseShortMixed') and version == 'M7-E' and mode == 'Facade':
                                ns = 9.0
                            rows.append(dict(block=block, version=version, operation=op, mode=mode,
                                             sequence=sequence, ns_op=ns))
        report, decision = MODULE.format_report(rows, 4, '100ms', 'go1.27.1')
        self.assertIn('DIAGNOSTIC SIGNAL ONLY', decision)
        self.assertIn('`ParseShort`', report)
        # Eight blocks and a 250ms run are needed for candidate status.
        full = []
        for new_block in range(1,9):
            for row in rows:
                if row['block'] == ((new_block - 1) % 4) + 1:
                    full.append(dict(row, block=new_block))
        _, accepted = MODULE.format_report(full, 8, '250ms', 'go1.27.1 linux/amd64')
        self.assertIn('PERFORMANCE-CANDIDATE', accepted)
        _, diagnostic = MODULE.format_report(full, 8, '250ms', 'go1.27.1 linux/amd64', True)
        self.assertIn('DIAGNOSTIC SIGNAL ONLY', diagnostic)
        # Regression in the common division path blocks acceptance.
        for row in rows:
            if row['operation'] == 'QuoRem' and row['mode'] == 'Facade' and row['version'] == 'M7-E':
                row['ns_op'] = 10.5
        _, decision = MODULE.format_report(rows, 4, '100ms', 'go1.27.1')
        self.assertIn('NO PROMOTION', decision)

    def test_zip_path_traversal_rejected(self):
        with tempfile.TemporaryDirectory() as td:
            z = Path(td)/'attack.zip'
            with zipfile.ZipFile(z, 'w') as archive:
                archive.writestr('../outside.txt', 'malicious')
            with self.assertRaises(ValueError):
                MODULE.stage(z, Path(td)/'staged')
            self.assertFalse((Path(td)/'outside.txt').exists())

    def test_staged_baseline_contains_only_allowed_change(self):
        with tempfile.TemporaryDirectory() as td:
            d, e = Path(td)/'D', Path(td)/'E'
            for root in (d,e):
                (root/'dec128/uint128').mkdir(parents=True)
                (root/'go.mod').write_text('module test\ngo 1.27\n')
                (root/'go.sum').write_text('same\n')
                (root/'dec128/uint128/migration_m7e_parser_test.go').write_text('package uint128')
                (root/'dec128/uint128/migration_m7e_targets_bench_test.go').write_text('package uint128_test')
            (d/'dec128/uint128/from_upstream_string.go').write_text('package uint128\n// old')
            (e/'dec128/uint128/from_upstream_string.go').write_text('package uint128\n// new')
            MODULE.enforce_single_source_change(d, e)
            (e/'dec128/uint128/extra.go').write_text('package uint128')
            with self.assertRaises(ValueError):
                MODULE.enforce_single_source_change(d,e)


if __name__ == '__main__':
    unittest.main()
