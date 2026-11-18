#!/usr/bin/env python3
"""Interleaved M7-D/M7-E benchmark, including independent upstream controls.

Inputs must be the untouched M7-D and M7-E repositories (directories or ZIPs).
Builds are compiled once from temporary copies, with identical benchmark sources.
No user source tree is modified. Results are descriptive, not a proof of causality.
"""

import argparse
import csv
import hashlib
import os
from pathlib import Path
import random
import re
import shutil
import statistics
import subprocess
import sys
import tempfile
import zipfile

OPERATIONS = ("ParseShort", "ParseShortMixed", "ParseLong", "QuoRem", "WideQuoRem")
MODES = ("Facade", "UpstreamDirect")
BENCH_REGEX = (
    r"^(BenchmarkMigrationM3FCodec|BenchmarkReleaseM7ETargets|"
    r"BenchmarkMigrationM3ThreeWay)$/(ParseShort|ParseLong|ParseShortMixed|"
    r"QuoRem|WideQuoRem)$/(Facade|UpstreamDirect)$"
)
BENCH_LINE = re.compile(
    r"^Benchmark\S+/(ParseShort|ParseShortMixed|ParseLong|QuoRem|WideQuoRem)/"
    r"(Facade|UpstreamDirect)(?:-\d+)?\s+\d+\s+([\d.]+) ns/op\s+"
    r"(\d+) B/op\s+(\d+) allocs/op$"
)


def run(cmd, cwd=None, env=None, log=None):
    proc = subprocess.run(cmd, cwd=cwd, env=env, text=True, stdout=subprocess.PIPE,
                          stderr=subprocess.STDOUT, check=False)
    if log is not None:
        log.write_text(proc.stdout)
    if proc.returncode:
        raise RuntimeError(f"exit {proc.returncode}: {' '.join(map(str, cmd))}\n{proc.stdout[-5000:]}")
    return proc.stdout


def root_dir(path):
    if (path / 'go.mod').is_file():
        return path
    nested = [f.parent for f in path.glob('*/go.mod')]
    if len(nested) != 1:
        raise ValueError(f"cannot determine repository root from {path}")
    return nested[0]


def stage(path, dst):
    path = Path(path).resolve()
    if path.is_dir():
        shutil.copytree(root_dir(path), dst)
    elif path.suffix.lower() == '.zip':
        extract = dst.parent / (dst.name + '_unpacked')
        with zipfile.ZipFile(path) as z:
            for entry in z.infolist():
                target = (extract / entry.filename).resolve()
                if not target.is_relative_to(extract.resolve()):
                    raise ValueError(f"unsafe ZIP path: {entry.filename}")
                if entry.is_dir():
                    continue
                # No source should ship a symlink in a benchmark ZIP.
                if (entry.external_attr >> 16) & 0o170000 == 0o120000:
                    raise ValueError(f"ZIP symlinks are not allowed: {entry.filename}")
            z.extractall(extract)
        shutil.copytree(root_dir(extract), dst)
    else:
        raise ValueError(f"provide a checkout directory or ZIP file: {path}")
    return dst


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def enforce_single_source_change(baseline, candidate):
    def production(root):
        return {str(p.relative_to(root)): sha(p)
                for p in root.rglob('*.go')
                if not p.name.endswith('_test.go') and 'tools' not in p.relative_to(root).parts}
    a, b = production(baseline), production(candidate)
    diffs = sorted(k for k in a.keys() | b.keys() if a.get(k) != b.get(k))
    allowed = ['dec128/uint128/from_upstream_string.go']
    if diffs != allowed:
        raise ValueError(f"expected only the parser adapter to change, found: {diffs}")
    for f in ['go.mod', 'go.sum']:
        if sha(baseline / f) != sha(candidate / f):
            raise ValueError(f"{f} differs between releases")
    # D did not have the supplementary mixed benchmark. Make tests identical
    # before compiling either binary, without editing the real M7-D archive.
    for filename in ['migration_m7e_targets_bench_test.go', 'migration_m7e_parser_test.go']:
        source = candidate / 'dec128/uint128' / filename
        if not source.is_file():
            raise ValueError(f"required candidate test not found: {filename}")
        target = baseline / 'dec128/uint128' / filename
        if not target.exists():
            shutil.copyfile(source, target)
        elif sha(source) != sha(target):
            raise ValueError(f"benchmark/test mismatch: {filename}")


