package runtime

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const (
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 5 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 120 * time.Second
)

type promhttpWrapper struct {
	log  LogCallbacks
	addr net.Addr

	reader sdkmetric.Reader
	conn   net.Listener

	srv              *http.Server
	finishServerChan <-chan struct{}
}

func parsePromhttpExporter(uri *url.URL) (*promhttpWrapper, error) {
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

	promreg := prometheus.NewRegistry()
	prometheusExporter, err := otelprometheus.New(
		otelprometheus.WithRegisterer(promreg),
	)
	if err != nil {
		panic(err)
	}

	return &promhttpWrapper{
		log:    nil, // will be initialized later
		addr:   addr,
		reader: prometheusExporter,
		conn:   nil, // will be initialized later
		srv: &http.Server{ //nolint:exhaustruct // server has a lot of fields
			// handler is 404 by default.
			Handler:           healthChecks(promreg, nil),
			ReadTimeout:       defaultReadTimeout,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
		},
	}, nil
}

func (g *promhttpWrapper) configure(ctx context.Context, log LogCallbacks) error {
	g.log = log

	return nil
}

func (g *promhttpWrapper) acquire(ctx context.Context) (err error) {
	listenConfig := &net.ListenConfig{
		Control:   nil,
		KeepAlive: 0,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable:   false,
			Idle:     defaultIdleTimeout,
			Interval: 0,
			Count:    0,
		},
	}

	g.conn, err = listenConfig.Listen(ctx, g.addr.Network(), g.addr.String())
	if err != nil {
		// TODO: handle error correctly
		return fmt.Errorf("listening on %q %q: %w", g.addr.Network(), g.addr.String(), err)
	}

	// calling metrics log here, cause address is already opened, and listener
	// will wait in any case until http server will start handling requests.
	g.log.MetricsStarted(g.addr)

	serverFinished := make(chan struct{})
	g.finishServerChan = serverFinished

	go func() {
		err := g.srv.Serve(g.conn)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			// TODO: it's VEEEERY very dangerous to panic here: it MAY return
			// non context error on the runtime while application is running,
			// which is, extremely bad.
			panic(fmt.Errorf("serving metrics: %w", err))
		}

		g.log.MetricsStopped(g.addr)

		close(serverFinished)
	}()

	return nil
}

func (g *promhttpWrapper) shutdown(ctx context.Context) (err error) {
	timeoutCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Minute)
	defer cancel()

	if err = g.srv.Shutdown(timeoutCtx); err != nil {
		return fmt.Errorf("shutting down metrics server: %w", err)
	}

	err = g.conn.Close()
	if err != nil {
		// poll.errNetClosing{}
		if err.Error() != "use of closed network connection" {
			return nil
		}

		return fmt.Errorf("closing connection: %w", err)
	}

	<-g.finishServerChan

	return nil
}

func healthChecks(promRegister prometheus.Gatherer, ready func(context.Context) bool) http.Handler {
	router := http.NewServeMux()
	router.Handle("/healthz", healthz())
	// TODO
	router.Handle("/readyz", readyz(ready))
	router.Handle("/startupz", readyz(ready))
	router.Handle("/metrics", promhttp.HandlerFor(promRegister, promhttp.HandlerOpts{
		EnableOpenMetrics:                   true,
		EnableOpenMetricsTextCreatedSamples: true,
		ErrorLog:                            nil,
		ErrorHandling:                       0,
		Registry:                            nil,
		DisableCompression:                  false,
		OfferedCompressions:                 nil,
		MaxRequestsInFlight:                 0,
		Timeout:                             defaultReadHeaderTimeout,
		ProcessStartTime:                    time.Time{},
	}))

	return router
}

func healthz() http.Handler {
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
