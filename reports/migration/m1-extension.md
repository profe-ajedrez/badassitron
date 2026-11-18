# M1 extension — safety corrections and second differential corpus

**Status:** M1 extended; Jokruger bridge remains blocked; M2 not started.

## Inputs and scope

- Starting artifact: `badassitron_M1_oracle_partial.zip`, not an earlier repository snapshot.
- Original M1 corpus **unchanged**: `cases.jsonl` and `legacy-results.jsonl`, 3,518 records.
- Additional independent fixture: `extended-cases.jsonl` and `extended-results.jsonl`, **1,725** records, generated with fixed seed `0xBADF00D`.
- Combined corpus: **5,243** observed records, with no assertion of exhaustive coverage.
- Additional operations: `Mod`, `QuoRem` (quotient plus independently captured remainder), `PowInt`, `EncodeToUint128`. Existing M1 operations remain available.

## Results

| Gate | Observation |
| --- | --- |
| Original M1 differential recheck | 3,518 cases; 0 unexpected differences |
| Supplemental fixture capture | 1,725 cases successfully executed |
| Independent Python Decimal audit of supplemental fixture | 1,634 numeric checks; 0 mismatches |
| Results not verified by independent numeric check | 563 skips (including NaN and out-of-scope operations); not automatically approved |
| Lossy unsigned encodings | 38 cases; existing `EncodeToUint128` truncation behavior, not treated as exact conversion |
| Python oracle tests | 6 PASS |
| Full `go test ./... -count=1` | PASS (local Go 1.23.2 workaround) |
| `go vet ./...` | PASS (local Go 1.23.2 workaround) |
| `go test -race ./dec128/...` | PASS (local Go 1.23.2 workaround) |
| Markdown relative links | 156 PASS before this document was added |

`check_extended.py` independently checks **values**, not the correctness of error status,
scale, negative zero, binary encoding, driver behavior, or whether an NaN was deserved.
The fixture itself preserves legacy observations, including deliberate truncation.

## Approved fixes delivered

- **D05:** `errors.Error.Value()` returns `invalid state code` for unrecognized codes, without indexing outside the table. Historical numeric codes remain unchanged. `Error()` now also handles invalid codes and `None` without panicking.
- **D04:** `PowInt(math.MinInt)` no longer negates `MinInt` directly; exponent magnitude is decomposed safely. This fixes termination, but does **not** promise a finite mathematical result when the decimal capacity is exceeded.

New tests exercise D04 and D05. The 3,518-case frozen M1 fixture is unchanged, because it does not contain the newly corrected edge cases. These approved changes require independent Go 1.27 verification.

## Reproduction

```sh
python3 tools/migration-oracle/generate_extended.py > testdata/migration/oracle/extended-cases.jsonl
GOWORK=off go run ./tools/migration-oracle < testdata/migration/oracle/extended-cases.jsonl > /tmp/extended-candidate.jsonl
python3 tools/migration-oracle/compare.py testdata/migration/oracle/extended-results.jsonl /tmp/extended-candidate.jsonl
python3 tools/migration-oracle/check_extended.py testdata/migration/oracle/extended-cases.jsonl /tmp/extended-candidate.jsonl
python3 -m unittest discover -s tools/migration-oracle -p 'test_*.py'
```

Do **not** substitute the new fixture for the immutable original M1 baseline. The extended fixture is supplementary and reflects D04/D05 corrections. For future Jokruger candidates, run the identical oracle command separately in the legacy and candidate modules, and use explicitly approved field-level deviations only.

## Blockers and decisions

- External access cannot resolve `proxy.golang.org`, so the pinned Jokruger v1.4.0 source, checksum and license/notice files cannot yet be compiled and audited in this environment.
- Go 1.27 toolchain also cannot be installed here; published `go 1.27` remains intact.
- Therefore the upstream bridge, atomic `dec128` replacement, and all performance claims about that replacement remain **unverified / not implemented**.
- `EncodeToUint128` truncation is already documented in `docs/OPTIMUS_STAGE.md`. This is a compatibility observation; redesigning it as a strict converter is not authorized by M1.