def stage_local_upstream(root, upstream):
    """Isolated Go 1.23 diagnostics only; no replacement survives packaging."""
    mod = root / 'go.mod'
    contents = mod.read_text()
    contents = re.sub(r'(?m)^go\s+1\.27\s*$', 'go 1.23', contents)
    contents += f'\nreplace github.com/jokruger/dec128 => {upstream}\n'
    mod.write_text(contents)


def parse_bench(raw):
    found = {}
    for line in raw.splitlines():
        match = BENCH_LINE.match(line.strip())
        if match:
            op, mode, ns, bytes_, allocs = match.groups()
            key = (op, mode)
            if key in found:
                raise ValueError(f"duplicate benchmark result: {key}")
            found[key] = (float(ns), int(bytes_), int(allocs))
    expected = {(op, mode) for op in OPERATIONS for mode in MODES}
    if found.keys() != expected:
        raise ValueError(f"benchmark mismatch: missing {sorted(expected-found.keys())}, extra {sorted(found.keys()-expected)}")
    return found


def bootstrap_interval(data, repetitions=20000):
    rng = random.Random(20261118)
    n = len(data)
    samples = sorted(statistics.median(data[rng.randrange(n)] for _ in range(n))
                     for _ in range(repetitions))
    return samples[int(.025 * repetitions)], samples[int(.975 * repetitions)]


