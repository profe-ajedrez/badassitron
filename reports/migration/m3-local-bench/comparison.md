# M3 uint128 opt-in benchmark comparison

Descriptive five-or-more-sample medians, not a statistical or release approval.
Environment: {'goos': 'linux', 'goarch': 'amd64', 'cpu': 'AMD EPYC 9V74 80-Core Processor                '}

| Operation | Historical ns/op | M3 ns/op | Delta | B/op historical → M3 | Review |
|---|---:|---:|---:|---:|---|
| `Add` | 0.906 | 1.724 | +90.3% | 0 → 0 | REVIEW |
| `MulExact` | 1.851 | 1.921 | +3.8% | 0 → 0 | within threshold |
| `MulOverflow` | 2.515 | 3.172 | +26.1% | 0 → 0 | REVIEW |
| `QuoRem` | 11.830 | 10.960 | -7.4% | 0 → 0 | within threshold |
| `WideQuoRem` | 11.490 | 12.110 | +5.4% | 0 → 0 | within threshold |

Review cases: Add, MulOverflow

Do not promote M3 if meaningful regressions remain unexplained.
Re-run using Go 1.27.1 and the exact published Jokruger v1.4.0 module.
