package config

import (
	"github.com/spf13/viper"
)

type HTTP struct {
	Port int `mapstructure:"port"`
}

type Services struct {
	UserService *GRPCService `mapstructure:"user_service"`
}

type Chat struct {
	HTTP     *HTTP     `mapstructure:"http"`
	Services *Services `mapstructure:"services"`
	Mongo    *Mongo    `mapstructure:"mongo"`
	Kafka    *Kafka    `mapstructure:"kafka"`
}

func NewChat() (*Chat, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(ChatConfigPath)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	c := &Chat{}
	err = v.Unmarshal(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
