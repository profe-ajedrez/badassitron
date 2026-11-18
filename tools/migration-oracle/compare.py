#!/usr/bin/env python3
"""Compare independently executed JSONL oracle streams, reporting all differences.

No default exemptions: every intentional change needs an explicit reviewed
ID and field in the opt-in exceptions file, rather than a broad operation mask.
"""
import argparse
import json
import sys
from pathlib import Path


def load(path):
    by_id = {}
    for n, line in enumerate(Path(path).read_text().splitlines(), 1):
        if not line.strip():
            continue
        row = json.loads(line)
        key = row['id']
        if key in by_id:
            raise ValueError(f'{path}:{n}: duplicate id {key}')
        by_id[key] = row
    return by_id


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('baseline')
    p.add_argument('candidate')
    p.add_argument('--approved', help='JSON list of {id,field,decision,baseline,candidate} entries')
    args = p.parse_args()
    a, b = load(args.baseline),load(args.candidate)
    approved = {}
    if args.approved:
        for x in json.loads(Path(args.approved).read_text()):
            k=(x['id'], x['field'])
            if k in approved:raise ValueError(f'duplicate approval: {k}')
            if not x['decision'] or not x['decision'].startswith('D'):
                raise ValueError(f'approval requires a decision ID: {k}')
            approved[k]=x
    differences=[]; accepted=[]
    for key in sorted(a.keys()|b.keys()):
        if key not in a or key not in b:
            differences.append((key,'record',a.get(key),b.get(key)))
            continue
        for field in sorted(a[key].keys()|b[key].keys()):
            left,right=a[key].get(field),b[key].get(field)
            if left == right:continue
            approval=approved.get((key,field))
            if approval and approval['baseline']==left and approval['candidate']==right:
                accepted.append((key,field,approval['decision']))
            else:
                differences.append((key,field,left,right))
    print(f'baseline_cases={len(a)} candidate_cases={len(b)} approved_differences={len(accepted)} unexpected_differences={len(differences)}')
    for item in accepted[:15]:print('APPROVED', *item)
    for item in differences[:50]:print('DIFF', *[repr(x) for x in item])
    if len(differences)>50:print(f'... {len(differences)-50} additional differences')
    # An outdated exception must not silently survive changes.
    unused=set(approved)-{(x[0],x[1]) for x in accepted}
    for key in sorted(unused):print('UNUSED APPROVAL',key)
    return int(bool(differences or unused))

if __name__=='__main__':sys.exit(main())
