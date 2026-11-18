// Command migration-m3-architecture checks that the public uint128 compatibility
// type is not silently reintroduced as the M2 decimal arithmetic backend.
// It is a SOURCE guard, not a substitute for a type-checked call graph,
// differential tests or a performance measurement.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	localUint    = "github.com/profe-ajedrez/badassitron/dec128/uint128"
	upstreamUint = "github.com/jokruger/dec128/uint128"
)

// These are representations, public declarations and data tables, not
// arithmetic entry points. Changes here require explicit review.
var allowedFacadeSelectors = map[string]bool{
	"Uint128": true, "Zero": true, "MaxStrLen": true,
	"Pow10Uint64": true, "Pow10Uint128": true,
}

func parse(path string) (*ast.File, *token.FileSet, error) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, path, nil, parser.ParseComments)
	return file, fs, err
}

func importAlias(file *ast.File, importPath string) string {
	for _, imp := range file.Imports {
		name, _ := strconv.Unquote(imp.Path.Value)
		if name != importPath {
			continue
		}
		if imp.Name != nil {
			return imp.Name.Name
		}
		return filepath.Base(name)
	}
	return ""
}

func selectorType(x ast.Expr, alias, typ string) bool {
	s, ok := x.(*ast.SelectorExpr)
	if !ok || s.Sel.Name != typ {
		return false
	}
	i, ok := s.X.(*ast.Ident)
	return ok && i.Name == alias
}

func checkM2File(path string) error {
	f, fs, err := parse(path)
	if err != nil {
		return err
	}
	up := importAlias(f, upstreamUint)
	lu := importAlias(f, localUint)
	if strings.Contains(filepath.Base(path), "candidate") || strings.Contains(filepath.Base(path), "arithmetic") || strings.Contains(filepath.Base(path), "round") || strings.Contains(filepath.Base(path), "parse") || strings.Contains(filepath.Base(path), "remainder") {
		if up == "" {
			return fmt.Errorf("%s: missing imported upstream coefficient engine", path)
		}
	}
	var bad error
	ast.Inspect(f, func(n ast.Node) bool {
		if bad != nil {
			return false
		}
		s, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		i, ok := s.X.(*ast.Ident)
		if ok && lu != "" && i.Name == lu && !allowedFacadeSelectors[s.Sel.Name] {
			bad = fmt.Errorf("%s: unexpected public uint128 selector %s.%s", fs.Position(s.Pos()), lu, s.Sel.Name)
		}
		// Reject direct method calls on facade literals or conversions. This
		// is not a general type-inference pass, so the New boundary is checked
		// separately below.
		if lu != "" {
			switch r := s.X.(type) {
			case *ast.CompositeLit:
				if selectorType(r.Type, lu, "Uint128") {
					bad = fmt.Errorf("%s: local uint128 arithmetic on a literal", fs.Position(s.Pos()))
				}
			case *ast.CallExpr:
				if selectorType(r.Fun, lu, "Uint128") {
					bad = fmt.Errorf("%s: local uint128 arithmetic after conversion", fs.Position(s.Pos()))
				}
			}
		}
		return true
	})
	if bad != nil {
		return bad
	}
	// The public coefficient type is only allowed at New and Coefficient
	// boundaries. In particular, the internal fromParts path must never
	// round-trip through the public facade or call its arithmetic methods.
	if filepath.Base(path) == "m2_candidate.go" {
		foundNew, foundFrom, foundCoef, foundBridge := false, false, false, false
		for _, d := range f.Decls {
			switch v := d.(type) {
			case *ast.GenDecl:
				for _, spec := range v.Specs {
					t, ok := spec.(*ast.TypeSpec)
					if !ok || t.Name.Name != "Dec128" {
						continue
					}
					st, ok := t.Type.(*ast.StructType)
					if !ok {
						return fmt.Errorf("Dec128 is no longer a struct")
					}
					for _, field := range st.Fields.List {
						if len(field.Names) > 0 && field.Names[0].Name == "coef" && selectorType(field.Type, up, "Uint128") {
							foundCoef = true
						}
					}
				}
			case *ast.FuncDecl:
				if v.Name.Name == "New" {
					if len(v.Type.Params.List) > 0 && selectorType(v.Type.Params.List[0].Type, lu, "Uint128") {
						foundNew = true
					}
					// No methods on the public uint128 parameter: conversion only.
					ast.Inspect(v.Body, func(n ast.Node) bool {
						call, ok := n.(*ast.CallExpr)
						if !ok {
							return true
						}
						selector, ok := call.Fun.(*ast.SelectorExpr)
						if !ok {
							return true
						}
						id, ok := selector.X.(*ast.Ident)
						if ok && id.Name == "c" {
							bad = fmt.Errorf("New calls facade method c.%s", selector.Sel.Name)
						}
						return true
					})
				} else if v.Name.Name == "fromParts" {
					if len(v.Type.Params.List) > 0 && selectorType(v.Type.Params.List[0].Type, up, "Uint128") {
						foundFrom = true
					}
					ast.Inspect(v.Body, func(n ast.Node) bool {
						call, ok := n.(*ast.CallExpr)
						if !ok {
							return true
						}
						if name, ok := call.Fun.(*ast.Ident); ok && name.Name == "New" {
							bad = fmt.Errorf("fromParts goes back through public New")
						}
						return true
					})
				} else if v.Name.Name == "newCoefficient" {
					if len(v.Type.Params.List) > 0 && selectorType(v.Type.Params.List[0].Type, up, "Uint128") {
						foundBridge = true
					}
				}
			}
		}
		if bad != nil {
			return bad
		}
		if !foundNew || !foundFrom || !foundCoef || !foundBridge {
			return fmt.Errorf("M2 API boundary invariants missing: New=%v fromParts=%v coef=%v newCoefficient=%v", foundNew, foundFrom, foundCoef, foundBridge)
		}
	}
	return nil
}

