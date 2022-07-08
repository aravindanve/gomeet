package config

type MongoConfig struct {
	ConnectionURI string
}

type MongoConfigProvider interface {
	MongoConfig() MongoConfig
}

type mongoConfigProvider struct {
	mongoConfig MongoConfig
}

func (p *mongoConfigProvider) MongoConfig() MongoConfig {
	return p.mongoConfig
}

func NewMongoConfigProvider() MongoConfigProvider {
	return &mongoConfigProvider{
		mongoConfig: MongoConfig{
			ConnectionURI: GetenvString("MONGO_CONNECTION_URI"),
		},
	}
}
