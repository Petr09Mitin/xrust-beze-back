package config

import "github.com/spf13/viper"

type VoiceRecognitionDServices struct {
	FileService        *GRPCService `mapstructure:"file_service"`
	AIVoiceRecognition *GRPCService `mapstructure:"ai_voice_recognition_service"`
}

type VoiceRecognitionD struct {
	Mongo    *Mongo                     `mapstructure:"mongo"`
	Kafka    *Kafka                     `mapstructure:"kafka"`
	Services *VoiceRecognitionDServices `mapstructure:"services"`
}

func NewVoiceRecognitionD() (*VoiceRecognitionD, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(VoiceRecognitionDConfigPath)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	cfg := &VoiceRecognitionD{}
	err = v.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
