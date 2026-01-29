package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/quenbyako/core/contrib/runtime"
)

func TestToMapUnix(t *testing.T) {
	envVars := []string{":=/test/unix", "PATH=:/test_val1:/test_val2", "VAR=REGULARVAR", "FOO=", "BAR"}
	result := ToMapUnix(envVars)
	require.Equal(t, map[string]string{
		":":    "/test/unix",
		"PATH": ":/test_val1:/test_val2",
		"VAR":  "REGULARVAR",
		"FOO":  "",
	}, result)
}

// On Windows, environment variables can start with '='.
// This test verifies this behavior without relying on a Windows environment.
// See env_windows.go in the Go source: https://github.com/golang/go/blob/master/src/syscall/env_windows.go
func TestToMapWindows(t *testing.T) {
	envVars := []string{"=::=::\\", "=C:=C:\\test", "VAR=REGULARVAR", "FOO=", "BAR"}
	result := ToMapWindows(envVars)
	require.Equal(t, map[string]string{
		"=::": "::\\",
		"=C:": "C:\\test",
		"VAR": "REGULARVAR",
		"FOO": "",
	}, result)
}
