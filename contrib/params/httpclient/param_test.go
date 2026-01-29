package httpclient_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/quenbyako/core"
	"github.com/quenbyako/core/contrib/runtime"

	. "github.com/quenbyako/core/contrib/params/httpclient"
)

func TestParam(t *testing.T) {
	t.Parallel()

	type Config struct {
		core.UnimplementedActionConfig

		Client Client `env:"HTTP_CLIENT"`
	}

	// TODO: remove runtime, we need to make test suite for creating correct
	// parameters test.

	// prepare runtime for parsing
	cmd := runtime.Run(func(ctx context.Context, appCtx core.AppContext[Config]) core.ExitCode {
		fmt.Println(appCtx.Config().Client)
		req, _ := http.NewRequest(http.MethodGet, "/pupalupa", http.NoBody)
		resp, err := appCtx.Config().Client.Do(req)
		if err != nil {
			t.Errorf("Request to /pupalupa failed: %v", err)
		} else {
			t.Logf("Response for /pupalupa: %v", resp.Status)
			resp.Body.Close()
		}

		req, _ = http.NewRequest(http.MethodGet, "http://example.com", http.NoBody)
		resp, err = appCtx.Config().Client.Do(req)
		if err != nil {
			t.Errorf("Request to http://example.com failed: %v", err)
		} else {
			t.Logf("Response for http://example.com: %v", resp.Status)
			resp.Body.Close()
		}

		return 0
	})

	ctx := runtime.WithEnvContext(t.Context(), map[string]string{
		"HTTP_CLIENT": "http://bad2b333d44a.ngrok-free.app",
	})

	cmd(ctx, []string{})
}
