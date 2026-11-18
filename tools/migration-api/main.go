// Command migration-api records the exported source-level declarations of a Go tree.
// It uses only the standard library and does not resolve dependencies or run init functions.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	Package   string `json:"package"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Signature string `json:"signature"`
}

func rendered(fset *token.FileSet, n any) string {
	var b strings.Builder
	if err := printer.Fprint(&b, fset, n); err != nil {
		panic(err)
	}
	return b.String()
}
func main() {
	root := flag.String("root", ".", "module root")
	tags := flag.String("tags", "", "comma-separated optional build tags (for example badassitron_m2)")
	flag.Parse()
	ctx := build.Default
	if *tags != "" {
		ctx.BuildTags = strings.Split(*tags, ",")
	}
	dirs := []string{"dec128", "dec128/errors", "dec128/uint128", "dec256", "dec256/uint256", "dec512", "dec512/uint512", "optimus", "withdec128", "withfloat64"}
	var entries []Entry
	for _, dir := range dirs {
		fset := token.NewFileSet()
		packageDir := filepath.Join(*root, dir)
		packages, err := parser.ParseDir(fset, packageDir, func(info os.FileInfo) bool {
			if strings.HasSuffix(info.Name(), "_test.go") {
				return false
			}
			ok, err := ctx.MatchFile(packageDir, info.Name())
			if err != nil {
				panic(err)
			}
			return ok
		}, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, pkg := range packages {
			for _, f := range pkg.Files {
				for _, decl := range f.Decls {
					switch d := decl.(type) {
					case *ast.FuncDecl:
						if !d.Name.IsExported() {
							continue
						}
						name := d.Name.Name
						if d.Recv != nil {
							name = rendered(fset, d.Recv.List[0].Type) + "." + name
						}
						entries = append(entries, Entry{dir, "func", name, rendered(fset, d.Type)})
					case *ast.GenDecl:
						for _, spec := range d.Specs {
							switch s := spec.(type) {
							case *ast.TypeSpec:
								if s.Name.IsExported() {
									entries = append(entries, Entry{dir, "type", s.Name.Name, rendered(fset, s.Type)})
								}
							case *ast.ValueSpec:
								for i, n := range s.Names {
									if n.IsExported() {
										sig := ""
										if s.Type != nil {
											sig = rendered(fset, s.Type)
										}
										if len(s.Values) > i {
											sig += " = " + rendered(fset, s.Values[i])
										}
										entries = append(entries, Entry{dir, d.Tok.String(), n.Name, strings.TrimSpace(sig)})
									}
								}
							}
						}
					}
				}
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		if a.Package != b.Package {
			return a.Package < b.Package
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Name < b.Name
	})
	if err := json.NewEncoder(os.Stdout).Encode(entries); err != nil {
		panic(err)
	}
}
