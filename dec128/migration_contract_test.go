package dec128

import (
 "testing"
 "github.com/profe-ajedrez/badassitron/dec128/errors"
)

// These tests pin observed legacy behavior before the atomic upstream migration.
// Known defects are tracked separately in docs/MIGRATION_DECISIONS.md.
func TestMigrationContractStoredScaleAndFormatting(t *testing.T) {
 d := FromString("12.3400")
 if d.IsNaN() || d.Precision()!=4 || d.StringFixed()!="12.3400" || d.String()!="12.34" {t.Fatalf("unexpected stored scale/format: %#v (%s, %s)",d,d.String(),d.StringFixed())}
 if got:=d.Canonical(); got.Precision()!=2 || got.String()!="12.34" {t.Fatalf("canonical: %s scale %d",got.StringFixed(),got.Precision())}
}
func TestMigrationContractLegacyParsing(t *testing.T) {
 for _,s:=range []string{"", "5."} {d:=FromString(s);if d.IsNaN(){t.Fatalf("historical parser rejects %q: %v",s,d.ErrorDetails())}}
}
func TestMigrationContractMulExactOrNaN(t *testing.T) {
 a:=FromString("0.0000000001")
 b:=FromString("0.0000000001")
 got:=a.Mul(b)
 if !got.IsNaN() || got.ErrorDetails()==nil || got.ErrorDetails().Error()!=errors.Overflow.Error() {t.Fatalf("legacy Mul: got %s (%v)",got.String(),got.ErrorDetails())}
 if got:=FromString("12.50").Mul(FromInt(3));got.IsNaN()||got.StringFixed()!="37.50" {t.Fatalf("exact Mul: %s",got.StringFixed())}
}
func TestMigrationContractRounding(t *testing.T) {
 for _,tc:=range []struct{input string; output string}{{"1.235","1.24"},{"-1.235","-1.24"}} {
  if got:=FromString(tc.input).Round(2).StringFixed();got!=tc.output {t.Fatalf("Round(%s) = %s, want %s",tc.input,got,tc.output)}
 }
 if got:=FromString("2.125").RoundBank(2).StringFixed();got!="2.12" {t.Fatalf("banker's tie = %s",got)}
}
func TestMigrationContractNaNAndZero(t *testing.T) {
 n:=NaN(errors.InvalidFormat)
 if !n.IsNaN()||n.ErrorDetails()==nil||!n.Equal(NaN(errors.Overflow)){t.Fatal("NaN contract changed")}
 z:=NewFromInt(0)
 if !z.Equal(Zero)||z.Precision()!=0{t.Fatal("zero contract changed")}
}
