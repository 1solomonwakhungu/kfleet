package version

import "testing"

func TestStringFallsBackWhenUnstamped(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })

	Version = ""
	if got := String(); got != "dev" {
		t.Fatalf("String() = %q, want %q", got, "dev")
	}

	Version = "v1.2.3"
	if got := String(); got != "v1.2.3" {
		t.Fatalf("String() = %q, want stamped version", got)
	}
}
