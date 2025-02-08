package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/tools/txtar"
)

func Test(t *testing.T) {
	ctx := context.TODO() // replace with t.Context

	archive, err := txtar.ParseFile("./testdata/todo.txtar")
	if err != nil {
		t.Fatal(err)
	}

	d := t.TempDir()
	files, err := txtar.FS(archive)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(d, files); err != nil {
		t.Fatal(err)
	}

	sqlc := exec.CommandContext(ctx, "sqlc", "generate")
	sqlc.Dir = d
	sqlc.Stderr = os.Stderr
	sqlc.Stdout = os.Stdout
	if err := sqlc.Run(); err != nil {
		t.Fatal(err)
	}

	if err := generate(ctx, d, "playground", filepath.Join(d, "queries.sql")); err != nil {
		log.Fatal(err)
	}

	goTest := exec.CommandContext(ctx, "go", "test")
	goTest.Dir = d
	goTest.Stderr = os.Stderr
	goTest.Stdout = os.Stdout
	if err := goTest.Run(); err != nil {
		t.Fatal(err)
	}
}
