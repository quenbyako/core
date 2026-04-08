package env_test

import (
	"maps"
	"testing"

	. "github.com/quenbyako/core/contrib/runtime/envold"
)

func TestToMapUnix(t *testing.T) {
	envVars := []string{
		":=/test/unix",
		"PATH=:/test_val1:/test_val2",
		"VAR=REGULARVAR",
		"FOO=",
		"BAR",
	}
	result := ToMapUnix(t, envVars)
	requireEqual(t, result, map[string]string{
		":":    "/test/unix",
		"PATH": ":/test_val1:/test_val2",
		"VAR":  "REGULARVAR",
		"FOO":  "",
	})
}

// On Windows, environment variables can start with '='. This test verifies this
// behavior without relying on a Windows environment. See env_windows.go in the
// Go source:
// https://github.com/golang/go/blob/master/src/syscall/env_windows.go
func TestToMapWindows(t *testing.T) {
	envVars := []string{
		"=::=::\\",
		"=C:=C:\\test",
		"VAR=REGULARVAR",
		"FOO=",
		"BAR",
	}
	result := ToMapWindows(t, envVars)
	requireEqual(t, result, map[string]string{
		"=::": "::\\",
		"=C:": "C:\\test",
		"VAR": "REGULARVAR",
		"FOO": "",
	})
}

func requireEqual(t *testing.T, got, want map[string]string) {
	t.Helper()

	if maps.Equal(got, want) {
		return
	}

	t.Fatalf("Expected %v, got %v", want, got)
}
