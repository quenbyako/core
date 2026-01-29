package runtime

func ToMapWindows(env []string) map[string]string { return toMapWindows(env) }
func ToMapUnix(env []string) map[string]string    { return toMapUnix(env) }
