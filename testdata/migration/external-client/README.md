# Independent consumer fixture for the M2 cutover

`client_test.go` imports **only exported package paths**, asserting interface
method sets, generic `FromString`, stored scale and signed zero, error codes,
text/JSON/SQL behavior, rounding-mode interoperability, and wide arithmetic.

The fixture can be run within this repository as
`go test ./testdata/migration/external-client`, and also as a separate module
(which is a stronger source-compatibility check):

```sh
cp -R testdata/migration/external-client /tmp/badassitron-client
cd /tmp/badassitron-client
cat > go.mod <<'MOD'
module example.net/badassitron-client-fixture

go 1.27

require github.com/profe-ajedrez/badassitron v0.0.0

replace github.com/profe-ajedrez/badassitron => /absolute/path/to/badassitron
MOD
GOWORK=off go test -mod=mod -tags badassitron_m2 ./... -count=1
```

Use an **absolute path** in the temporary `replace`, or a released version.
Do not commit this replacement into the Badassitron module. For the official
release gate, use Go 1.27 and resolve Jokruger v1.4.0 from Go Modules, not a
local upstream snapshot. A temporary 1.23-compatible root/module override
may be used offline for *local-only* testing, and must not be shipped.
