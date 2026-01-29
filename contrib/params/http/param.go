package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/quenbyako/core"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Server abstracts HTTP service registration and serving lifecycle. Register
// installs a root handler; Serve blocks until context cancellation initiating
// graceful shutdown.
type Server interface {
	Register(http.Handler)

	Serve(ctx context.Context) error
}

func init() {
	core.RegisterEnvParser(parseHTTPServer)
}

const (
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultReadTimeout       = 5 * time.Second
	DefaultWriteTimeout      = 10 * time.Second
	DefaultIdleTimeout       = 120 * time.Second

	DefaultServerStopTimeout = 5 * time.Second
)

type httpServerWrapper struct {
	log        *slog.Logger
	addr       net.Addr
	pathPrefix string

	conn net.Listener

	meterProvider metric.MeterProvider
	traceProvider trace.TracerProvider

	srv     *http.Server
	handler http.Handler
}

var _ core.EnvParam = (*httpServerWrapper)(nil)
var _ Server = (*httpServerWrapper)(nil)

func parseHTTPServer(ctx context.Context, v string) (Server, error) {
	u, err := url.Parse(v)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "http" {
		return nil, fmt.Errorf("unsupported HTTP scheme %q", u.Scheme)
	}

	host := u.Hostname()
	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("invalid HTTP host %q", host)
	}
	port := u.Port()
	if port == "" {
		return nil, fmt.Errorf("invalid HTTP port %q", port)
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP port %q", port)
	}
	if portNum < 0 || portNum > 65535 {
		return nil, fmt.Errorf("out of range HTTP port %q", port)
	}
	addr := &net.TCPAddr{IP: ip, Port: portNum}

	pathPrefix := u.Path
	if !strings.HasSuffix(pathPrefix, "/") {
		// setting prefix last slash to make router correct for handling
		// subpaths.
		pathPrefix += "/"
	}

	return &httpServerWrapper{
		addr:       addr,
		srv:        newHTTPServer(),
		pathPrefix: pathPrefix,
	}, nil
}

func (g *httpServerWrapper) Configure(ctx context.Context, data *core.ConfigureData) error {
	g.log = slog.New(data.Logger)
	g.meterProvider = data.Metric
	g.traceProvider = data.Trace

	return nil
}

func (g *httpServerWrapper) Acquire(ctx context.Context, data *core.AcquireData) error {
	var err error
	g.conn, err = net.Listen(g.addr.Network(), g.addr.String()) // TODO: handle error correctly
	if err != nil {
		return fmt.Errorf("listening on %q %q: %w", g.addr.Network(), g.addr.String(), err)
	}

	return nil
}

func (h *httpServerWrapper) Register(handler http.Handler) {
	// NOTE(rcooper): makes no sense to make this thread-safe, because
	// initialization usually performs in one goroutine.
	if h.handler != nil {
		panic("already registered")
	}

	h.handler = handler
}

func (h *httpServerWrapper) Serve(ctx context.Context) error {
	if h.conn == nil {
		panic("connection is not acquired")
	}

	if h.handler == nil {
		h.handler = http.HandlerFunc(http.NotFound)
	}

	mux := http.NewServeMux()
	mux.Handle(h.pathPrefix, chainMiddlewares(
		mwStripPrefix(strings.TrimSuffix(h.pathPrefix, "/")),
		mwOTelHTTP(h.traceProvider, h.meterProvider),
	)(h.handler))
	h.srv.Handler = mux

	stopLocker := make(chan struct{})
	var shutdownErr error
	go func(err *error) {
		defer close(stopLocker)
		<-ctx.Done()

		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		*err = h.srv.Shutdown(timeoutCtx)
	}(&shutdownErr)

	h.log.Info(
		"starting HTTP server",
		slog.String("addr", h.addr.String()),
	)

	err := h.srv.Serve(h.conn)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serving server: %w", err)
	}

	<-stopLocker

	h.log.Info(
		"stopped HTTP server",
		slog.String("addr", h.addr.String()),
	)

	return shutdownErr
}

func (h *httpServerWrapper) Shutdown(ctx context.Context, data *core.ShutdownData) error {
	err := h.conn.Close()
	if err != nil {
		// poll.errNetClosing{}
		if err.Error() != "use of closed network connection" {
			return nil
		}

		return fmt.Errorf("closing connection: %w", err)
	}

	return nil
}

func newHTTPServer() *http.Server {
	return &http.Server{ //nolint:exhaustruct // server has a lot of fields
		// handler is 404 by default.
		Handler:           nil,
		ReadTimeout:       DefaultReadTimeout,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
		WriteTimeout:      DefaultWriteTimeout,
		IdleTimeout:       DefaultIdleTimeout,
	}
}

type middleware = func(http.Handler) http.Handler

func chainMiddlewares(middlewares ...middleware) middleware {
	return func(handler http.Handler) http.Handler {
		// Start with the final handler
		chainedHandler := handler

		// Apply middlewares in reverse order
		for i := len(middlewares) - 1; i >= 0; i-- {
			chainedHandler = middlewares[i](chainedHandler)
		}

		return chainedHandler
	}
}

func mwStripPrefix(path string) middleware {
	return func(next http.Handler) http.Handler {
		return http.StripPrefix(path, next)
	}
}

func mwOTelHTTP(t trace.TracerProvider, m metric.MeterProvider) middleware {
	spanFormatter := func(operation string, r *http.Request) string {
		return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	}

	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "",
			otelhttp.WithSpanNameFormatter(spanFormatter),
			otelhttp.WithTracerProvider(t),
			otelhttp.WithMeterProvider(m),
		)
	}
}
