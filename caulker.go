package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log/slog"

	"golang.org/x/tools/go/packages"
)

type Options struct {
	Logger *slog.Logger

	CheckGoroutines bool
	CheckSlices     bool
}

type Result struct{}

type Growth map[string]struct {
	GrownAtLines    []int
	ShrinkedAtLines []int
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
	for _, path := range pkg.GoFiles {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file: %w", err)
		}

		// we don't care about bad declarations for parsing purposes
		_, genDecl, _ := splitDeclarations(file.Decls)
		for _, decl := range genDecl {
			switch decl.Tok {
			case token.TYPE:
				spec, ok := decl.Specs[0].(*ast.TypeSpec)
				if !ok {
					panic(fmt.Sprintf("bug: expected *ast.TypeSpec but found %T", decl.Specs[0]))
				}

				switch t := spec.Type.(type) {
				case *ast.StructType:
					growableFields := findGrowableFields(t)
					fmt.Printf("growable fields: %v\n", growableFields)
				case *ast.Ident:
					panic("non-struct types not yet supported")
				}
			case token.VAR:
				panic("growable vars are not yet supported")
			default:
				continue
			}
		}
	}

	return nil, nil
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
