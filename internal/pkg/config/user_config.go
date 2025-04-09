package config

import "github.com/spf13/viper"

type UserServiceServices struct {
	File *GRPCService `mapstructure:"file_service"`
	Auth *GRPCService `mapstructure:"auth_service"`
}

type User struct {
	Mongo    *Mongo               `mapstructure:"mongo"`
	HTTP     *HTTP                `mapstructure:"http"`
	GRPC     *GRPC                `mapstructure:"grpc"`
	Services *UserServiceServices `mapstructure:"services"`
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
