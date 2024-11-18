package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

type stringList []string

func (s *stringList) String() string {
	return fmt.Sprint(*s)
}

func (s *stringList) Set(value string) error {
	*s = strings.Split(value, " ")
	return nil
}

var (
	disable stringList

	verbose bool
	silent  bool
)

const (
	CheckNameSlices     = "slices"
	CheckNameGoroutines = "goroutines"
)

func init() {
	flag.Var(&disable, "disable", "disable a specific check")

	flag.BoolVar(&verbose, "verbose", false, "enable verbose logging")
	flag.BoolVar(&silent, "silent", false, "disable all logging except for errors")
}

func loadPackages() ([]*packages.Package, error) {
	path := flag.Arg(0)

	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedName,
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			pkgErrors := make([]string, len(pkg.Errors))
			for i, err := range pkg.Errors {
				pkgErrors[i] = err.Error()
			}

			return nil, fmt.Errorf("failed to load package: %s", strings.Join(pkgErrors, ", "))
		}
	}

	return pkgs, nil
}

func run() error {
	var logLevel slog.Level
	switch {
	case silent:
		logLevel = slog.LevelError
	case verbose:
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})).WithGroup("caulker")

	caulker := NewCaulker(Options{
		Logger: logger,

		CheckGoroutines: !slices.Contains(disable, CheckNameGoroutines),
		CheckSlices:     !slices.Contains(disable, CheckNameSlices),
	})

	pkgs, err := loadPackages()
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}

	for _, pkg := range pkgs {
		results, err := caulker.Check(pkg)
		if err != nil {
			return fmt.Errorf("encountered error while checking package: %w", err)
		}

		for _, result := range results {
			fmt.Printf("%v\n", result)
		}
	}

	return nil
}

func main() {
	flag.Parse()

	if flag.NArg() > 1 {
		fmt.Printf("Error: unexpected arguments: %v\n", flag.Args())
		flag.Usage()
		return
	}

	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
