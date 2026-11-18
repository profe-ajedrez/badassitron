#!/usr/bin/env python3
"""Independent decimal arithmetic audit for exact operations in oracle JSONL.

Checks only successfully parsed operands and finite results of Add/Sub/Mul.
It does NOT approve an error/NaN or certify division-rounding behavior.
"""
import argparse
from decimal import Decimal, InvalidOperation, localcontext
import json
from pathlib import Path
import sys

def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('cases');ap.add_argument('results'); args=ap.parse_args()
    cases={r['id']:r for r in map(json.loads,Path(args.cases).read_text().splitlines())}
    results={r['id']:r for r in map(json.loads,Path(args.results).read_text().splitlines())}
    count=0;wrong=[]; nan_count=0
    with localcontext() as ctx:
        ctx.prec=300
        for key,c in cases.items():
            if c['op'] not in ('add','sub','mul'):continue
            r=results[key]
            try:a,b=Decimal(c['a']),Decimal(c['b'])
            except InvalidOperation:continue
            if not a.is_finite() or not b.is_finite():continue
            if r['nan']:
                nan_count+=1
                continue
            expected={'add':lambda:a+b,'sub':lambda:a-b,'mul':lambda:a*b}[c['op']]()
            try:actual=Decimal(r['string'])
            except InvalidOperation:
                wrong.append((key,c['op'],'unparseable result',r['string']))
                continue
            count+=1
            if expected!=actual:wrong.append((key,c['op'],str(expected),r['string']))
    print(f'exact_results_checked={count} nan_or_error_skipped={nan_count} arithmetic_mismatches={len(wrong)}')
    for row in wrong[:40]:print('MISMATCH', *row)
    return int(bool(wrong))

if __name__=='__main__':sys.exit(main())
