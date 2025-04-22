package config

import "github.com/spf13/viper"

type StudyMaterialServices struct {
	FileService *GRPCService `mapstructure:"file_service"`
	UserService *GRPCService `mapstructure:"user_service"`
	AuthService *GRPCService `mapstructure:"auth_service"`
	MLTags      *GRPCService `mapstructure:"ai_tags_service"`
}

type StudyMaterial struct {
	Mongo    *Mongo                 `mapstructure:"mongo"`
	HTTP     *HTTP                  `mapstructure:"http"`
	GRPC     *GRPC                  `mapstructure:"grpc"`
	Kafka    *Kafka                 `mapstructure:"kafka"`
	Services *StudyMaterialServices `mapstructure:"services"`
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
