package state

import "time"

type Config struct {
	Interval time.Duration
	Suffix   string
}

func NewConfig(interval time.Duration, suffix string) *Config {
	return &Config{
		Interval: interval,
		Suffix:   suffix,
	}
}
