package dsp

type Config struct {
	FFTSize int
	HopSize int
}

func DefaultConfig() Config {
	return Config{FFTSize: 4096, HopSize: 2048}
}
