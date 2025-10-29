package secrets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/quenbyako/core/secrets"
	"github.com/vincent-petithory/dataurl"
)

type multiEngine struct {
	closed atomic.Bool

	storages map[string]secrets.Engine
}

func BuildSecretEngine(ctx context.Context, u map[string]*url.URL) (secrets.Engine, error) {
	if len(u) == 0 {
		return &multiEngine{}, nil
	}

	storages := make(map[string]secrets.Engine, len(u))
	for scheme, url := range u {
		if url == nil {
			// urls MIGHT be nil, cause user doesn't call them each time.
			//
			// However, we should throw an error that we know about this type,
			// but user just didn't provide it.
			storages[scheme] = secrets.NewUnsetStorage(scheme)
			continue
		}

		storage, err := newSecretStorage(ctx, url)
		if err != nil {
			return &multiEngine{}, fmt.Errorf("creating storage for scheme %q: %w", scheme, err)
		}
		storages[scheme] = storage
	}
	return &multiEngine{storages: storages}, nil
}

func (e *multiEngine) GetSecret(ctx context.Context, addr string) (secrets.Secret, error) {
	if e.closed.Load() {
		return nil, io.ErrClosedPipe
	}

	if addr == "" {
		return secrets.NewEmptySecret(), nil
	}

	// data is not correct url scheme, cause usually we are using data:// or
	// something like this.
	//
	// Still, we had to check it in that way.
	if strings.HasPrefix(addr, "data:") {
		data, err := dataurl.DecodeString(addr)
		if err != nil {
			return nil, fmt.Errorf("decoding data URL: %w", err)
		}

		return secrets.NewPlainSecret(data.Data), nil
	}

	key, err := url.Parse(addr)
	if err != nil {

		return nil, fmt.Errorf("parsing secret URL %q: %w", addr, err)
	}

	storage, ok := e.storages[key.Scheme]
	if !ok {
		return nil, fmt.Errorf("no storage for scheme %q", key.Scheme)
	}

	secret, err := storage.GetSecret(ctx, key.Opaque)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from storage: %w", err)
	}

	return secret, nil
}

func (e *multiEngine) Close() error {
	if e.closed.CompareAndSwap(false, true) {
		return nil
	}

	var errs []error
	for _, s := range e.storages {
		if err := s.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
