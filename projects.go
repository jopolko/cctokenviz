package main

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// A folder counts as a project when it holds one of these.
var projectMarkers = []string{".git", "go.mod", "package.json", "pyproject.toml", "Cargo.toml", "pom.xml", "CLAUDE.md", ".claude"}

type ttlEntry struct {
	val string
	at  time.Time
}

var (
	projCacheMu sync.Mutex
	projCache   = map[string]ttlEntry{} // "kind|path" -> answer, expires after projCacheTTL
)

const projCacheTTL = 60 * time.Second

func cached(key string, compute func() string) string {
	projCacheMu.Lock()
	if e, ok := projCache[key]; ok && time.Since(e.at) < projCacheTTL {
		projCacheMu.Unlock()
		return e.val
	}
	projCacheMu.Unlock()
	v := compute()
	projCacheMu.Lock()
	if len(projCache) > 5000 {
		projCache = map[string]ttlEntry{}
	}
	projCache[key] = ttlEntry{v, time.Now()}
	projCacheMu.Unlock()
	return v
}

func isProjectDir(dir string) bool {
	for _, m := range projectMarkers {
		if _, err := os.Stat(filepath.Join(dir, m)); err == nil {
			return true
		}
	}
	return false
}

func projectSubdirs(dir string) []string {
	var out []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") && isProjectDir(filepath.Join(dir, e.Name())) {
			out = append(out, e.Name())
		}
	}
	return out
}

// isContainer reports whether dir is a folder of projects (two or more project folders
// inside) rather than a project itself, e.g. ~/projects.
func isContainer(dir string) bool {
	return cached("c|"+strings.ToLower(dir), func() string {
		if len(projectSubdirs(dir)) >= 2 {
			return "y"
		}
		return ""
	}) == "y"
}

// normalizePath makes an absolute path comparable; git-bash style /c/Users/x becomes C:\Users\x
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	if runtime.GOOS == "windows" && len(p) >= 3 && p[0] == '/' && p[2] == '/' &&
		((p[1] >= 'a' && p[1] <= 'z') || (p[1] >= 'A' && p[1] <= 'Z')) {
		p = strings.ToUpper(p[1:2]) + ":" + p[2:]
	}
	if !filepath.IsAbs(p) && !(len(p) > 1 && p[1] == ':') && !strings.HasPrefix(p, "/") {
		return ""
	}
	return filepath.Clean(filepath.FromSlash(p))
}

// under returns the remainder of p below root (empty if p is root), and whether p is inside it
func under(root, p string) (string, bool) {
	r := strings.ToLower(strings.ReplaceAll(root, "\\", "/"))
	q := strings.ReplaceAll(p, "\\", "/")
	ql := strings.ToLower(q)
	r = strings.TrimRight(r, "/")
	if ql == r {
		return "", true
	}
	if strings.HasPrefix(ql, r+"/") {
		return q[len(r)+1:], true
	}
	return "", false
}

// projectFromPath names the project a file or folder belongs to:
//  1. a folder you added in Settings (a folder of projects, or a single project),
//  2. the first folder under any directory named "projects",
//  3. the nearest enclosing folder that looks like a project (.git, go.mod, ...).
//
// It returns "" when the path is in none of these.
func projectFromPath(path string, roots []string) string {
	p := normalizePath(path)
	if p == "" {
		return ""
	}
	segs := strings.Split(strings.ReplaceAll(p, "\\", "/"), "/")
	for _, s := range segs {
		if strings.EqualFold(s, ".claude") { // Claude's own files, never a project
			return ""
		}
	}

	for _, r := range roots {
		r = normalizePath(r)
		if r == "" {
			continue
		}
		rel, ok := under(r, p)
		if !ok {
			continue
		}
		if !isContainer(r) {
			return filepath.Base(r)
		}
		if rel != "" {
			return strings.SplitN(rel, "/", 2)[0]
		}
		return ""
	}

	for i, s := range segs {
		if strings.EqualFold(s, "projects") && i+1 < len(segs) && segs[i+1] != "" {
			return segs[i+1]
		}
	}

	return cached("p|"+strings.ToLower(p), func() string {
		home, _ := os.UserHomeDir()
		cur := p
		for {
			parent := filepath.Dir(cur)
			if parent == cur || strings.EqualFold(cur, home) {
				return ""
			}
			if fi, err := os.Stat(cur); err == nil && fi.IsDir() && isProjectDir(cur) && !isContainer(cur) {
				return filepath.Base(cur)
			}
			cur = parent
		}
	})
}

// discoverableProjects lists project folders to offer in the picker: the folders you added,
// and any "projects" directory seen in session logs or in your home folder.
func discoverableProjects(roots []string, seenProjectsDirs []string) []string {
	set := map[string]bool{}
	add := func(dir string) {
		dir = normalizePath(dir)
		if dir == "" {
			return
		}
		if isContainer(dir) {
			for _, n := range projectSubdirs(dir) {
				set[n] = true
			}
		} else if isProjectDir(dir) {
			set[filepath.Base(dir)] = true
		}
	}
	for _, r := range roots {
		add(r)
	}
	for _, d := range seenProjectsDirs {
		add(d)
	}
	if home, err := os.UserHomeDir(); err == nil {
		add(filepath.Join(home, "projects"))
	}
	out := make([]string, 0, len(set))
	for n := range set {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
