package config

type Kafka struct {
	Version   string   `mapstructure:"version"`
	Addresses []string `mapstructure:"addresses"`
}
