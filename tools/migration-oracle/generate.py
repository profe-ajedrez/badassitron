#!/usr/bin/env python3
"""Produce deterministic, boundary-heavy dec128 operation cases (stdlib only)."""
import json
import random
import sys

rng = random.Random(0xBAD128)
values = [
    '', '0', '-0', '0.0', '-0.00', '5.', '+1.0', '1', '-1', '2', '-2',
    '12.3400', '-12.3400', '1.235', '-1.235', '2.125', '-2.125',
    '0.0000000000000000001', '-0.0000000000000000001',
    '0.0000000001', '100000000000000000000',
    '18446744073709551615', '18446744073709551616',
    '340282366920938463463374607431768211455',
    '340282366920938463463374607431768211456',
    '0.1234567890123456789', '-0.1234567890123456789',
    '1.0000000000000000000', '3.1415926535897932384',
    'not-a-decimal', '.', '1..0', '1e2', ' 1',
]
for _ in range(160):
    digits = rng.randint(1, 38)
    s = ''.join(str(rng.randint(0, 9)) for _ in range(digits))
    if rng.getrandbits(1): s = '-' + s
    if rng.getrandbits(1):
        pos = rng.randint(1, digits)
        offset = 1 if s.startswith('-') else 0
        s = s[:pos+offset] + '.' + s[pos+offset:]
    values.append(s)

def emit(op, a='', b='', scale=0, mode=0, src_type=''):
    emit.counter += 1
    print(json.dumps({'id': f'M1-{emit.counter:05d}', 'op':op, 'a':a, 'b':b,
                      'scale':scale, 'mode':mode, 'src_type':src_type},separators=(',',':')))
emit.counter=0
for a in values: emit('parse',a)
for op in ['add','sub','mul','div','compare']:
    for _ in range(180): emit(op, rng.choice(values),rng.choice(values))
    for a,b in [('0','0'),('1','0'),('1','3'),('1e20','1'),('0.0000000001','0.0000000001'),
                ('18446744073709551615','18446744073709551615'),
                ('340282366920938463463374607431768211455','2')]: emit(op,a,b)
for op in ['canonical','round','roundbank','rescale']:
    for a in values[:70]:
        for scale in [0,2,10,19,20]: emit(op,a,scale=scale)
for op in ['mulround','divround','mulstatus','divstatus']:
    for _ in range(240):
        emit(op,rng.choice(values),rng.choice(values),rng.choice([0,2,10,19,20]),rng.randrange(0,9))
for code in range(11): emit('nan',str(code))
for typ, vals in {'string':['1.2300','invalid'], 'bytes':['3.1400','invalid'],
                  'int64':['42','-7'], 'float64':['1.25','-2.5'], 'nil':[''], 'unsupported':['']}.items():
    for a in vals:emit('scan',a,src_type=typ)
for op in ['unmarshaljson','unmarshaltext']:
    for a in (['"12.3400"','null','0','"invalid"','""'] if op=='unmarshaljson' else ['12.3400','','invalid']): emit(op,a)
