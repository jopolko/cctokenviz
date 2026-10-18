package main

import (
	"encoding/json"
	"testing"
)

func TestProjectFromToolUse(t *testing.T) {
	cases := map[string]string{
		`[{"type":"tool_use","input":{"file_path":"C:\\Users\\josh\\projects\\lookout\\a.py"}}]`:                   "lookout",
		`[{"type":"tool_use","input":{"command":"cd /c/Users/josh/projects/kidsevents && ls"}}]`:                   "kidsevents",
		`[{"type":"tool_use","input":{"file_path":"C:\Users\josh\.claude\projects\C--Users-josh-x\memory\a.md"}}]`: "",
		`[{"type":"text","text":"projects/lookout"}]`:                                                              "",
		`"plain string"`: "",
	}
	for in, want := range cases {
		if got := projectFromToolUse(json.RawMessage(in), nil); got != want {
			t.Errorf("%s: got %q want %q", in, got, want)
		}
	}
}
