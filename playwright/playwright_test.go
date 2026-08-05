package main

import (
	"os"
	"path/filepath"
	"testing"
)

// stagePythonBinaries creates a fake layer directory under a temp dir and
// touches each of the given usr/bin binaries so resolvePythonBinary has
// something to find.
func stagePythonBinaries(t *testing.T, binaries ...string) string {
	t.Helper()

	base := t.TempDir()
	binDir := filepath.Join(base, "usr", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("failed to create %s: %s", binDir, err)
	}

	for _, binary := range binaries {
		full := filepath.Join(binDir, binary)
		if err := os.WriteFile(full, nil, 0755); err != nil {
			t.Fatalf("failed to stage %s: %s", full, err)
		}
	}

	return base
}

func TestResolvePythonBinary(t *testing.T) {
	tests := []struct {
		name     string
		binaries []string
		// expected is the usr/bin-relative path we expect resolvePythonBinary
		// to return, or "" when we expect an error.
		expected string
	}{
		{
			name:     "prefers python3",
			binaries: []string{"python3"},
			expected: "usr/bin/python3",
		},
		{
			name:     "falls back to python",
			binaries: []string{"python"},
			expected: "usr/bin/python",
		},
		{
			name:     "falls back to a version-suffixed binary",
			binaries: []string{"python3.10"},
			expected: "usr/bin/python3.10",
		},
		{
			name:     "python3 takes priority over a version-suffixed binary",
			binaries: []string{"python3", "python3.10"},
			expected: "usr/bin/python3",
		},
		{
			name:     "python takes priority over a version-suffixed binary",
			binaries: []string{"python", "python3.12"},
			expected: "usr/bin/python",
		},
		{
			name:     "ignores non-interpreter matches",
			binaries: []string{"python3.10-config", "python3.10m"},
			expected: "",
		},
		{
			name:     "errors when nothing is present",
			binaries: []string{},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base := stagePythonBinaries(t, test.binaries...)

			resolved, err := resolvePythonBinary(base)

			if test.expected == "" {
				if err == nil {
					t.Fatalf("expected an error but got %q", resolved)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			want := filepath.Join(base, test.expected)
			if resolved != want {
				t.Fatalf("expected %q but got %q", want, resolved)
			}
		})
	}
}
