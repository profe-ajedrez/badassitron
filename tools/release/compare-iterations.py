#!/usr/bin/env python3
"""Compare two same-build facade/upstream benchmark runs by normalized overhead.

Both runs have independent direct-upstream controls. Subtracting percentage
points describes how facade overhead changed; it does not prove significance.
"""
import argparse
from pathlib import Path
from statistics import median

import importlib.util

_comp_spec = importlib.util.spec_from_file_location('release_compare_backend', Path(__file__).with_name('compare_backend.py'))
_comp_mod = importlib.util.module_from_spec(_comp_spec)
_comp_spec.loader.exec_module(_comp_mod)
parse = _comp_mod.parse


def collect(raw, samples):
    measurements, allocs = parse(raw)
    operations = {op for op, _ in measurements}
    if not operations:
        raise ValueError('no facade/upstream benchmark rows')
    result = {}
    for op in operations:
        facade, direct = (measurements.get((op, mode), []) for mode in ('Facade', 'UpstreamDirect'))
        if len(facade) != samples or len(direct) != samples:
            raise ValueError(f'{op}: sample counts {len(facade)}/{len(direct)} (expected {samples})')
        if not median(direct):
            raise ValueError(f'{op}: zero direct benchmark time')
        fa_alloc, up_alloc = allocs[(op, 'Facade')], allocs[(op, 'UpstreamDirect')]
        if len(fa_alloc) != 1 or len(up_alloc) != 1:
            raise ValueError(f'{op}: inconsistent allocation counts')
        result[op] = median(facade), median(direct), next(iter(fa_alloc)), next(iter(up_alloc))
    return result


def render(old_raw, new_raw, samples, old_label="M7-A", new_label="M7-C"):
    old = collect(old_raw, samples)
    new = collect(new_raw, samples)
    if old.keys() != new.keys():
        raise ValueError(f'operation mismatch: old-only={sorted(old.keys()-new.keys())} new-only={sorted(new.keys()-old.keys())}')
    lines = [
        '# Release adapter iteration comparison', '',
        f'{samples} samples per route. Each build is normalized to its own direct Jokruger results.',
        'Negative overhead delta (percentage points) indicates improvement. This is descriptive, not a statistical guarantee.', '',
        f'| Operation | {old_label} facade | {old_label} direct | {old_label} overhead | {new_label} facade | {new_label} direct | {new_label} overhead | Delta (pp) |',
        '|---|---:|---:|---:|---:|---:|---:|---:|',
    ]
    for op in sorted(old):
        fa0, up0, a0, b0 = old[op]
        fa1, up1, a1, b1 = new[op]
        before, after = 100*(fa0/up0-1), 100*(fa1/up1-1)
        lines.append(f'| `{op}` | {fa0:.3f} | {up0:.3f} | {before:+.1f}% | {fa1:.3f} | {up1:.3f} | {after:+.1f}% | {after-before:+.1f} |')
        if a0 != a1 or b0 != b1:
            lines.append(f'| Allocation change `{op}` | {a0} | {b0} | | {a1} | {b1} | | |')
    lines += ['', 'The builds may use different code generation and run at different CPU frequencies. ',
              'Use interleaved repetitions and consumer benchmarks before accepting a performance claim.', '']
    return '\n'.join(lines)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('old_raw', type=Path)
    parser.add_argument('new_raw', type=Path)
    parser.add_argument('--samples', type=int, default=7)
    parser.add_argument('--old-label', default='M7-A')
    parser.add_argument('--new-label', default='M7-C')
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    result = render(args.old_raw.read_text(), args.new_raw.read_text(), args.samples,
                    args.old_label, args.new_label)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(result)
    print(result)


if __name__ == '__main__':
    try:
        main()
    except ValueError as exc:
        raise SystemExit(f'FAIL: {exc}')
