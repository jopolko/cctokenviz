package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateLegacyMergesIntoExistingFolder(t *testing.T) {
	home := t.TempDir()
	os.MkdirAll(filepath.Join(home, ".headroom"), 0o755)
	os.MkdirAll(filepath.Join(home, ".cctokenviz"), 0o755)
	os.WriteFile(filepath.Join(home, ".headroom.json"), []byte("{}"), 0o644)
	os.WriteFile(filepath.Join(home, ".headroom", "order.json"), []byte("old"), 0o644)
	os.WriteFile(filepath.Join(home, ".headroom", "status-a.json"), []byte("old"), 0o644)
	os.WriteFile(filepath.Join(home, ".cctokenviz", "status-a.json"), []byte("new"), 0o644)

	migrateLegacyIn(home)

	for _, f := range []string{".cctokenviz.json", filepath.Join(".cctokenviz", "order.json")} {
		if _, err := os.Stat(filepath.Join(home, f)); err != nil {
			t.Errorf("%s not migrated", f)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(home, ".cctokenviz", "status-a.json")); string(b) != "new" {
		t.Errorf("existing file was overwritten: %s", b)
	}
}
