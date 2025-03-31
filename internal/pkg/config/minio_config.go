package config

const (
	TempMinioBucket    = "temp"
	AvatarsMinioBucket = "avatars"
)

type Minio struct {
	Endpoint string `mapstructure:"endpoint"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	UseSSL   bool   `mapstructure:"use_ssl"`
}