func check(root string) error {
	goMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return err
	}
	if !strings.Contains(string(goMod), "github.com/jokruger/dec128 v1.4.0") || strings.Contains(string(goMod), "\nreplace ") {
		return fmt.Errorf("module must pin published Jokruger v1.4.0 without replacement")
	}
	return checkSource(root)
}

// checkSource can be exercised in a throwaway offline test copy that has a
// local Go-module replacement; the final delivered tree is checked by check.
func checkSource(root string) error {
	files, err := filepath.Glob(filepath.Join(root, "dec128", "m2_*.go"))
	if err != nil {
		return err
	}
	if len(files) != 6 {
		return fmt.Errorf("expected six M2 implementation files, found %d", len(files))
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(string(data), "//go:build") {
			return fmt.Errorf("%s: production M2 source must not be build-tagged", path)
		}
		if err := checkM2File(path); err != nil {
			return err
		}
	}
	// A public type alias to Jokruger's Uint128 would change the method set
	// and errors.Error return signatures, even if its limbs match by layout.
	publicType, _, err := parse(filepath.Join(root, "dec128/uint128/uint128.go"))
	if err != nil {
		return err
	}
	foundStableType := false
	for _, decl := range publicType.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, item := range gd.Specs {
			ts, ok := item.(*ast.TypeSpec)
			if !ok || ts.Name.Name != "Uint128" {
				continue
			}
			if ts.Assign.IsValid() {
				return fmt.Errorf("public Uint128 must remain a nominal type, not an alias")
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok || len(st.Fields.List) != 2 {
				return fmt.Errorf("public Uint128 must retain its two-limb struct")
			}
			for i, want := range []string{"Lo", "Hi"} {
				fld := st.Fields.List[i]
				typ, ok := fld.Type.(*ast.Ident)
				if len(fld.Names) != 1 || fld.Names[0].Name != want || !ok || typ.Name != "uint64" {
					return fmt.Errorf("public Uint128 field %d must remain %s uint64", i, want)
				}
			}
			foundStableType = true
		}
	}
	if !foundStableType {
		return fmt.Errorf("missing public Uint128 struct")
	}

	// Exactly one imported unsigned implementation is compiled; no historical
	// arithmetic sources are retained in the release tree (even under tags).
	for _, filename := range []string{"m3_adapter.go", "from_upstream_string.go", "to_upstream_string.go"} {
		path := filepath.Join(root, "dec128/uint128", filename)
		data, err := os.ReadFile(path)
		if err != nil { return err }
		if strings.HasPrefix(string(data), "//go:build") {
			return fmt.Errorf("%s: imported unsigned source must not be build-tagged", path)
		}
	}
	for _, filename := range []string{"arithmetic.go", "bitwise.go", "tools.go", "from_legacy_string.go", "to_legacy_string.go"} {
		if _, err := os.Stat(filepath.Join(root, "dec128/uint128", filename)); err == nil {
			return fmt.Errorf("historical uint128 implementation still shipped: %s", filename)
		} else if !os.IsNotExist(err) { return err }
	}
	for _, filename := range []string{"arithmetic.go", "const.go", "dec128.go", "from.go", "internal.go", "rescale_round.go", "round_ops.go", "rounding.go", "sql_scan.go", "sqrt_round.go", "state.go", "tools.go"} {
		if _, err := os.Stat(filepath.Join(root, "dec128", filename)); err == nil {
			return fmt.Errorf("historical dec128 implementation still shipped: %s", filename)
		} else if !os.IsNotExist(err) { return err }
	}
	for _, dir := range []string{"dec256", "dec512", "optimus"} {
		err := filepath.Walk(filepath.Join(root, dir), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			f, _, err := parse(path)
			if err != nil {
				return err
			}
			if importAlias(f, upstreamUint) != "" {
				return fmt.Errorf("%s: unexpected direct migration of wide/optimus consumer", path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	if err := check(root); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("PASS: M2 coefficients and default public uint128 arithmetic use imported Jokruger; compatibility types remain stable")
	fmt.Println("PASS: historic arithmetical implementations absent; wide and optimus consumers retain compatible public types")
}
