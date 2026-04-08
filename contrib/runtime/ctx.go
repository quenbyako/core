package runtime

import "context"

type envCtxKey struct{}

func WithEnvContext(ctx context.Context, env map[string]string) context.Context {
	return context.WithValue(ctx, envCtxKey{}, env)
}

func CtxEnv(ctx context.Context) (map[string]string, bool) {
	if env, ok := ctx.Value(envCtxKey{}).(map[string]string); ok {
		return env, true
	}

	return envold.ToMap(os.Environ()), false
}
