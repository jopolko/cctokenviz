package main

import (
	"os"
	"path/filepath"
	"testing"
)

func assistantLine(cwd string, input, cacheRead int) string {
	return `{"type":"assistant","cwd":"` + cwd + `","message":{"usage":{"input_tokens":` +
		itoa(input) + `,"cache_read_input_tokens":` + itoa(cacheRead) + `,"output_tokens":5}}}` + "\n"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}

func appendFile(t *testing.T, path, text string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		t.Fatal(err)
	}
}

func TestParseSessionIncremental(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "abc.jsonl")
	m := &Monitor{claudeDir: dir}

	appendFile(t, path, assistantLine(`C:\\Users\\me\\projects\\my-app`, 2, 1000))
	info := m.parseSession(path, "proj")
	if info == nil || info.UsedTokens != 1002 || info.ProjectName != "my-app" {
		t.Fatalf("first parse: %+v", info)
	}

	// A half-written line must not be consumed or change the result
	full := assistantLine(`C:\\Users\\me\\projects\\my-app`, 3, 5000)
	appendFile(t, path, full[:len(full)/2])
	if info = m.parseSession(path, "proj"); info.UsedTokens != 1002 {
		t.Fatalf("partial line was consumed: %+v", info)
	}

	// Once the line is completed, it is picked up
	appendFile(t, path, full[len(full)/2:])
	if info = m.parseSession(path, "proj"); info.UsedTokens != 5003 {
		t.Fatalf("completed line not read: %+v", info)
	}

	// A rewritten, shorter file starts over
	os.WriteFile(path, []byte(assistantLine(`C:\\Users\\me\\projects\\other`, 1, 10)), 0o644)
	if info = m.parseSession(path, "proj"); info.UsedTokens != 11 || info.ProjectName != "other" {
		t.Fatalf("shrunk file not reparsed: %+v", info)
	}
}
