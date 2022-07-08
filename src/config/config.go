package config

type Config struct {
	HttpConfigProvider
	MongoConfigProvider
	GoogleOAuth2ConfigProvider
	AuthConfigProvider
}

func NewConfig() *Config {
	return &Config{
		HttpConfigProvider:         NewHttpConfigProvider(),
		MongoConfigProvider:        NewMongoConfigProvider(),
		GoogleOAuth2ConfigProvider: NewGoogleOAuth2ConfigProvider(),
		AuthConfigProvider:         NewAuthConfigProvider(),
	}
}
