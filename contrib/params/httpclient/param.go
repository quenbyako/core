package httpclient

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/quenbyako/core"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Client abstracts HTTP client operations, providing methods to perform HTTP requests,
// access the base URL, and retrieve the underlying native http.Client.
//
type Client interface {
	Do(req *http.Request) (*http.Response, error)

	BaseURL() *url.URL
	Native() *http.Client
}

func init() {
	core.RegisterEnvParser(parseHTTPClient)
}

const (
	DefaultReadHeaderTimeout = 5 * time.Second
	DefaultReadTimeout       = 5 * time.Second
	DefaultWriteTimeout      = 10 * time.Second
	DefaultIdleTimeout       = 120 * time.Second

	DefaultRequestTimeout = 10 * time.Second
)

type clientWrapper struct {
	baseURL      *url.URL
	pingEndpoint bool // whether to ping endpoint on Configure, or do it lazily

	client *http.Client

	log *slog.Logger
}

var _ core.EnvParam = (*clientWrapper)(nil)
var _ Client = (*clientWrapper)(nil)

func parseHTTPClient(ctx context.Context, v string) (Client, error) {
	const autoTLSScheme = "http+auto"

	v = setDefaultScheme(v, autoTLSScheme)

	u, err := url.Parse(v)
	if err != nil {
		return nil, err
	}

	// TODO: support AUTO scheme, when client doesn't know about TLS, but it
	// needs to connect in any available way.

	switch u.Scheme {
	case "http":
	case "https":
	case autoTLSScheme:
	default:
		return nil, fmt.Errorf("unsupported HTTP scheme %q", u.Scheme)
	}

	if port := u.Port(); port != "" {
		portNum, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("invalid HTTP port %q", port)
		}
		if portNum < 0 || portNum > 65535 {
			return nil, fmt.Errorf("out of range HTTP port %q", port)
		}
	}

	if u.Path == "" {
		// normalizing empty path to "/"
		u.Path = "/"
	}

	return &clientWrapper{
		baseURL: u,
	}, nil
}

func (c *clientWrapper) Configure(ctx context.Context, data *core.ConfigureData) error {
	c.log = slog.New(data.Logger)
	c.client = &http.Client{
		Transport: &baseURLTransport{
			base:      otelhttp.NewTransport(http.DefaultTransport),
			targetURL: c.baseURL,
		},
	}

	return nil
}

func (c *clientWrapper) Acquire(ctx context.Context, data *core.AcquireData) error {
	if !c.pingEndpoint {
		return nil
	}

	c.log.Info("pinging endpoint", slog.String("addr", c.baseURL.String()))

	ctx, cancel := context.WithTimeout(ctx, DefaultRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("ping failed with status: %s", resp.Status)
	}

	c.log.Info("ping successful", slog.String("addr", c.baseURL.String()))
	return nil
}

func (c *clientWrapper) Shutdown(ctx context.Context, data *core.ShutdownData) error {
	c.client.CloseIdleConnections()

	return nil
}

func (c *clientWrapper) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func (c *clientWrapper) BaseURL() *url.URL {
	u := *c.baseURL
	if c.baseURL.User != nil {
		userinfo := *c.baseURL.User
		u.User = &userinfo
	}

	return &u
}

func (c *clientWrapper) Native() *http.Client { return c.client }

// baseURLTransport оборачивает другой RoundTripper и дополняет URL.
type baseURLTransport struct {
	base http.RoundTripper

	targetURL *url.URL
}

func (t *baseURLTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL != nil && req.URL.Scheme != "" && req.URL.Host != "" {
		// nothing changes, just round trip as is
		return t.base.RoundTrip(req)
	}

	// since it parses each time — we can write queries into this map
	queries := t.targetURL.Query()
	maps.Insert(queries, maps.All(req.URL.Query()))

	joinedPath, err := url.JoinPath(t.targetURL.Path, strings.TrimPrefix(req.URL.Path, "/"))
	if err != nil {
		return nil, fmt.Errorf("failed to join paths %q and %q: %w", t.targetURL.Path, req.URL.Path, err)
	}

	reqClone := req.Clone(req.Context())
	reqClone.Host = cmp.Or(req.Host, t.targetURL.Host)
	reqClone.URL = &url.URL{
		Scheme:      cmp.Or(req.URL.Scheme, t.targetURL.Scheme),
		Opaque:      t.targetURL.Opaque,
		User:        cmp.Or(req.URL.User, t.targetURL.User),
		Host:        cmp.Or(req.URL.Host, t.targetURL.Host),
		Path:        joinedPath,
		RawPath:     "",
		OmitHost:    t.targetURL.OmitHost,
		ForceQuery:  t.targetURL.ForceQuery,
		RawQuery:    queries.Encode(),
		Fragment:    cmp.Or(req.URL.Fragment, t.targetURL.Fragment),
		RawFragment: "",
	}

	return t.base.RoundTrip(reqClone)
}

// should be called only if client set in "auto" mode to verify that server
// supports TLS
func CheckServerTLS(ctx context.Context, u *url.URL) (bool, error) {

	// using basic http client with short timeout firstly.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			// Отключаем KeepAlive, нам нужен один быстрый тест
			DisableKeepAlives: true,
			// Жесткие таймауты на handshake и заголовки
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: -1,
			}).DialContext,
			ResponseHeaderTimeout: 2 * time.Second,
		},
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, u.String(), http.NoBody)
	if err != nil {
		panic(fmt.Sprintf("unexpected error creating HEAD request to %q: %v", u.String(), err))
	}
	req.Header.Set("User-Agent", "Http-Probe (like curl)")

	resp, err := client.Do(req)
	if err != nil {
		// TODO: cover all possible errors related to http request!
		return false, fmt.Errorf("unexpected error sending HEAD request to %q: %w", u.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		s := string(body)

		// Heuristics of HTTPS-only services:
		// 1) If response contains "SSL" or "TLS" words — very likely it's HTTPS-only service.
		if strings.Contains(s, "HTTPS") && strings.Contains(s, "port") {
			return true, nil // Detected! This is a TLS port.
		}
	}

	if resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest {
		// Redirects to HTTPS?
		loc, err := resp.Location()
		if err == nil && loc.Scheme == "https" {
			return true, nil
		}
	}

	// using plain http.
	return false, nil
}

// Reference: https://github.com/golang/go/blob/82c371a3/src/net/url/url.go#L444
func setDefaultScheme(raw, scheme string) string {
	for i := 0; i < len(raw); i++ {
		c := raw[i]

		switch {
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z':
		// do nothing
		case '0' <= c && c <= '9' || c == '+' || c == '-' || c == '.':
			if i == 0 {
				return scheme + "://" + raw
			}
		case c == ':':
			// checking next two slashes
			if i != 0 && len(raw) > i+2 && raw[i+1] == '/' && raw[i+2] == '/' {
				return raw
			}

			return scheme + "://" + raw
		default:
			// we have encountered an invalid character,
			// so there is no valid scheme
			return scheme + "://" + raw
		}
	}

	return scheme + "://" + raw
}
