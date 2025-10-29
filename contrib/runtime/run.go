package runtime

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/quenbyako/core"
	"github.com/quenbyako/core/contrib/runtime/env"
	envold "github.com/quenbyako/core/contrib/runtime/envold"
	"github.com/quenbyako/core/contrib/secrets"
	"github.com/quenbyako/core/internal"
	"github.com/quenbyako/core/observability"
)

const alternativeLib = false

func Run[T core.ActionConfig](action core.ActionFunc[T]) func(context.Context, []string) core.ExitCode {
	return func(ctx context.Context, _ []string) core.ExitCode {
		var config T

		envRaw := os.Environ()
		environ := make(map[string]string, len(envRaw))
		for _, e := range envRaw {
			p := strings.SplitN(e, "=", 2)
			if len(p) == 2 {
				environ[p[0]] = p[1]
			}
		}

		var activeParams func() []core.EnvParam

		var err error
		if alternativeLib {
			err = env.Parse(ctx, &config, env.WithEnvironment(environ))
		} else {
			mappers := make(map[reflect.Type]envold.ParserFunc)
			for typ, f := range internal.GetAllParseFunc() {
				mappers[typ] = func(v string) (any, error) { return f(ctx, v) }
			}

			var opt envold.Options
			opt, activeParams = envParams(environ, mappers)

			err = envold.ParseWithOptions(&config, opt)
		}

		// warn: aggregate error is not returned by value, not by pointer
		if e := new(envold.AggregateError); errors.As(err, e) {
			var missedFields []string

			for _, err := range e.Errors {
				if e := new(envold.VarIsNotSetError); errors.As(err, e) {
					missedFields = append(missedFields, e.Key)
				} else {
					panic(err)
				}
			}

			slices.Sort(missedFields)

			if len(missedFields) > 0 {
				fmt.Fprintf(os.Stderr, "missing required environment variables: %v\n", missedFields)
			} else {
				panic("internal error: env.AggregateError without env.VarIsNotSetError")
			}

			return 1
		} else if err != nil {
			panic(err)
		}

		logHandler := defaultLogger(os.Stderr, config.GetLogLevel())
		var log LogCallbacks = defaultLogs(logHandler)

		log.EffectiveEnvironment(getEffectiveEnvironment(&config, environ))

		var clientCert tls.Certificate
		if certPath, keyPath := config.ClientCertPaths(); certPath != "" && keyPath != "" {
			var err error
			if clientCert, err = tls.LoadX509KeyPair(certPath, keyPath); err != nil {
				panic(fmt.Errorf("loading client certificate: %w", err))
			}
		}

		secretEngine, err := secrets.BuildSecretEngine(ctx, config.GetSecretDSNs())
		if err != nil {
			panic(fmt.Errorf("building secret engine: %w", err))
		}
		caCerts := loadCertificates(config.GetCertPaths())
		version, _ := core.VersionFromContext(ctx)
		pipes, _ := core.PipelinesFromContext(ctx)

		opts := []observability.NewOption{
			observability.WithLogLevel(config.GetLogLevel()),
			observability.WithLogWriter(pipes.Stderr()),
		}
		if u := config.GetTraceEndpoint(); u != nil {
			opts = append(opts, observability.WithOtelAddr(u))
		}

		m, err := observability.New(ctx, opts...)
		if err != nil {
			panic(fmt.Errorf("setting up observability: %w", err))
		}

		cfgData := core.ConfigureData{
			AppCert: clientCert,
			Pool:    caCerts,
			Logger:  logHandler,
			Secrets: secretEngine,
			Version: version,
			Metric:  m,
			Trace:   m,
		}

		// configuring
		var errs []error
		configurations := activeParams()

		for _, v := range configurations {
			if err := v.Configure(ctx, &cfgData); err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			for _, err := range errs {
				fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
			}
			return 1
		}

		acquireData := core.AcquireData{
			ConfigureData: cfgData,
		}

		for _, v := range configurations {
			if err := v.Acquire(ctx, &acquireData); err != nil {
				panic(fmt.Errorf("configuring %T: %w", v, err))
			}
		}

		code := action(ctx, &appCtx[T]{
			IsPipeline: pipes.IsPipeline(),
			stdin:      pipes.Stdin(),
			stdout:     pipes.Stdout(),
			log:        logHandler,
			metric:     m,
			trace:      m,
			config:     config,
			version:    version,
		})

		shutdownData := core.ShutdownData{
			AcquireData: acquireData,
		}

		for _, v := range configurations {
			if err := v.Shutdown(ctx, &shutdownData); err != nil {
				panic(fmt.Errorf("shutting down %T: %w", v, err))
			}
		}

		return code
	}
}

func envParams(e map[string]string, mappers map[reflect.Type]envold.ParserFunc) (envold.Options, func() []core.EnvParam) {
	var activeParams []core.EnvParam

	return envold.Options{
		TagName:             "env",
		PrefixTagName:       "prefix",
		DefaultValueTagName: "default",
		RequiredIfNoDef:     true,
		Environment:         e,
		FuncMap:             mappers,
		OnSet: func(tag string, value any, isDefault bool) {
			if v, ok := value.(core.EnvParam); ok {
				activeParams = append(activeParams, v)
			}
		},
	}, func() []core.EnvParam { return activeParams }
}

func getEffectiveEnvironment(config any, e map[string]string) map[string]string {
	opts, _ := envParams(nil, nil)
	fields, err := envold.GetFieldParamsWithOptions(config, opts)
	if err != nil {
		panic(err)
	}

	params := make(map[string]string)
	for _, field := range fields {
		params[field.Key] = field.DefaultValue
	}

	for k, v := range e {
		if _, ok := params[k]; ok {
			params[k] = v
		}
	}

	return params
}
