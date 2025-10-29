package secrets

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/quenbyako/core/secrets"
)

type FileStorage struct {
	secrets map[string]string
}

var _ secrets.Engine = (*FileStorage)(nil)

func NewFile(path string) (secrets.Engine, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	envs, err := godotenv.Parse(file)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		secrets: envs,
	}, nil
}

func (c *FileStorage) GetSecret(_ context.Context, key string) (secrets.Secret, error) {
	secret, ok := c.secrets[key]
	if !ok {
		return nil, secrets.ErrSecretNotFound
	}

	return secrets.NewPlainSecret([]byte(secret)), nil
}

func (c *FileStorage) Close() error { return nil }
