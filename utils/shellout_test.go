package utils

import (
	"strings"
	"testing"
)

func TestShellOutSuccess(t *testing.T) {
	err := ShellOut("true", "test true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShellOutFailure(t *testing.T) {
	err := ShellOut("false", "test false")
	if err == nil {
		t.Fatal("expected error for failing command")
	}
}

func TestShellOutEmptyCommand(t *testing.T) {
	err := ShellOut("", "empty")
	if err != nil {
		t.Fatalf("empty command should not error: %v", err)
	}
}

func TestShellOutErrorContainsInfo(t *testing.T) {
	err := ShellOut("false", "my-task")
	if err == nil {
		t.Fatal("expected error")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "my-task") {
		t.Fatalf("error should mention task name 'my-task', got: %s", errStr)
	}
	if !strings.Contains(errStr, "false") {
		t.Fatalf("error should mention command 'false', got: %s", errStr)
	}
}

func TestShellOutEcho(t *testing.T) {
	// Just make sure a real command runs without error
	err := ShellOut("echo hello", "echo test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
