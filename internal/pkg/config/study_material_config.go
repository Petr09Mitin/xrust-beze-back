package config

import "github.com/spf13/viper"

type StudyMaterialServiceServices struct {
	User *GRPCService `mapstructure:"user_service"`
	File *GRPCService `mapstructure:"file_service"`
	Auth *GRPCService `mapstructure:"auth_service"`
}

type StudyMaterial struct {
	Mongo    *Mongo                        `mapstructure:"mongo"`
	HTTP     *HTTP                         `mapstructure:"http"`
	GRPC     *GRPC                         `mapstructure:"grpc"`
	Services *StudyMaterialServiceServices `mapstructure:"services"`
}

func NewStudyMaterial() (*StudyMaterial, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(StudyMaterialConfigPath)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	cfg := &StudyMaterial{}
	err = v.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
