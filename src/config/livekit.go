package config

import "time"

const (
	liveKitRoomTokenTTL = 15 * time.Minute
)

type LiveKitConfig struct {
	APIURL       string
	APIKey       string
	APISecret    string
	RoomTokenTTL time.Duration
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
			RoomTokenTTL: liveKitRoomTokenTTL,
		},
	}
}
