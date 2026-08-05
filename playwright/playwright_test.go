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
		name string
		// layerBinaries / systemBinaries are the usr/bin binaries staged in the
		// layer root and the (fake) system root respectively.
		layerBinaries  []string
		systemBinaries []string
		// expected is the usr/bin-relative path we expect resolvePythonBinary to
		// return, or "" when we expect an error.
		expected string
		// expectedInSystem asserts the returned path is under the system root
		// rather than the layer root.
		expectedInSystem bool
	}{
		{
			name:          "prefers python3",
			layerBinaries: []string{"python3"},
			expected:      "usr/bin/python3",
		},
		{
			name:          "falls back to python",
			layerBinaries: []string{"python"},
			expected:      "usr/bin/python",
		},
		{
			name:          "falls back to a version-suffixed binary",
			layerBinaries: []string{"python3.10"},
			expected:      "usr/bin/python3.10",
		},
		{
			name:          "python3 takes priority over a version-suffixed binary",
			layerBinaries: []string{"python3", "python3.10"},
			expected:      "usr/bin/python3",
		},
		{
			name:          "python takes priority over a version-suffixed binary",
			layerBinaries: []string{"python", "python3.12"},
			expected:      "usr/bin/python",
		},
		{
			name:          "ignores non-interpreter matches",
			layerBinaries: []string{"python3.10-config", "python3.10m"},
			expected:      "",
		},
		{
			name:          "errors when nothing is present",
			layerBinaries: []string{},
			expected:      "",
		},
		{
			name:             "falls back to the system python3 when the layer is empty",
			layerBinaries:    []string{},
			systemBinaries:   []string{"python3"},
			expected:         "usr/bin/python3",
			expectedInSystem: true,
		},
		{
			name:             "falls back to the system python when only python is present",
			layerBinaries:    []string{},
			systemBinaries:   []string{"python"},
			expected:         "usr/bin/python",
			expectedInSystem: true,
		},
		{
			name:             "falls back to a version-suffixed system binary",
			layerBinaries:    []string{},
			systemBinaries:   []string{"python3.10"},
			expected:         "usr/bin/python3.10",
			expectedInSystem: true,
		},
		{
			name:           "layer python wins over the system python",
			layerBinaries:  []string{"python3"},
			systemBinaries: []string{"python3"},
			expected:       "usr/bin/python3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base := stagePythonBinaries(t, test.layerBinaries...)

			// Pin the system root to an isolated temp dir so results never
			// depend on the host's real /usr/bin.
			sysRoot := stagePythonBinaries(t, test.systemBinaries...)
			original := systemRoot
			systemRoot = sysRoot
			t.Cleanup(func() { systemRoot = original })

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

			root := base
			if test.expectedInSystem {
				root = sysRoot
			}
			want := filepath.Join(root, test.expected)
			if resolved != want {
				t.Fatalf("expected %q but got %q", want, resolved)
			}
		})
	}
}
