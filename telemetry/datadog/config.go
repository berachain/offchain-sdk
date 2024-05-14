package datadog

import "fmt"

type Config struct {
	Enabled    bool
	StatsdAddr string
	Namespace  string
}

func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.StatsdAddr == "" {
		return fmt.Errorf("invalid Datadog statsd address: %s", c.StatsdAddr)
	}

	if c.Namespace == "" {
		return fmt.Errorf("invalid Datadog namespace: %s", c.Namespace)
	}

	return nil
}
