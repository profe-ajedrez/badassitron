#!/usr/bin/env python3
"""Independent arithmetic checker and supplemental fixture integrity tests."""
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
TOOL = ROOT / 'tools/migration-oracle'
CASES = ROOT / 'testdata/migration/oracle/extended-cases.jsonl'
RESULTS = ROOT / 'testdata/migration/oracle/extended-results.jsonl'
spec = importlib.util.spec_from_file_location('checker', TOOL / 'check_extended.py')
checker = importlib.util.module_from_spec(spec)
spec.loader.exec_module(checker)


class ExtendedOracleTests(unittest.TestCase):
    def test_fixture_and_exact_arithmetic(self):
        checked, skipped, lossy, errors = checker.check(checker.load(CASES), checker.load(RESULTS))
        self.assertEqual((checked, skipped, lossy, errors), (1634, 563, 38, []))

    def test_detects_corrupt_remainder(self):
        cases = checker.load(CASES)
        results = checker.load(RESULTS)
        key = next(k for k, r in results.items() if r['op'] == 'quorem' and not r['nan'] and r.get('extra'))
        results[key]['extra']['string'] = '999999'
        *_, errors = checker.check(cases, results)
        self.assertTrue(any(e[0] == key and e[1] == 'quorem.remainder' for e in errors))

    def test_generation_is_deterministic(self):
        output = subprocess.check_output([sys.executable, str(TOOL / 'generate_extended.py')])
        self.assertEqual(output, CASES.read_bytes())


if __name__ == '__main__':
    unittest.main()
