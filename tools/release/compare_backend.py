#!/usr/bin/env python3
"""Compare same-build public Badassitron facade and direct Jokruger medians."""
import argparse
from collections import defaultdict
from pathlib import Path
import re
from statistics import median

ROW = re.compile(r'^BenchmarkMigrationM3(?:ThreeWay|FCodec)/([^/]+)/(Facade|UpstreamDirect)(?:-\d+)?\s+\d+\s+([0-9.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op')


def parse(text):
    measurements = defaultdict(list)
    allocs = defaultdict(set)
    for line in text.splitlines():
        m = ROW.match(line)
        if not m:
            continue
        op, kind, ns, b, n = m.groups()
        measurements[(op, kind)].append(float(ns))
        allocs[(op, kind)].add((int(b), int(n)))
    return measurements, allocs


def render(text, count):
    values, allocs = parse(text)
    operations = sorted({op for op, mode in values})
    if not operations:
        raise ValueError('no facade/upstream benchmark rows found')
    lines = ['# Imported Badassitron facade versus direct Jokruger', '',
             f'Both calls compiled in the same default build. {count} samples per route; medians in ns/op. Positive delta means facade overhead.', '',
             '| Operation | Facade | Direct upstream | Overhead | Allocations (facade/upstream) |',
             '|---|---:|---:|---:|---:|']
    for op in operations:
        facade = values.get((op,'Facade'), [])
        upstream = values.get((op,'UpstreamDirect'), [])
        if len(facade) != count or len(upstream) != count:
            raise ValueError(f'{op}: expected {count} samples each; found {len(facade)} facade and {len(upstream)} upstream')
        a = median(facade)
        b = median(upstream)
        if b == 0:
            raise ValueError(f'{op}: zero baseline')
        fa = allocs[(op,'Facade')]
        up = allocs[(op,'UpstreamDirect')]
        if len(fa) != 1 or len(up) != 1:
            raise ValueError(f'{op}: allocation results inconsistent')
        lines.append(f'| `{op}` | {a:.3f} | {b:.3f} | {(a/b-1)*100:+.1f}% | {next(iter(fa))} / {next(iter(up))} |')
    lines += ['', 'Results are descriptive microbenchmarks, not a pass/fail threshold or an equivalence proof.',
              'Compare consumer workloads separately. Historical baselines remain in M3-F official evidence.', '']
    return '\n'.join(lines)


def main():
    p=argparse.ArgumentParser()
    p.add_argument('input')
    p.add_argument('--samples',type=int,default=7)
    p.add_argument('--output',required=True)
    a=p.parse_args()
    out=render(Path(a.input).read_text(),a.samples)
    Path(a.output).write_text(out)
    print(out)

if __name__=='__main__':
    try: main()
    except ValueError as exc: raise SystemExit(f'FAIL: {exc}')
