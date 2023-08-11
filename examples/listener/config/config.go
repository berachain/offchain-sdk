package config

import "time"

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
	Jobs Jobs
}
