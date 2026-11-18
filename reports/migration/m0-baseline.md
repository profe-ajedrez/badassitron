# M0 execution record

**Input SHA-256:** `d6628a2fd4852d06343b95ead5b8e0351d7994acd401921faeb38eb60b6bf7df`

- Source module declaration: Go 1.27 (left unchanged in this deliverable).
- Local diagnostic compiler: Go 1.23.2 linux/amd64, Intel Xeon Platinum 8573C.
- Local dependency mode: `GOWORK=off GOPROXY=off GOTOOLCHAIN=local`.
- Exported declarations captured: **1,033**, across 10 public package paths;
  includes declarations beyond the decimal migration's immediate scope.
- Normal suite: **PASS**, both before and after M0 additions under the
  temporary Go 1.23 module-version workaround; see
  [`m0-tests-after.txt`](m0-tests-after.txt).
- Selected migration contract tests: **PASS** with the same workaround.
- `go vet ./...`: **PASS** with the same workaround; `m0-vet.exit = 0`.
- Fast-path smoke benchmark: **COMPLETED**, `m0-fastpath.exit = 0`;
  two short repetitions, `m0-fastpath.txt`.
- `go test -race ./dec128 ./dec128/uint128 ./dec128/errors`: **PASS**
  under the temporary Go 1.23 module-version workaround, see
  `m0-race-dec128.txt`.
- `go test -race ./...`: **INCOMPLETE**; execution reached packages before
  sandbox time limit, but did not finish; do not claim PASS.
- Toolchain Go 1.27 and Jokruger v1.4.0: **UNAVAILABLE** in this execution
  environment; remote download was blocked.
- API snapshot and contract docs: **WRITTEN**; full review/sign-off pending.
- M1 adapter or M2 replacement: **NOT STARTED**; cannot verify upstream code
  or run dependency tests here.

**Gate status: M0 PARTIAL / NOT GREEN.** Missing complete race, full golden
coverage, upstream provenance review and approval of unresolved mappings.
