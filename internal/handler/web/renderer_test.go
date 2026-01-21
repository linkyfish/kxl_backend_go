package web

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplatesCompile(t *testing.T) {
	templateDir := filepath.Join("..", "..", "..", "templates")
	r, err := NewRenderer(templateDir)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	err = filepath.WalkDir(templateDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}
		rel, err := filepath.Rel(templateDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		t.Logf("compile %s", rel)
		if _, err := r.get(rel); err != nil {
			t.Errorf("compile %s: %v", rel, err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk templates: %v", err)
	}
}
