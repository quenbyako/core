package core

import (
	"os"
	"sync"
)

var (
	defaultVersion = sync.OnceValue(func() AppVersion {
		return NewVersion(DefaultVersion, DefaultCommit, DefaultDate)
	})

	defaultPipeline = sync.OnceValue(func() Pipeline {
		return PipelineFromFiles(os.Stdin, os.Stdout, os.Stderr)
	})

	defaultAppName = sync.OnceValue(func() AppName {
		return NewAppName(DefaultAppName, DefaultAppTitle)
	})
)
