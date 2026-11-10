package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestModulePathMatchesRepository guards the module path.
//
// The module path must match the repository the project is published from.
// The GoLand integration rewrites go.mod to follow the VCS remote, so a
// mismatch between the two is a silent drift: the build would still work,
// but the import paths would not resolve for anyone else.
func TestModulePathMatchesRepository(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			got := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			const want = "github.com/incurl/sumate"
			if got != want {
				t.Errorf("module path is %q, want %q", got, want)
			}
			return
		}
	}
	t.Fatal("no module directive in go.mod")
}
