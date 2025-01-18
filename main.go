package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"slices"

	"golang.org/x/tools/go/packages"

	"github.com/crhntr/generate-smaller-sqlc-interfaces/interfaces"
)

func main() {
	ctx := context.Background()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	var pkgPath string
	flag.StringVar(&pkgPath, "p", ".", "Package path")
	flag.StringVar(&wd, "C", wd, "Working directory")
	flag.Parse()

	filePaths, err := filepaths(wd, flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	if err := generate(ctx, wd, pkgPath, filePaths...); err != nil {
		log.Fatal(err)
	}
}

func generate(ctx context.Context, wd, pkgPath string, filePaths ...string) error {
	outFile := filepath.Join(wd, "querier.go")
	pkg, err := loadPackage(ctx, wd, pkgPath)
	if err != nil {
		return err
	}

	output, err := os.ReadFile(outFile)
	if err != nil {
		log.Fatal(err)
	}

	out := bytes.NewBuffer(output)

	if err := interfaces.Generate(out, pkg, filePaths...); err != nil {
		return err
	}

	output, err = format.Source(out.Bytes())
	if err != nil {
		return err
	}
	if err := os.WriteFile(outFile, output, 0644); err != nil {
		return err
	}

	return nil
}

func loadPackage(ctx context.Context, wd, pkgPath string) (*packages.Package, error) {
	loaded, err := packages.Load(&packages.Config{
		Mode:    packages.LoadSyntax,
		Dir:     wd,
		Context: ctx,
	}, pkgPath)
	if err != nil {
		return nil, err
	}
	if len(loaded) != 1 {
		return nil, fmt.Errorf("expected exactly one package, got %d", len(loaded))
	}
	return loaded[0], nil
}

func filepaths(cd string, args []string) ([]string, error) {
	var filePaths []string
	for _, arg := range args {
		matches, err := filepath.Glob(arg)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			fmt.Println(match)
			if !filepath.IsAbs(match) {
				match, err = filepath.Abs(filepath.Join(cd, match))
				if err != nil {
					return nil, err
				}
			}
			filePaths = append(filePaths, match)
		}
		slices.Sort(filePaths)
		filePaths = slices.Compact(filePaths)
	}
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no files found in %s", cd)
	}
	return filePaths, nil
}
