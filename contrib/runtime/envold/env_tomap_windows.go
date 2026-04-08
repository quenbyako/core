//go:build windows

package env

func ToMap(env []string) map[string]string { return toMapWindows(env) }
