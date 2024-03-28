package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TODO: Test that the process is terminated after the pattern is matched
func TestMatchesPatternInStdOut(t *testing.T) {
	test_command := "echo pattern matched"
	pattern := "pattern matched"

	cmd := exec.Command("./exec_until", "-p", pattern, test_command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("exec_until command failed: %v, %s", err, output)
	}

	expected_output := "pattern matched"

	if !strings.Contains(string(output), expected_output) {
		t.Errorf("Unexpected output: got %s, want %s", string(output), expected_output)
	}
}

func TestMatchesPatternInStdErr(t *testing.T) {
	expected_message := "message to stderr"
	cmd := exec.Command("./exec_until", "-p", expected_message, fmt.Sprintf("echo \"%s\" >&2", expected_message))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Error("should not throw error", err, string(out))
	}
}

func TestErrorTimeout(t *testing.T) {
	test_command := "sleep 0.1"     // Command will not produce expected pattern within timeout
	timeout := 1 * time.Millisecond // Timeout shorter than time needed for pattern matching

	cmd := exec.Command("./exec_until", "-p", "whatever", "-t", timeout.String(), test_command)
	message, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("should have exited with an exit code")
	}

	if !strings.Contains(string(message), ERROR_TIMEOUT) {
		t.Error("should have errored due to timeout")
	}
}

func TestErrorPatternNotFound(t *testing.T) {
	cmd := exec.Command("./exec_until", "-p", "whatever", "echo asd")
	message, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("should have exited with an exit code")
	}

	if !strings.Contains(string(message), ERROR_PATTERN_NOT_FOUND) {
		t.Error("should have exited due to completing without a match")
	}
}
