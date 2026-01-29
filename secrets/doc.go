// Package secrets defines interfaces and helpers for retrieving secret values
// (credentials, configuration blobs) from pluggable engines. Engines abstract
// underlying storage backends while Secret values encapsulate retrieval and
// optional lazy decoding policies.
package secrets
