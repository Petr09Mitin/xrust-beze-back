package config

type Mongo struct {
	Host     string `mapstructure:"host" env:"MONGO_HOST"`
	Port     int    `mapstructure:"port" env:"MONGO_PORT"`
	Username string `mapstructure:"username" env:"MONGO_USERNAME"`
	Password string `mapstructure:"password" env:"MONGO_PASSWORD"`
	Database string `mapstructure:"database" env:"MONGO_DATABASE"`
}
