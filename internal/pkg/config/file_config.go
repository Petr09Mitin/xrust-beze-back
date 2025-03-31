package config

import "github.com/spf13/viper"

type File struct {
	HTTP  *HTTP  `mapstructure:"http"`
	GRPC  *GRPC  `mapstructure:"grpc"`
	Minio *Minio `mapstructure:"minio"`
}

func NewFile() (*File, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(FileConfigPath)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	f := &File{}
	err = v.Unmarshal(f)
	if err != nil {
		return nil, err
	}
	return f, nil
}
