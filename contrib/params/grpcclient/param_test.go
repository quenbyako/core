package grpcclient_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/quenbyako/core"
	"github.com/quenbyako/core/contrib/params/grpc"
	"github.com/quenbyako/core/contrib/runtime"
	reflectionpb "google.golang.org/grpc/reflection/grpc_reflection_v1"

	. "github.com/quenbyako/core/contrib/params/grpcclient"
)

func TestParam(t *testing.T) {
	t.Parallel()

	type Config struct {
		core.UnimplementedActionConfig

		Server grpc.Server `env:"GRPC_SERVER"`
		Client Client      `env:"GRPC_CLIENT"`
	}

	// TODO: remove runtime, we need to make test suite for creating correct
	// parameters test.

	// prepare runtime for parsing
	cmd := runtime.Run(func(ctx context.Context, appCtx core.AppContext[Config]) core.ExitCode {
		innerCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		var wg sync.WaitGroup
		wg.Go(func() { appCtx.Config().Server.Serve(innerCtx) })

		c := reflectionpb.NewServerReflectionClient(appCtx.Config().Client)
		resp, err := c.ServerReflectionInfo(ctx)
		if err != nil {
			t.Errorf("failed to create reflection client: %v", err)
			return 1
		}
		resp.CloseSend()

		for {
			_, err := resp.Recv()
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				t.Errorf("failed to receive reflection response: %v", err)
				return 1
			}
		}

		cancel()

		wg.Wait()
		return 0
	})

	ctx := runtime.WithEnvContext(t.Context(), map[string]string{
		"GRPC_SERVER": "grpc://127.0.0.1:50051",
		"GRPC_CLIENT": "grpc://127.0.0.1:50051",
	})

	cmd(ctx, []string{})
}
