// Package port supplies environment-parsed network listener parameters with
// optional TLS wrapping.
package port

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"

	"github.com/quenbyako/core"
)

// Listener is an alias to [net.Listener] exposed for semantic clarity.
type Listener = net.Listener

func init() { //nolint:gochecknoinits // there is no other way to register parsers
	core.RegisterEnvParser(parseListener)
}

type netListenerWrapper struct {
	net.Listener

	network *url.URL
	config  *tls.Config
}

var (
	_ Listener      = (*netListenerWrapper)(nil)
	_ core.EnvParam = (*netListenerWrapper)(nil)
)

//nolint:ireturn // returns interface on intention.
func parseListener(ctx context.Context, v string) (Listener, error) {
	uri, err := url.Parse(v)
	if err != nil {
		return nil, fmt.Errorf("parsing listener URL %q: %w", v, err)
	}

	return &netListenerWrapper{
		Listener: nil, // will be initialized later
		network:  uri,
		config:   nil,
	}, nil
}

func (l *netListenerWrapper) Configure(ctx context.Context, data *core.ConfigureData) error {
	l.config = &tls.Config{
		MinVersion:                          tls.VersionTLS12,
		Rand:                                nil,
		RootCAs:                             data.Pool,
		ClientCAs:                           data.Pool,
		InsecureSkipVerify:                  false,
		PreferServerCipherSuites:            true, // TODO: any other options
		Time:                                nil,
		Certificates:                        nil,
		NameToCertificate:                   nil,
		GetCertificate:                      nil,
		GetClientCertificate:                nil,
		GetConfigForClient:                  nil,
		VerifyPeerCertificate:               nil,
		VerifyConnection:                    nil,
		NextProtos:                          nil,
		ServerName:                          "",
		ClientAuth:                          0,
		CipherSuites:                        nil,
		SessionTicketsDisabled:              false,
		SessionTicketKey:                    [32]byte{},
		ClientSessionCache:                  nil,
		UnwrapSession:                       nil,
		WrapSession:                         nil,
		MaxVersion:                          0,
		CurvePreferences:                    nil,
		DynamicRecordSizingDisabled:         false,
		Renegotiation:                       0,
		KeyLogWriter:                        nil,
		EncryptedClientHelloConfigList:      nil,
		EncryptedClientHelloRejectionVerify: nil,
		GetEncryptedClientHelloKeys:         nil,
		EncryptedClientHelloKeys:            nil,
	}

	return nil
}

func (l *netListenerWrapper) Acquire(ctx context.Context, data *core.AcquireData) (err error) {
	listenConfig := &net.ListenConfig{
		Control:   nil,
		KeepAlive: 0,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable:   false,
			Idle:     0,
			Interval: 0,
			Count:    0,
		},
	}

	l.Listener, err = listenConfig.Listen(ctx, l.network.Scheme, l.network.Host)
	if err != nil {
		// TODO: handle error
		return fmt.Errorf("listening on %q %q: %w", l.network.Scheme, l.network.Host, err)
	}

	if l.config != nil {
		l.Listener = tls.NewListener(l.Listener, l.config)
	}

	return nil
}

func (l *netListenerWrapper) Shutdown(ctx context.Context, data *core.ShutdownData) error {
	if err := l.Close(); err != nil {
		return fmt.Errorf("closing connection: %w", err)
	}

	return nil
}
