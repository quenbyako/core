package runtime

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
)

func loadCertificates(additionalPaths []string) *x509.CertPool {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		// On Windows, SystemCertPool() always returns nil, nil.
		// On other systems, a non-nil error means we couldn't get the system pool.
		// In either case, we create a new cert pool.
		fmt.Println("warning: failed to load system CA certificates, using empty cert pool") // TODO: use logger LATER
		certPool = x509.NewCertPool()
	}

	for _, globPath := range additionalPaths {
		// TODO: no os filesystem!!! only [fs.FS]!
		paths, err := filepath.Glob(globPath)
		if err != nil {
			// todo: for now panic, later return error
			panic(fmt.Errorf("parsing glob %q: %w", globPath, err))
		}

		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				// todo: for now panic, later return error
				panic(fmt.Errorf("reading CA certificate %q: %w", path, err))
			}

			cert, err := x509.ParseCertificate(data)
			if err != nil {
				// todo: for now panic, later return error
				panic(fmt.Errorf("parsing CA certificate %q: %w", path, err))
			}

			certPool.AddCert(cert)
		}
	}

	return certPool
}