def format_report(rows, blocks, benchtime, go_version, local_diagnostic=False):
    details = {}
    for op in OPERATIONS:
        paired, upstream_drift, normalized = [], [], []
        for block in range(1, blocks + 1):
            def median_for(ver, mode):
                vals = [row['ns_op'] for row in rows if row['block'] == block and row['version'] == ver
                        and row['operation'] == op and row['mode'] == mode]
                assert len(vals) == 2, (block, ver, mode, vals)
                return statistics.median(vals)
            da = median_for('M7-D', 'Facade')
            ea = median_for('M7-E', 'Facade')
            du = median_for('M7-D', 'UpstreamDirect')
            eu = median_for('M7-E', 'UpstreamDirect')
            paired.append(100*(ea/da-1))
            upstream_drift.append(100*(eu/du-1))
            normalized.append(100*((ea/eu)/(da/du)-1))
        all_a = [row['ns_op'] for row in rows if row['version'] == 'M7-D' and row['operation'] == op and row['mode'] == 'Facade']
        all_b = [row['ns_op'] for row in rows if row['version'] == 'M7-E' and row['operation'] == op and row['mode'] == 'Facade']
        all_ad = [row['ns_op'] for row in rows if row['version'] == 'M7-D' and row['operation'] == op and row['mode'] == 'UpstreamDirect']
        all_bd = [row['ns_op'] for row in rows if row['version'] == 'M7-E' and row['operation'] == op and row['mode'] == 'UpstreamDirect']
        details[op] = {'D':statistics.median(all_a),'E':statistics.median(all_b),'UD':statistics.median(all_ad),
                       'UE':statistics.median(all_bd), 'paired':statistics.median(paired),
                       'paired_ci':bootstrap_interval(paired), 'normalized':statistics.median(normalized),
                       'drift':statistics.median(upstream_drift), 'wins':sum(v < 0 for v in paired),
                       'pair_values':paired, 'norm_values':normalized}
    text = [
        '# Interleaved M7-D vs M7-E — parser adapter', '',
        f'- Toolchain: `{go_version}`; **{blocks} ABBA/BAAB blocks**, {2*blocks} independent invocations per release.',
        f'- Benchmark duration: `{benchtime}` for each subbenchmark; both binaries compiled once from copies.',
        '- Identical benchmark sources and operands in both builds; direct Jokruger is measured in each invocation.',
        '- The only differing production source file is `dec128/uint128/from_upstream_string.go`.',
        '- Negative facade delta means M7-E faster. Bootstrap interval resamples blocks (descriptive, not a formal guarantee).',
        '',
        '| Operation | M7-D facade | M7-E facade | Paired delta | 95% bootstrap CI | Direct Jokruger drift | Normalized delta | Blocks E wins |',
        '|---|---:|---:|---:|---:|---:|---:|---:|',
    ]
    for op in OPERATIONS:
        d = details[op]
        text.append(f"| `{op}` | {d['D']:.3f} | {d['E']:.3f} | {d['paired']:+.1f}% | "
                    f"[{d['paired_ci'][0]:+.1f}%, {d['paired_ci'][1]:+.1f}%] | "
                    f"{d['drift']:+.1f}% | {d['normalized']:+.1f}% | {d['wins']}/{blocks} |")
    parse = details['ParseShort']
    control_bad = [op for op in ('ParseLong','QuoRem','WideQuoRem') if details[op]['paired'] > 3.0]
    strong_signal = parse['paired'] <= -3 and parse['paired_ci'][1] < 0 and not control_bad
    official_go = bool(re.search(r'go1\.27\.[1-9]\d*\b', go_version))
    duration = re.fullmatch(r'([0-9]+)(ms|s)', benchtime)
    duration_ms = (int(duration[1]) * (1000 if duration[2] == 's' else 1)) if duration else 0
    enough_evidence = official_go and blocks >= 8 and duration_ms >= 250 and not local_diagnostic
    if strong_signal and enough_evidence:
        decision = 'PERFORMANCE-CANDIDATE ONLY: verify complete Go 1.27.1 release suite before promotion'
    elif strong_signal:
        decision = 'DIAGNOSTIC SIGNAL ONLY: insufficient Go version, repetitions, or benchmark duration; NO PROMOTION'
    elif parse['paired'] >= 2 or parse['paired_ci'][0] > 0 or control_bad:
        decision = 'NO PROMOTION: negative parser results or a control regression'
    else:
        decision = 'INCONCLUSIVE: keep M7-D; collect official interleaved data'
    text += ['', f'**Decision gate:** {decision}.', '',
             'This run does not alter the M7-D release baseline. A candidate is not accepted merely because',
             'its facade/upstream percentage improved in separate, non-interleaved campaigns.', '',
             '## Block-level facade changes (%)', '']
    text.append('| Block | ' + ' | '.join(f'`{op}`' for op in OPERATIONS) + ' |')
    text.append('|---|' + '---:|'*len(OPERATIONS))
    for block in range(blocks):
        text.append('| '+str(block+1)+' | '+' | '.join(f"{details[op]['pair_values'][block]:+.2f}%" for op in OPERATIONS)+' |')
    text.extend(['', '## Conditions and limitations', '',
                 'CPU scheduling, thermal state, and code-layout differences can affect sub-nanosecond comparisons.',
                 'Direct Jokruger samples track common environmental drift; they do not remove all confounding.',
                 'Re-run with Go 1.27.1, BENCHTIME at least 250ms and at least eight blocks on the target hardware.',
                 'Functional `tools/release/verify.sh` and license/provenance approval are separate gates.', ''])
    return '\n'.join(text), decision


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--baseline', type=Path, required=True, help='M7-D ZIP or extracted checkout')
    parser.add_argument('--candidate', type=Path, required=True, help='M7-E ZIP or extracted checkout')
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--blocks', type=int, default=8)
    parser.add_argument('--benchtime', default='250ms')
    parser.add_argument('--local-upstream', type=Path,
                        help='LOCAL DIAGNOSTICS ONLY: use Jokruger source snapshot with older Go, never ship replace')
    args = parser.parse_args()
    if args.blocks < 3:
        parser.error('--blocks must be >= 3')
    out = args.output.resolve()
    out.mkdir(parents=True, exist_ok=True)
    version = run(['go','version']).strip()
    official = bool(re.search(r'go1\.27\.[1-9]\d*\b', version))
    if not official and not args.local_upstream:
        parser.error('Go 1.27.1+ required for official comparisons; use --local-upstream only for local diagnostics')
    env = dict(os.environ, GOWORK='off', GOMAXPROCS='1', GOTOOLCHAIN='local')
    with tempfile.TemporaryDirectory(prefix='m7de-ab-') as t:
        root = Path(t)
        D = stage(args.baseline, root / 'baseline')
        E = stage(args.candidate, root / 'candidate')
        enforce_single_source_change(D, E)
        if args.local_upstream:
            upstream = stage(args.local_upstream, root / 'upstream')
            contents = upstream.joinpath('go.mod').read_text()
            contents = re.sub(r'(?m)^go\s+1\.24\s*$', 'go 1.23', contents)
            upstream.joinpath('go.mod').write_text(contents)
            stage_local_upstream(D, upstream)
            stage_local_upstream(E, upstream)
            env['GOPROXY'] = 'off'
            env['GOSUMDB'] = 'off'
        for label, staged in [('M7-D',D),('M7-E',E)]:
            print(f'Compiling and checking {label} using {version}', flush=True)
            run(['go','test','./dec128/uint128','-count=1'], cwd=staged, env=env,
                log=out / f'test_{label}.txt')
            run(['go','test','-c','-o',str(root / f'{label}.test'),'./dec128/uint128'], cwd=staged, env=env,
                log=out / f'compile_{label}.txt')
        rows = []
        with (out / 'raw.txt').open('w') as raw:
            for block in range(1,args.blocks+1):
                order = ['M7-D','M7-E','M7-E','M7-D'] if block % 2 else ['M7-E','M7-D','M7-D','M7-E']
                print(f'Block {block}/{args.blocks}: {"-".join(order)}', flush=True)
                for seq,label in enumerate(order, 1):
                    cmd = [str(root/f'{label}.test'), '-test.run=^$', f'-test.bench={BENCH_REGEX}',
                           f'-test.benchtime={args.benchtime}', '-test.benchmem', '-test.count=1', '-test.cpu=1']
                    stdout = run(cmd, env=env)
                    parsed = parse_bench(stdout)
                    raw.write(f'### block={block} sequence={seq} version={label}\n{stdout}\n')
                    raw.flush()
                    for (op, mode), (ns, byte_count, allocs) in parsed.items():
                        rows.append(dict(block=block, sequence=seq, version=label, operation=op,
                                         mode=mode, ns_op=ns, bytes_per_op=byte_count, allocs_per_op=allocs))
        with (out / 'samples.csv').open('w', newline='') as f:
            writer = csv.DictWriter(f, fieldnames=['block','sequence','version','operation','mode','ns_op','bytes_per_op','allocs_per_op'])
            writer.writeheader()
            writer.writerows(rows)
        if any(row['allocs_per_op'] != 0 or row['bytes_per_op'] != 0 for row in rows):
            raise RuntimeError('allocations detected (see samples.csv)')
        report, decision = format_report(rows, args.blocks, args.benchtime, version, bool(args.local_upstream))
        (out/'comparison.md').write_text(report + '\n')
        (out/'environment.txt').write_text(version+'\n' +
            ('LOCAL DIAGNOSTIC ONLY; temporary source-only replace was used\n' if args.local_upstream else 'OFFICIAL MODULE; no replace\n') +
            f'baseline={args.baseline.resolve()}\ncandidate={args.candidate.resolve()}\n')
        print(report)
        print(f'Files: {out}', flush=True)


if __name__ == '__main__':
    try:
        main()
    except (OSError, ValueError, RuntimeError, zipfile.BadZipFile) as error:
        raise SystemExit(f'FAIL: {error}')
