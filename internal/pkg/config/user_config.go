package config

import "github.com/spf13/viper"

type User struct {
	Mongo *Mongo `mapstructure:"mongo"`
	HTTP  *HTTP  `mapstructure:"http"`
	GRPC  *GRPC  `mapstructure:"grpc"`
}

func NewUser() (*User, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(UserConfigPath)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	u := &User{}
	err = v.Unmarshal(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}
