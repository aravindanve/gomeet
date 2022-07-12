package config

type Config interface {
	HttpConfigProvider
	MongoConfigProvider
	GoogleOAuth2ConfigProvider
	LiveKitConfigProvider
	AuthConfigProvider
}

type config struct {
	HttpConfigProvider
	MongoConfigProvider
	GoogleOAuth2ConfigProvider
	LiveKitConfigProvider
	AuthConfigProvider
}

func NewConfig() Config {
	return &config{
		HttpConfigProvider:         NewHttpConfigProvider(),
		MongoConfigProvider:        NewMongoConfigProvider(),
		GoogleOAuth2ConfigProvider: NewGoogleOAuth2ConfigProvider(),
		LiveKitConfigProvider:      NewLiveKitConfigProvider(),
		AuthConfigProvider:         NewAuthConfigProvider(),
	}
}
