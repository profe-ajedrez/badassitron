#!/usr/bin/env python3
"""M1 supplemental deterministic cases; does not alter the frozen M1 fixture."""
import json
import random
import sys

rng = random.Random(0xBADF00D)
values = ["0", "-0", "1", "-1", "2", "-2", "3", "-3", "7", "10",
          "12.3400", "-12.3400", "1.25", "-1.25", "0.0000000000000000001",
          "18446744073709551615", "18446744073709551616",
          "340282366920938463463374607431768211455", "1e20", "invalid", "0.00"]
for _ in range(70):
    a = rng.randint(-10**24, 10**24)
    scale = rng.randint(0, 19)
    digits=str(abs(a)).zfill(scale+1)
    number=(digits[:-scale]+'.'+digits[-scale:]) if scale else digits
    values.append(('-' if a < 0 else '')+number)

def make(op, **kwargs):
    make.serial += 1
    payload = {'id': f'M1X-{make.serial:05d}', 'op':op}
    payload.update(kwargs)
    print(json.dumps(payload, separators=(',',':')))
make.serial = 0
for op in ('mod', 'quorem'):
    for a in values[:35]:
        for b in values[:12]:
            make(op, a=a, b=b)
    for _ in range(150):make(op,a=rng.choice(values),b=rng.choice(values))
for a in values[:35]:
    for n in (-7,-3,-2,-1,0,1,2,3,7):make('powint',a=a,n=n)
for a in values[:45]:
    for scale in (0,1,2,10,19,20): make('encode_uint128',a=a,scale=scale)
