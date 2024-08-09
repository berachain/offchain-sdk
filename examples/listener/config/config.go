package config

import (
	"time"

	"github.com/berachain/offchain-sdk/v2/telemetry"
	"github.com/berachain/offchain-sdk/v2/tools/limiter"
)

type SubStruct struct {
	AddressToListen string
	EventName       string
}

type PollingStruct struct {
	Interval time.Duration
}

type Jobs struct {
	Sub    SubStruct
	Poller PollingStruct
}

type Config struct {
	Jobs      Jobs
	Metrics   telemetry.Config
	RateLimit limiter.Config
}
