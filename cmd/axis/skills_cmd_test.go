package main

import "testing"

func TestNewSkillsCommand(t *testing.T) {
	cmd := newSkillsCommand()
	if cmd == nil {
		t.Fatal("newSkillsCommand returned nil")
	}
	if cmd.Use != "skills" {
		t.Errorf("Use = %q, want skills", cmd.Use)
	}
	subs := cmd.Commands()
	names := make(map[string]bool)
	for _, s := range subs {
		names[s.Name()] = true
	}
	for _, want := range []string{"list", "show", "validate", "create"} {
		if !names[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}
