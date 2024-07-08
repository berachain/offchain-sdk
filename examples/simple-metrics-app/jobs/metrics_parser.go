package jobs

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/berachain/offchain-sdk/job"
	sdk "github.com/berachain/offchain-sdk/types"
)

// Compile time check to ensure that Listener implements job.Basic.
var _ job.Polling = &Parser{}
var _ job.HasSetup = &Parser{}

// Listener is a simple job that logs the current block when it is run.
type Parser struct {
	Interval    time.Duration
	metricRegex *regexp.Regexp
}

func (Parser) RegistryKey() string {
	return "Parser"
}

func (w *Parser) IntervalTime(_ context.Context) time.Duration {
	return w.Interval
}

func (w *Parser) Setup(_ context.Context) error {
	w.metricRegex = regexp.MustCompile(`^(\w+)(\{[^}]*\})?\s+(\d+(\.\d+)?)$`)
	return nil
}

// Execute implements job.Basic.
func (w *Parser) Execute(ctx context.Context, _ any) (any, error) {
	sdkCtx := sdk.UnwrapContext(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/metrics", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		matches := w.metricRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		metricName := matches[1]
		if metricName == "example_listener_app_rpc_request_count" {
			tags := matches[2]
			value := matches[3]
			sdkCtx.Logger().Info("metric", "name", metricName, "tags", tags, "value", value)
			return true, nil
		}
	}
	sdkCtx.Logger().Warn("no matching metric found")
	return true, nil
}
