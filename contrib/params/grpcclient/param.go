package grpcclient

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/quenbyako/core"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

// Client abstracts a gRPC client connection and provides access to the
// underlying target information.
type Client interface {
	grpc.ClientConnInterface

	BaseTarget() resolver.Target
}

func init() {
	core.RegisterEnvParser(parseGRPCClient)
}

type clientWrapper struct {
	addr resolver.Target

	conn *grpc.ClientConn
	log  *slog.Logger
}

var _ core.EnvParam = (*clientWrapper)(nil)
var _ Client = (*clientWrapper)(nil)

func parseGRPCClient(ctx context.Context, v string) (Client, error) {
	v = setDefaultScheme(v, "grpc")

	u, err := url.Parse(v)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "grpc" {
		// this is default scheme, resolver works strangely, empty scheme means
		// direct connection.
		u.Scheme = ""
	}
	if u.Path == "/" {
		u.Path = ""
	}
	if u.Path == "" {
		u.Path = u.Host
		u.Host = ""
	}

	return &clientWrapper{addr: resolver.Target{URL: *u}}, nil
}

func targetReverse(t resolver.Target) string {
	endpoint := t.Endpoint()
	if t.URL.Scheme != "" {
		endpoint = t.URL.Scheme + "://" + t.URL.Host + "/" + endpoint
	}

	return endpoint
}

func (c *clientWrapper) Configure(ctx context.Context, data *core.ConfigureData) (err error) {
	c.log = otelslog.NewLogger("grpcclient", otelslog.WithLoggerProvider(data.Logger))

	c.conn, err = grpc.NewClient(targetReverse(c.addr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	return nil
}

func (c *clientWrapper) Acquire(ctx context.Context, data *core.AcquireData) error {
	return nil
}

func (c *clientWrapper) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return c.conn.Invoke(ctx, method, args, reply, opts...)
}

func (c *clientWrapper) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return c.conn.NewStream(ctx, desc, method, opts...)
}

func (c *clientWrapper) BaseTarget() resolver.Target { return c.addr }

func (h *clientWrapper) Shutdown(ctx context.Context, data *core.ShutdownData) error {
	err := h.conn.Close()

	return err
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
