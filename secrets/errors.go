package secrets

import (
	"errors"
)

var (
	// ErrSecretNotFound indicates that the requested secret address did not
	// resolve to an existing value within the engine.
	ErrSecretNotFound = errors.New("secret not found")
	// ErrSecretNotSet reports that a secret placeholder exists but has not yet
	// been populated with data.
	ErrSecretNotSet = errors.New("secret not set")
	// ErrEngineNotConfigured signals that a higher-level component attempted
	// to use an Engine that was not injected / initialized.
	ErrEngineNotConfigured = errors.New("secrets engine not configured")
)
