#!/usr/bin/env python3
"""Check finite supplemental oracle results against exact Python Decimal arithmetic.

NaN is never independently approved: cases returning NaN are reported as skipped.
Checks numeric equality only, not scale, error code, signed zero or SQL behavior.
"""
import argparse
from decimal import Decimal, InvalidOperation, localcontext
import json
from pathlib import Path
import sys

MAX_COEF = (1 << 128) - 1


def check(cases, results):
    if set(cases) != set(results):
        raise ValueError('case/result ID sets differ')
    checked = 0
    skipped = 0
    mismatches = []
    lossy_encodes = 0
    with localcontext() as ctx:
        ctx.prec = 350
        for key, c in cases.items():
            r = results[key]
            try:
                a = Decimal(c.get('a', ''))
                b = Decimal(c.get('b', '')) if c['op'] in ('mod', 'quorem') else None
            except InvalidOperation:
                skipped += 1
                continue
            if not a.is_finite() or (b is not None and not b.is_finite()):
                skipped += 1
                continue
            if r['nan']:
                skipped += 1
                continue
            op = c['op']
            try:
                actual = Decimal(r['string'])
            except InvalidOperation:
                mismatches.append((key, op, 'valid finite number', r['string']))
                continue
            if op in ('mod', 'quorem'):
                if not b:
                    mismatches.append((key, op, 'NaN division by zero', r['string']))
                    continue
                # Truncation toward zero, never floor division for negative operands.
                quotient = int(a / b)
                remainder = a - Decimal(quotient) * b
                expected = remainder if op == 'mod' else Decimal(quotient)
                checked += 1
                if actual != expected:
                    mismatches.append((key, op, str(expected), r['string']))
                if op == 'quorem':
                    extra = r.get('extra')
                    if not extra:
                        mismatches.append((key, 'quorem.remainder', str(remainder), 'missing'))
                    elif extra['nan'] or Decimal(extra['string']) != remainder:
                        mismatches.append((key, 'quorem.remainder', str(remainder), extra['string']))
                    else:
                        checked += 1
            elif op == 'powint' and c['n'] >= 0:
                expected = Decimal(1) if c['n'] == 0 else a ** c['n']
                checked += 1
                if actual != expected:
                    mismatches.append((key, op, str(expected), r['string']))
            elif op == 'encode_uint128':
                if r.get('encode_error'):
                    skipped += 1
                    continue
                expected = a * (Decimal(10) ** c['scale'])
                encoded = int(r['encoded'])
                # Legacy EncodeToUint128 intentionally truncates toward zero
                # when the requested scale is less than the source scale.
                if expected != int(expected):
                    lossy_encodes += 1
                checked += 1
                if int(expected) != encoded or not 0 <= encoded <= MAX_COEF:
                    mismatches.append((key, op, str(int(expected)), str(encoded)))
            else:
                skipped += 1
    return checked, skipped, lossy_encodes, mismatches


def load(path):
    rows = [json.loads(l) for l in Path(path).read_text().splitlines() if l.strip()]
    d = {r['id']: r for r in rows}
    if len(rows) != len(d):
        raise ValueError('duplicate case ID')
    return d


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('cases'); p.add_argument('results')
    args = p.parse_args()
    checked, skipped, lossy, errors = check(load(args.cases), load(args.results))
    print(f'extended_exact_checks={checked} skipped={skipped} lossy_encodes={lossy} mismatches={len(errors)}')
    for item in errors[:30]: print('MISMATCH', *item)
    return bool(errors)


if __name__ == '__main__':
    sys.exit(main())
