//go:build !windows

package runtime

func toMap(env []string) map[string]string { return toMapUnix(env) }
