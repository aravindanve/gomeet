package config

type Config interface {
	HttpConfigProvider
	MongoConfigProvider
	GoogleOAuth2ConfigProvider
	AuthConfigProvider
}

type config struct {
	HttpConfigProvider
	MongoConfigProvider
	GoogleOAuth2ConfigProvider
	AuthConfigProvider
}

func NewConfig() Config {
	return &config{
		HttpConfigProvider:         NewHttpConfigProvider(),
		MongoConfigProvider:        NewMongoConfigProvider(),
		GoogleOAuth2ConfigProvider: NewGoogleOAuth2ConfigProvider(),
		AuthConfigProvider:         NewAuthConfigProvider(),
	}
}
