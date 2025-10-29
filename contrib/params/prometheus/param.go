// Package prometheus provides environment-parsed configuration for exposing
// Prometheus metrics and basic health endpoints.
package prometheus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/quenbyako/core"
)

// Exporter represents a serving metrics endpoint; Serve blocks until
// cancellation performing graceful shutdown.
type Exporter interface {
	Serve(ctx context.Context) error
}

func init() { //nolint:gochecknoinits // there is no other way to register parsers
	core.RegisterEnvParser(parsePromhttpExporter)
}

const (
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultReadTimeout       = 5 * time.Second
	DefaultWriteTimeout      = 10 * time.Second
	DefaultIdleTimeout       = 120 * time.Second

	DefaultServerStopTimeout = 5 * time.Second
)

type promhttpWrapper struct {
	log  *slog.Logger
	addr net.Addr

	conn net.Listener

	srv *http.Server
}

var (
	_ core.EnvParam = (*promhttpWrapper)(nil)
	_ Exporter      = (*promhttpWrapper)(nil)
)

//nolint:ireturn // returns interface on intention.
func parsePromhttpExporter(ctx context.Context, v string) (Exporter, error) {
	uri, err := url.Parse(v)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	if uri.Scheme != "http" {
		return nil, fmt.Errorf("unsupported HTTP scheme %q", uri.Scheme)
	}

	host := uri.Hostname()
	ipAddr := net.ParseIP(host)

	if ipAddr == nil {
		return nil, fmt.Errorf("invalid HTTP host %q", host)
	}

	port := uri.Port()
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

	addr := &net.TCPAddr{IP: ipAddr, Port: portNum, Zone: ""}

	return &promhttpWrapper{
		log:  nil, // will be initialized later
		addr: addr,
		conn: nil, // will be initialized later
		srv:  newHTTPServer(),
	}, nil
}

func (g *promhttpWrapper) Configure(ctx context.Context, data *core.ConfigureData) error {
	g.log = slog.New(data.Logger)
	g.srv.Handler = healthChecks(nil, nil)

	return nil
}

func (g *promhttpWrapper) Acquire(ctx context.Context, data *core.AcquireData) error {
	var err error

	listenConfig := &net.ListenConfig{
		Control:   nil,
		KeepAlive: 0,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable:   false,
			Idle:     DefaultIdleTimeout,
			Interval: 0,
			Count:    0,
		},
	}

	g.conn, err = listenConfig.Listen(ctx, g.addr.Network(), g.addr.String())
	if err != nil {
		// TODO: handle error correctly
		return fmt.Errorf("listening on %q %q: %w", g.addr.Network(), g.addr.String(), err)
	}

	return nil
}

func (g *promhttpWrapper) Serve(ctx context.Context) error {
	if g.conn == nil {
		panic("uninitialized") //nolint:forbidigo // unreachable
	}

	if g.srv.Handler == nil {
		g.srv.Handler = http.NotFoundHandler()
	}

	stopLocker := make(chan struct{})

	var shutdownErr error
	go func(err *error) {
		defer close(stopLocker)

		<-ctx.Done()

		timeoutCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Minute)
		defer cancel()

		*err = g.srv.Shutdown(timeoutCtx)
	}(&shutdownErr)

	g.log.Info(
		"starting metrics server",
		slog.String("addr", g.addr.String()),
	)

	err := g.srv.Serve(g.conn)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serving metrics: %w", err)
	}

	<-stopLocker

	g.log.Info(
		"stopped metrics server",
		slog.String("addr", g.addr.String()),
	)

	return shutdownErr
}

func (g *promhttpWrapper) Shutdown(ctx context.Context, data *core.ShutdownData) error {
	err := g.conn.Close()
	if err != nil {
		// poll.errNetClosing{}
		if err.Error() != "use of closed network connection" {
			return nil
		}

		return fmt.Errorf("closing connection: %w", err)
	}

	return nil
}

func healthChecks(promRegister prometheus.Gatherer, ready func(context.Context) bool) http.Handler {
	router := http.NewServeMux()
	router.Handle("/livez", livez())
	router.Handle("/readyz", readyz(ready))
	router.Handle("/metrics", promhttp.HandlerFor(promRegister, promhttp.HandlerOpts{
		EnableOpenMetrics:                   true,
		EnableOpenMetricsTextCreatedSamples: true,
		ErrorLog:                            nil,
		ErrorHandling:                       0,
		Registry:                            nil,
		DisableCompression:                  false,
		OfferedCompressions:                 nil,
		MaxRequestsInFlight:                 0,
		Timeout:                             DefaultReadHeaderTimeout,
		ProcessStartTime:                    time.Time{},
	}))

	return router
}

func livez() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func readyz(ready func(context.Context) bool) http.Handler {
	if ready == nil {
		ready = func(context.Context) bool { return true }
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ready(r.Context()) {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusServiceUnavailable)
	})
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
