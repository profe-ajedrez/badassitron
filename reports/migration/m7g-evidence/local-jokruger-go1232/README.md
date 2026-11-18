# M7-G supplementary Jokruger source-snapshot experiments

Local exploratory evidence only. The attached `dec128-main(2).zip` contains module
`github.com/jokruger/dec128` with `go 1.24`; the archive itself does not prove that
it matches the **tagged v1.4.0** module required by Badassitron.

- `bench/m7e`, `bench/g2`, `bench/g3`: one three-block, 90-ms-per-subbenchmark
  exploratory A/B campaign for each of the **same three variants** already
  documented in the original M7-G report.
- `bench/m7e-final-local`: **one** final eight-block, 250-ms-per-subbenchmark
  comparison of M7-E (the strongest exploratory local variant), with raw records
  and all samples. It **does not meet** the acceptance gate.
- `validation`: full and race test logs, `vet`, tagged tests and the 5,243
  frozen oracle comparisons. All use a temporary validation-only module override.
- `assembly`: actual Go 1.23.2 disassembly of the **Badassitron adapter**
  (not synthetic stub code) and Go compiler inlining reports. Upstream parser
  source/algorithm is NOT copied into the package.

**Tooling label caveat:** `bench-interleaved.py` historically names the second
candidate `M7-E` in every comparison report and row even when the actual
input is the `g2` or `g3` variant. The directory name and the execution commands
in the technical report identify the actual source used. The raw series have
not been relabeled or fabricated.

No benchmark in this folder used Go 1.27.1 or authenticated Jokruger v1.4.0.
The final technical decision remains **B: keep M7-D unchanged**.
