# M1 differential harness — partial implementation

**Status:** harness implemented and exercised; upstream bridge blocked.

- Source: M0 `badassitron_M0_implemented.zip`.
- Cases: 3,518 deterministic JSONL requests.
- Result records: 3,518; no execution errors.
- Self-comparison: 0 unexpected differences, 0 approvals.
- Independent Python Decimal audit: **389** finite exact Add/Sub/Mul results checked; **0** arithmetic mismatches; **159** NaN/error cases skipped rather than approved.
- Hash `cases.jsonl`: `50b1ffb7b3af9343ac49202be9b8129d9f7721748b4d617c46fb6a08afc93de8`.
- Hash `legacy-results.jsonl`: `76deba9a48962b9a7f2cb78b56bb59d2f24b3683ab538e553ebd95c261d7610f`.
- Toolchain: Go 1.23.2, temporary `go.mod` 1.23 adjustment restored to 1.27.
- Full `go test ./... -count=1`: PASS in that temporary Go 1.23 environment.
- Markdown local-link validator: PASS (156 links in 41 files).
- `go vet ./...`: PASS with the temporary Go 1.23 adjustment.
- `python3 -m unittest discover -s tools/migration-oracle -p "test_*.py"`: 3 PASS.
- Dual-process identical-checkout test: 3,518/3,518 records identical.
- Deliberately corrupted candidate record: correctly rejected by comparator.
- Network check: `proxy.golang.org` DNS failed; `github.com/jokruger/dec128@v1.4.0` could not be retrieved and **was not added**.

## Scope

The `tools/migration-oracle` tool records observable public behavior,
including parse, arithmetic, error, scale, signed zero, JSON/text/SQL Valuer,
and `WithStatus`. It runs in isolated processes for independent legacy and
candidate module roots. Numeric correctness still needs `math/big` as the
second oracle. The baseline deliberately records some old defects, so a
future candidate should use case+field-specific approved differences.

## Remaining M1 gates

1. Resolve and inspect the real v1.4.0 module and its license/notice.
2. Verify API shape and error vocabulary against the pinned version.
3. Implement a real, compilable Jokruger bridge independently from the
   production `dec128` package.
4. Compare separately compiled legacy and candidate roots, on a Go 1.27
   host, with exact approved corrections and the independent `math/big` oracle.
5. Expand goldens for binary serialization and SQL driver integration.

**No engine replacement or width-tier modifications occurred.**
