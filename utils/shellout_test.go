package utils

import (
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
	if errStr == "" {
		t.Fatal("error message should not be empty")
	}
}

func TestShellOutEcho(t *testing.T) {
	// Just make sure a real command runs without error
	err := ShellOut("echo hello", "echo test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
