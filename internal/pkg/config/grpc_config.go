package config

type GRPC struct {
	Port int `mapstructure:"port"`
}

type GRPCService struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Timeout    int    `mapstructure:"timeout"`
	MaxRetries int    `mapstructure:"max_retries"`
}
