package runtime

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"os"
	"slices"

	"github.com/quenbyako/core"
	"github.com/quenbyako/core/contrib/runtime/env"
	envold "github.com/quenbyako/core/contrib/runtime/envold"
	"github.com/quenbyako/core/contrib/runtime/observability"
	"github.com/quenbyako/core/contrib/secrets"
)

const alternativeLib = false

func Run[T core.ActionConfig](action core.ActionFunc[T]) func(context.Context, []string) core.ExitCode {
	return func(ctx context.Context, _ []string) core.ExitCode {
		var config T

		environ, _ := CtxEnv(ctx)

		var activeParams func() []core.EnvParam

		var err error
		if alternativeLib {
			err = env.Parse(ctx, &config, env.WithEnvironment(environ))
		} else {
			var opt []envold.NewOption
			opt, activeParams = envParams(environ)
			opt = append(opt, envold.WithFuncMap(core.GetParseFunc))

			err = envold.Parse(ctx, &config, opt...)
		}

		// warn: aggregate error is not returned by value, not by pointer
		if e := new(envold.AggregateError); errors.As(err, e) {
			var missedFields []string

			for _, err := range e.Errors {
				if e := new(envold.VarIsNotSetError); errors.As(err, e) {
					missedFields = append(missedFields, e.Key)
				} else {
					panic(fmt.Errorf("%w", err))
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

		otelCfg := config.GetObservabilityConfig()

		logHandler := defaultLogger(os.Stderr, config.GetLogLevel())
		var log LogCallbacks = defaultLogs(logHandler)

		log.EffectiveEnvironment(getEffectiveEnvironment(ctx, &config, environ))

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
		if u := otelCfg.TraceEndpoint; u != nil {
			opts = append(opts, observability.WithOtelAddr(u))
		}
		if u := otelCfg.OtlpMetadata; u != nil {
			opts = append(opts, observability.WithOtlpMetadata(u))
		}
		var metricServer *promhttpWrapper
		if addr := otelCfg.MetricsEndpoint; addr != nil {
			metricServer, err = parsePromhttpExporter(addr)
			if err != nil {
				panic(fmt.Errorf("parsing metrics address %q: %w", addr, err))
			}
			opts = append(opts, observability.WithMetricReader(metricServer.reader))
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
		var configErrs []error
		configurations := activeParams()

		// metrics server has quite specific configuration, so separating it out
		// of other params
		if err := metricServer.configure(ctx, log); err != nil {
			configErrs = append(configErrs, fmt.Errorf("configuring metric server: %w", err))
		}
		for _, v := range configurations {
			if err := v.Configure(ctx, &cfgData); err != nil {
				configErrs = append(configErrs, err)
			}
		}

		if len(configErrs) > 0 {
			for _, err := range configErrs {
				fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
			}
			return 1
		}

		acquireData := core.AcquireData{}

		var acquireErrs []error

		if err := metricServer.acquire(ctx); err != nil {
			acquireErrs = append(acquireErrs, fmt.Errorf("acquiring metric server: %w", err))
		}
		for _, v := range configurations {
			if err := v.Acquire(ctx, &acquireData); err != nil {
				acquireErrs = append(acquireErrs, fmt.Errorf("acquiring %T: %w", v, err))
			}
		}

		if len(acquireErrs) > 0 {
			for _, err := range acquireErrs {
				fmt.Fprintf(os.Stderr, "acquiring resources: %v\n", err)
			}

			return 1
		}

		code := action(ctx, &appCtx[T]{
			IsPipeline: pipes.IsPipeline(),
			stdin:      pipes.Stdin(),
			stdout:     pipes.Stdout(),
			log:        m,
			metric:     m,
			trace:      m,
			config:     config,
			version:    version,
		})

		shutdownData := core.ShutdownData{}

		var shutdownErrs []error

		// Cleanly shut down observability signals (flushing OTLP buffers)
		if s, ok := m.(interface{ Shutdown(context.Context) error }); ok {
			if err := s.Shutdown(ctx); err != nil {
				shutdownErrs = append(shutdownErrs, fmt.Errorf("observability shutdown: %w", err))
			}
		}

		if err := metricServer.shutdown(ctx); err != nil {
			shutdownErrs = append(shutdownErrs, fmt.Errorf("shutting down metric server: %w", err))
		}

		for _, v := range configurations {
			if err := v.Shutdown(ctx, &shutdownData); err != nil {
				shutdownErrs = append(shutdownErrs, err)
			}
		}

		if len(shutdownErrs) > 0 {
			for _, err := range shutdownErrs {
				fmt.Fprintf(os.Stderr, "shutdown error: %v\n", err)
			}

			return 1
		}

		return code
	}
}

func envParams(env map[string]string) ([]envold.NewOption, func() []core.EnvParam) {
	var activeParams []core.EnvParam

	return []envold.NewOption{
		envold.WithTagName("env"),
		envold.WithPrefixTagName("prefix"),
		envold.WithDefaultValueTagName("default"),
		envold.WithRequiredIfNoDef(),
		envold.WithEnvironment(env),
		envold.WithOnSet(func(tag string, value any, isDefault bool) {
			if v, ok := value.(core.EnvParam); ok {
				activeParams = append(activeParams, v)
			}
		}),
	}, func() []core.EnvParam { return activeParams }
}

func getEffectiveEnvironment(ctx context.Context, config any, e map[string]string) map[string]string {
	opts, _ := envParams(nil)
	fields, err := envold.GetFieldParams(ctx, config, opts...)
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
