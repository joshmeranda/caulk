package main

import (
	"fmt"
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
	pkg.Fset.Iterate(func(file *token.File) bool {
		fmt.Printf("checking file: %v\n", file.Name())
		return true
	})

	return nil, nil
}
