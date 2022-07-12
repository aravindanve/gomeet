package config

import "time"

const (
	liveKitJoinTokenTTL = 15 * time.Minute
)

type LiveKitConfig struct {
	APIURL       string
	APIKey       string
	APISecret    string
	JoinTokenTTL time.Duration
}

type LiveKitConfigProvider interface {
	LiveKitConfig() LiveKitConfig
}

type livekitConfigProvider struct {
	livekitConfig LiveKitConfig
}

func (p *livekitConfigProvider) LiveKitConfig() LiveKitConfig {
	return p.livekitConfig
}

func NewLiveKitConfigProvider() LiveKitConfigProvider {
	return &livekitConfigProvider{
		livekitConfig: LiveKitConfig{
			APIURL:       GetenvString("LIVEKIT_API_URL"),
			APIKey:       GetenvString("LIVEKIT_API_KEY"),
			APISecret:    GetenvString("LIVEKIT_API_SECRET"),
			JoinTokenTTL: liveKitJoinTokenTTL,
		},
	}
}
