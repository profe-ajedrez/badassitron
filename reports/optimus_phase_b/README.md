# Phase B execution evidence

Source repository: user-provided Optimus v3 Phase A ZIP. All commands ran
without contacting the network, on Go 1.23.2 with an external compatible
module file. Source `go.mod` remains unchanged.

- `comparable_baseline_benchmarks.txt`: fresh v3 source plus only the same
  Phase B benchmark file used in v4. `-cpu=1 -count=3 -benchtime=250ms`.
- `final_phaseB_benchmarks.txt`: updated version of Mod/QuoRem measured with
  `-cpu=1 -count=3 -benchtime=300ms` (the later power adjustment does not
  affect these operations).
- `power_before_optimization.txt`, `power_candidate2.txt`: equivalent
  uncommon-power benchmark before/after, `-benchtime=400ms -count=4`.
- `pow_coarse_baseline.txt`, `pow_coarse_updated.txt`: `1.01^5000` benchmark,
  `-benchtime=350ms -count=3`.
- `phaseB_differential_final.txt`: independent deterministic tests after
  all numerical changes.
- `phaseB_fuzz.txt`: 142,028 native fuzz executions; fuzz campaign ran before
  the later power-only adjustment, which leaves Mod/QuoRem unchanged.
- `final_phaseA_fuzz.txt`: fresh Phase A arithmetic fuzzing after the power
  optimization, 41,326 executions.
- `full_optimus_race.txt`, `related_race.txt`: complete race runs PASS.
- `final_core_test.txt`, `final_optimus_test.txt`, `final_vet.txt`:
  related package tests, final Optimus tests, and vet PASS.
- `final_global_test.txt`: unrelated pre-existing `withfloat64` build failure.
- `api_baseline.txt`, `api_phaseB.txt`: verified byte-identical API listings.
