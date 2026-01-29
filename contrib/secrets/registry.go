package secrets

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/quenbyako/core/secrets"
	"github.com/vincent-petithory/dataurl"
)

func newSecretStorage(ctx context.Context, u *url.URL) (secrets.Engine, error) {
	switch u.Scheme {
	case "file":
		return NewFile(path.Join(u.Host, u.Path))
	case "vault":
		return NewVault(ctx, u)
	case "data":
		data, err := dataurl.DecodeString(u.String())
		if err != nil {
			return nil, err
		}
		return secrets.NewConstantStorage(data.Data), nil
	default:
		return nil, fmt.Errorf("unsupported secret storage scheme: %q", u.Scheme)
	}
}
