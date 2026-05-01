// Package fixtures centraliza fixtures embebidas reutilizables por tests.
package fixtures

import (
	"embed"
	"io/fs"
	"testing"
)

//go:embed testdata/*
var files embed.FS

// Read devuelve los bytes de un archivo dentro de testdata y falla el test si
// el fixture no existe.
func Read(t testing.TB, name string) []byte {
	t.Helper()

	contents, err := files.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}

	return contents
}

// FS devuelve el filesystem embebido con raíz en testdata para tests que
// necesitan abrir fixtures como archivos.
func FS(t testing.TB) fs.FS {
	t.Helper()

	sub, err := fs.Sub(files, "testdata")
	if err != nil {
		t.Fatalf("open fixture filesystem: %v", err)
	}

	return sub
}
