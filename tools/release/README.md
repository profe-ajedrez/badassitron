# Release verification (maintainers)

The release tooling verifies the source and tests; it does **not** approve
historical copyright ownership or authorize publication.

## Verify a candidate

Run from the **exact source checkout** with Go 1.27.1 or newer and access to
the officially tagged `github.com/jokruger/dec128 v1.4.0` module:

```sh
GOWORK=off bash tools/release/verify.sh
```

The script checks dependency/module identity and retained upstream license
bytes, the single-backend architecture, public API signatures, tests,
vet, race behavior, frozen differential oracles, the external Go client,
repository hygiene, documentation links, example index coverage, and the
output of every runnable example. It fails if the required compiler or
module cannot be obtained.

To check documentation and example output independently:

```sh
python3 tools/check_doc_links.py
python3 tools/check_doc_navigation.py
GOWORK=off python3 tools/check_example_outputs.py
```

`go test ./examples/...` only compiles example programs. The output checker
executes each program and compares the result byte-for-byte with the
`Expected output` block in that program's README.

## Optional benchmarks

```sh
BENCHTIME=250ms COUNT=7 bash tools/release/bench-backend.sh reports/backend-benchmark
```

These same-build results compare selected facade operations with directly
imported Jokruger. Timing depends on Go version, CPU, sample size, and input
cases. It is **not** a general performance guarantee. Historical parser A/B
experiments and their measurement tools remain archived for engineering
review; they do not select an alternate production backend.

## Compatibility and release boundaries

Deprecated rollback build tags do not enable the former arithmetic engine.
The default decimal API and historic integer error codes remain supported by
the same imported backend. The technical gate has to be rerun **after** any
source or packaging changes, including documentation updates that affect
examples. A passing command is not a legal provenance clearance.

[Release checklist](../../docs/RELEASE_STATUS.md) ·
[Third-party notices](../../THIRD_PARTY_NOTICES.md) ·
[Engineering history](../../docs/MIGRATION_ARCHIVE.md)
