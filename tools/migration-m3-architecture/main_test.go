package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestM2SourceGuardRejectsFacadeArithmetic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "m2_round.go")
	source := `package dec128
import lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
import ju "github.com/jokruger/dec128/uint128"
var _ = ju.One
var _ = lu.Uint128{Lo: 1}.Add // forbidden local arithmetic at boundary
`
	if err := os.WriteFile(p, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := checkM2File(p); err == nil || !strings.Contains(err.Error(), "local uint128 arithmetic on a literal") {
		t.Fatalf("did not reject selector: %v", err)
	}
}

func TestM2SourceGuardRejectsNewMethodCall(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "m2_candidate.go")
	source := `package dec128
import lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
import ju "github.com/jokruger/dec128/uint128"
type Dec128 struct {coef ju.Uint128}
func New(c lu.Uint128) Dec128 {c.Add(lu.Zero);return Dec128{}}
func newCoefficient(c ju.Uint128) Dec128 {return Dec128{}}
func fromParts(c ju.Uint128) Dec128 {return newCoefficient(c)}
`
	if err := os.WriteFile(p, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := checkM2File(p); err == nil || !strings.Contains(err.Error(), "New calls facade method") {
		t.Fatalf("did not reject boundary call: %v", err)
	}
}

func TestM2SourceGuardRejectsFacadeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "m2_candidate.go")
	source := `package dec128
import lu "github.com/profe-ajedrez/badassitron/dec128/uint128"
import ju "github.com/jokruger/dec128/uint128"
type Dec128 struct {coef ju.Uint128}
func New(c lu.Uint128) Dec128 {return Dec128{}}
func newCoefficient(c ju.Uint128) Dec128 {return Dec128{}}
func fromParts(c ju.Uint128) Dec128 {return New(lu.Uint128(c))}
`
	if err := os.WriteFile(p, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := checkM2File(p); err == nil || !strings.Contains(err.Error(), "fromParts goes back") {
		t.Fatalf("did not reject round trip: %v", err)
	}
}

func TestProductionArchitecture(t *testing.T) {
	if err := checkSource("../.."); err != nil {
		t.Fatal(err)
	}
}
