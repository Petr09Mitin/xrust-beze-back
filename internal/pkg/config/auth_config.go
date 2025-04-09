package config

import "github.com/spf13/viper"

type Redis struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type Cookie struct {
	Name     string `mapstructure:"name"`
	Domain   string `mapstructure:"domain"`
	MaxAge   int    `mapstructure:"max_age"`
	Secure   bool   `mapstructure:"secure"`
	HttpOnly bool   `mapstructure:"http_only"`
}

type Auth struct {
	HTTP     *HTTP     `mapstructure:"http"`
	GRPC     *GRPC     `mapstructure:"grpc"`
	Redis    *Redis    `mapstructure:"redis"`
	Cookie   *Cookie   `mapstructure:"cookie"`
	Services *Services `mapstructure:"services"`
}

func NewAuth() (*Auth, error) {
	v := viper.New()
	v.AutomaticEnv()              // Поддержка переменных окружения (если не задано в файле)
	v.SetConfigName("auth")       // Имя файла (без расширения)
	v.SetConfigType("yaml")       // Формат (yaml, json, toml)
	v.AddConfigPath("./configs/") // Путь к папке с конфигом
	v.AddConfigPath(".")          // Текущая директория
	v.SetConfigFile(AuthConfigPath)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Auth{}
	err = v.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
