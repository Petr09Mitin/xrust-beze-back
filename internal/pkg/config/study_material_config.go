package config

import "github.com/spf13/viper"

type StudyMaterial struct {
	Mongo *Mongo `mapstructure:"mongo"`
	Kafka *Kafka `mapstructure:"kafka"`
}

func NewStudyMaterial() (*StudyMaterial, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(StudyMaterialConfigPath)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	c := &StudyMaterial{}
	err = v.Unmarshal(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
