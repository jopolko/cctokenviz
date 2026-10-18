package main

import (
	"os"
	"path/filepath"
)

// This app used to be called HeadRoom and kept its files in ~/.headroom and ~/.headroom.json.
// On launch those are moved to the new names so settings, order and dismissals carry over.
// If the new folder already exists (a status line script may have created it first), only
// the files it lacks are moved in.
func migrateLegacyData() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	migrateLegacyIn(home)
}

func migrateLegacyIn(home string) {
	oldCfg, newCfg := filepath.Join(home, ".headroom.json"), filepath.Join(home, ".cctokenviz.json")
	if _, err := os.Stat(newCfg); err != nil {
		os.Rename(oldCfg, newCfg)
	}

	oldDir, newDir := filepath.Join(home, ".headroom"), filepath.Join(home, ".cctokenviz")
	entries, err := os.ReadDir(oldDir)
	if err != nil {
		return
	}
	if _, err := os.Stat(newDir); err != nil {
		if os.Rename(oldDir, newDir) == nil {
			return
		}
		os.MkdirAll(newDir, 0o755)
	}
	for _, e := range entries {
		dst := filepath.Join(newDir, e.Name())
		if _, err := os.Stat(dst); err != nil {
			os.Rename(filepath.Join(oldDir, e.Name()), dst)
		}
	}
	os.Remove(oldDir) // only succeeds if now empty
}
