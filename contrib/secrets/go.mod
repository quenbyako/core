module github.com/quenbyako/core/contrib/secrets

go 1.25.1

replace github.com/quenbyako/core => ../..

require (
	github.com/hashicorp/go-retryablehttp v0.7.8
	github.com/hashicorp/vault/api v1.22.0
	github.com/hashicorp/vault/api/auth/cert v0.0.0-20251105000312-680e5b5b3b18
	github.com/joho/godotenv v1.5.1
	github.com/quenbyako/core v0.0.0-00010101000000-000000000000
	github.com/vincent-petithory/dataurl v1.0.0
)

require (
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-7 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)
