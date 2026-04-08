package env //nolint:testpackage // exporting to test package necessary functions

import (
	"testing"
)

func ToMapWindows(t *testing.T, env []string) map[string]string {
	t.Helper()

	return toMapWindows(env)
}

func ToMapUnix(t *testing.T, env []string) map[string]string {
	t.Helper()

	return toMapUnix(env)
}
