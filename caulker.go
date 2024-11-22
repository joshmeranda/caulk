package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"log/slog"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	shrinkSliceFuncNames = []string{
		"slices.Remove",
		"slices.RemoveFunc",
	}
)

type Options struct {
	Logger *slog.Logger

	CheckGoroutines bool
	CheckSlices     bool
}

type Caulker struct {
	Options
}

func NewCaulker(opts Options) *Caulker {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}

	return &Caulker{
		Options: opts,
	}
}

func (c *Caulker) Check(pkg *packages.Package) ([]Result, error) {
	targets := make([]Target, 0)
	shrinks := make([]Shrink, 0)
	results := make([]Result, 0)

	// todo: only supports single file packages
	for _, path := range pkg.GoFiles {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file: %w", err)
		}

		// we don't care about bad declarations for parsing purposes
		_, genDecls, funcDecls := splitDeclarations(file.Decls)

		for _, decl := range genDecls {
			switch decl.Tok {
			case token.TYPE:
				spec, ok := decl.Specs[0].(*ast.TypeSpec)
				if !ok {
					panic(fmt.Sprintf("bug: expected *ast.TypeSpec but found %T", decl.Specs[0]))
				}

				switch t := spec.Type.(type) {
				case *ast.StructType:
					fields := findGrowableFields(t)
					for _, f := range fields {
						targets = append(targets, Target{
							Identity: spec.Name,
							Field:    f,
						})
					}
				case *ast.Ident:
					panic("non-struct types not yet supported")
				}
			case token.VAR:
				panic("growable vars are not yet supported")
			default:
				continue
			}
		}

		for _, decl := range funcDecls {
			if decl.Recv == nil {
				continue
			}

			newUpdates := shrinksFromFunc(decl)
			shrinks = append(shrinks, newUpdates...)
		}

		for _, target := range targets {
			shrunk := slices.ContainsFunc(shrinks, func(u Shrink) bool {
				return target.Equals(u.Target)
			})

			if !shrunk {
				results = append(results, Result{
					Target: target,
					Pos:    target.Position(fset),
				})
			}
		}
	}

	return results, nil
}

func findGrowableFields(t *ast.StructType) []*ast.Field {
	fields := make([]*ast.Field, 0, len(t.Fields.List))

	for _, f := range t.Fields.List {
		switch f.Type.(type) {
		case *ast.ArrayType:
			fields = append(fields, f)
		case *ast.MapType:
			panic("maps are not yet supported")
		}
	}

	return fields
}

// todo: probably should rename Growth to something more generic
func shrinksFromFunc(f *ast.FuncDecl) []Shrink {
	shrinks := make([]Shrink, 0)

	var recv *ast.Field

	if f.Recv != nil {
		recv = f.Recv.List[0]
	}

	for _, stmt := range f.Body.List {
		if shrink, ok := shrinkFromStmt(recv, stmt); ok {
			shrinks = append(shrinks, shrink)
		}
	}

	return shrinks
}

func shrinkFromStmt(recv *ast.Field, stmt ast.Stmt) (Shrink, bool) {
	switch stmt := stmt.(type) {
	case *ast.AssignStmt:
		// todo: currently only supports single assignments
		if len(stmt.Lhs) != 1 {
			return Shrink{}, false
		}

		var target Target

		switch lhs := stmt.Lhs[0].(type) {
		case *ast.SelectorExpr:
			target = Target{
				Identity: recv.Type.(*ast.StarExpr).X.(*ast.IndexExpr).X.(*ast.Ident),
				Field: &ast.Field{
					Names: []*ast.Ident{lhs.Sel},
				},
			}

			switch rhs := stmt.Rhs[0].(type) {
			case *ast.CallExpr:
				funcName := types.ExprString(rhs)
				isShrinkFunc := slices.ContainsFunc(shrinkSliceFuncNames, func(name string) bool {
					return strings.HasPrefix(funcName, name)
				})

				if isShrinkFunc {
					return Shrink{
						Target: target,
						Pos:    stmt.Pos(),
					}, true
				}
			}
		}
	}

	return Shrink{}, false
}

func splitDeclarations(decls []ast.Decl) (bad []*ast.BadDecl, gen []*ast.GenDecl, funcs []*ast.FuncDecl) {
	for _, d := range decls {
		switch d := d.(type) {
		case *ast.GenDecl:
			gen = append(gen, d)
		case *ast.FuncDecl:
			funcs = append(funcs, d)
		case *ast.BadDecl:
			bad = append(bad, d)
		default:
			panic(fmt.Sprintf("bug: unknown declaration type: %T", d))
		}
	}

	return
}
