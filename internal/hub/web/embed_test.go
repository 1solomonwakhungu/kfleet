package web

import (
	"io/fs"
	"testing"
)

func TestEmbeddedDist(t *testing.T) {
	entries, err := fs.Glob(dist, "dist/*")
	if err != nil {
		t.Fatalf("fs.Glob(dist, \"dist/*\") error = %v", err)
	}
	_, indexErr := fs.ReadFile(dist, "dist/index.html")

	tests := []struct {
		name string
		want bool
		got  bool
	}{
		// The committed .gitkeep keeps the embed non-empty on a fresh checkout.
		{"embedded dist has entries", true, len(entries) > 0},
		// Empty() is true on a fresh checkout (placeholder dist) and false
		// after make web-build; derive the expectation from the embedded FS.
		{"Empty matches index.html presence", indexErr != nil, Empty()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s: got %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}
