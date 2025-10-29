package core

import (
	"context"
	"io"
	"os"
)

type ctxPipelineKey struct{}

// Pipeline captures standard stream handles plus a flag indicating whether
// stdin appears to be connected to a non-tty source (e.g., pipe or file) for
// pipeline-aware behavior. Accessors return the raw handles; callers should
// not close them directly.
type Pipeline struct {
	stdin      io.Reader
	stdout     io.Writer
	stderr     io.Writer
	isPipeline bool
}

// PipelineFromFiles builds a Pipeline from explicit [*os.File] handles (nil
// values are replaced with process defaults). The IsPipeline bit is derived
// by inspecting stdin's mode to detect non-interactive usage.
func PipelineFromFiles(stdin, stdout, stderr *os.File) Pipeline {
	if stdin == nil {
		stdin = os.Stdin
	}

	if stdout == nil {
		stdout = os.Stdout
	}

	if stderr == nil {
		stderr = os.Stderr
	}

	return Pipeline{
		stdin:      stdin,
		stdout:     stdout,
		stderr:     stderr,
		isPipeline: isPipeline(stdin),
	}
}

// WithPipelines stores a [Pipeline] in a derived context for later retrieval.
// Use [PipelinesFromContext] to extract it; if absent, a cached default is
// provided.
func WithPipelines(ctx context.Context, p Pipeline) context.Context {
	return context.WithValue(ctx, ctxPipelineKey{}, p)
}

// PipelinesFromContext retrieves a [Pipeline] previously attached with
// [WithPipelines]. The boolean reports whether an explicit value was set.
// When false a lazily-created default wrapping the process stdio streams
// is returned.
func PipelinesFromContext(ctx context.Context) (Pipeline, bool) {
	if p, ok := ctx.Value(ctxPipelineKey{}).(Pipeline); ok {
		return p, true
	}

	return defaultPipeline(), false
}

// Stdin returns the input stream associated with the pipeline.
func (p Pipeline) Stdin() io.Reader { return p.stdin }

// Stdout returns the output stream associated with the pipeline.
func (p Pipeline) Stdout() io.Writer { return p.stdout }

// Stderr returns the error stream associated with the pipeline.
func (p Pipeline) Stderr() io.Writer { return p.stderr }

// IsPipeline reports whether stdin appears to be a non-interactive source
// (pipe/file). This is useful for adapting behavior (e.g., buffered reads).
func (p Pipeline) IsPipeline() bool { return p.isPipeline }

func isPipeline(in *os.File) bool {
	stat, err := in.Stat()
	if err != nil {
		return false
	}

	return (stat.Mode() & os.ModeCharDevice) == 0
}
