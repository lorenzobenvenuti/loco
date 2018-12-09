package state

type Config struct {
	Interval string
	Suffix   string
}

func NewConfig(interval string, suffix string) *Config {
	return &Config{
		Interval: interval,
		Suffix:   suffix,
	}
}
