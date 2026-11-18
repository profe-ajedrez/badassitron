#!/usr/bin/env python3
"""Regression checks for the JSONL comparison gate (stdlib unittest)."""
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / 'tools/migration-oracle/compare.py'
BASE = ROOT / 'testdata/migration/oracle/legacy-results.jsonl'

class CompareTests(unittest.TestCase):
    def compare(self, left, right, approved=None):
        args=[sys.executable,str(SCRIPT),str(left),str(right)]
        if approved:args+=['--approved',str(approved)]
        return subprocess.run(args, capture_output=True,text=True)

    def test_identical(self):
        self.assertEqual(self.compare(BASE,BASE).returncode,0)

    def test_mismatch_fails(self):
        with tempfile.TemporaryDirectory() as td:
            lines=BASE.read_text().splitlines()
            row=json.loads(lines[0]);row['string']='wrong';lines[0]=json.dumps(row)
            file=Path(td)/'changed.jsonl';file.write_text('\n'.join(lines)+'\n')
            self.assertEqual(self.compare(BASE,file).returncode,1)

    def test_precise_approved_difference(self):
        with tempfile.TemporaryDirectory() as td:
            lines=BASE.read_text().splitlines()
            row=json.loads(lines[0]); old=row['string'];row['string']='approved-change';lines[0]=json.dumps(row)
            file=Path(td)/'changed.jsonl';file.write_text('\n'.join(lines)+'\n')
            approval=Path(td)/'approval.json'
            approval.write_text(json.dumps([{'id':row['id'],'field':'string','decision':'D03',
                                            'baseline':old,'candidate':'approved-change'}]))
            self.assertEqual(self.compare(BASE,file,approval).returncode,0)
            self.assertEqual(self.compare(BASE,BASE,approval).returncode,1,
                             'stale approval must fail')

if __name__=='__main__':unittest.main()
