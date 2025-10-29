package core

import (
	"context"
	"crypto/sha1" //nolint:gosec // this is a git hash algorithm
	"encoding/hex"
	"time"

	"golang.org/x/mod/semver"
)

type ctxVersionKey struct{}

// WithVersion attaches an [AppVersion] to a derived context for later
// retrieval via [VersionFromContext]. The stored value is immutable.
func WithVersion(ctx context.Context, v AppVersion) context.Context {
	return context.WithValue(ctx, ctxVersionKey{}, v)
}

// VersionFromContext extracts an [AppVersion] previously attached with
// [WithVersion]. When absent it returns a lazily constructed default and
// false. Callers can use the boolean to differentiate explicit vs.
// fallback version data.
func VersionFromContext(ctx context.Context) (AppVersion, bool) {
	if v, ok := ctx.Value(ctxVersionKey{}).(AppVersion); ok {
		return v, true
	}

	return defaultVersion(), false
}

const (
	DefaultVersion = "v0.0.0-dev"
	DefaultCommit  = "none"
	DefaultDate    = "unknown"

	DefaultDateFormat = time.RFC3339
)

// AppVersion represents build-time version metadata including semantic
// version string, commit hash and build date. Raw values preserve original
// inputs; normalized fields store validated / parsed representations.
// [AppVersion.Valid] reports aggregate correctness allowing callers to warn on
// malformed embeddings without failing hard.
type AppVersion struct {
	date         time.Time
	versionRaw   string
	commitRaw    string
	dateRaw      string
	version      string
	commit       [sha1.Size]byte
	versionValid bool
	commitValid  bool
	dateValid    bool
}

// NewVersion creates a version abstraction, based on raw compiled static
// variables.
//
// IMPORTANT: this function WON'T panic or return an error on bad input. Since
// it's expected to provide correct values on compile-time, returning error or
// panicking might broke whole build.
//
// Instead, Version has [AppVersion.Valid] method, that returns whether the object
// is correct or not. Caller may use this info to warn user that version info is
// invalid.
func NewVersion(versionRaw, commitRaw, dateRaw string) AppVersion {
	if versionRaw == "" {
		versionRaw = DefaultVersion
	}

	if commitRaw == "" {
		commitRaw = DefaultCommit
	}

	if dateRaw == "" {
		dateRaw = DefaultDate
	}

	version, versionValid := buildVersion(versionRaw)
	commit, commitHashValid := buildCommitHash(commitRaw)
	date, dateValid := buildDate(dateRaw)

	return AppVersion{
		versionRaw: versionRaw,
		commitRaw:  commitRaw,
		dateRaw:    dateRaw,

		version: version,
		commit:  commit,
		date:    date,

		versionValid: versionValid,
		commitValid:  commitHashValid,
		dateValid:    dateValid,
	}
}

// Valid reports whether version, commit hash and build date were all parsed
// successfully.
func (v AppVersion) Valid() bool {
	return v.versionValid && v.commitValid && v.dateValid
}

// Version returns the normalized semantic version string and a boolean
// denoting validity.
func (v AppVersion) Version() (string, bool) { return v.version, v.versionValid }

// CommitHash returns the parsed full commit hash bytes and validity.
func (v AppVersion) CommitHash() ([sha1.Size]byte, bool) { return v.commit, v.commitValid }

// Date returns the parsed build timestamp and validity.
func (v AppVersion) Date() (time.Time, bool) { return v.date, v.dateValid }

// VersionCommit combines semantic version and short commit hash as
// version#<short-hash>.
//
// More info: https://github.com/semver/semver/issues/614
func (v AppVersion) VersionCommit() (res string, valid bool) {
	version, versionValid := v.Version()
	commitShort, commitValid := v.ShortHash()

	return version + "#" + hex.EncodeToString(commitShort[:]), versionValid && commitValid
}

// String returns a human-friendly composite string including raw version,
// short commit hash (or raw commit when invalid) and build date (parsed or
// raw). This is suitable for logging.
func (v AppVersion) String() (res string) {
	res += v.versionRaw
	res += "-"

	if hashShort, ok := v.ShortHash(); ok {
		res += hex.EncodeToString(hashShort[:])
	} else {
		res += v.commitRaw
	}

	res += "-"
	if date, ok := v.Date(); ok {
		res += date.Format(DefaultDateFormat)
	} else {
		res += v.dateRaw
	}

	return res
}

// ShortHash extracts the first 7 bytes of the commit hash along with validity.
func (v AppVersion) ShortHash() (short [7]byte, valid bool) {
	copy(short[:], v.commit[:7])

	return short, v.commitValid
}

func buildVersion(v string) (version string, valid bool) {
	return v, semver.IsValid(v)
}

func buildCommitHash(h string) (hash [sha1.Size]byte, valid bool) {
	data, err := hex.DecodeString(h)
	if err != nil {
		copy(hash[:], h)
		return hash, false
	}

	copy(hash[:], data)

	return hash, len(data) == sha1.Size
}

func buildDate(d string) (date time.Time, valid bool) {
	date, err := time.Parse(DefaultDateFormat, d)
	if err != nil {
		return time.Time{}, false
	}

	return date, true
}
