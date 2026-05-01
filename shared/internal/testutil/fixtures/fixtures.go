package fixtures

import (
	"embed"
	"io/fs"
	"testing"
)

//go:embed testdata/*
var files embed.FS

// Read returns fixture bytes from this package's testdata directory.
func Read(t testing.TB, name string) []byte {
	t.Helper()

	contents, err := files.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}

	return contents
}

// FS returns the embedded fixture filesystem rooted at testdata.
func FS(t testing.TB) fs.FS {
	t.Helper()

	sub, err := fs.Sub(files, "testdata")
	if err != nil {
		t.Fatalf("open fixture filesystem: %v", err)
	}

	return sub
}
