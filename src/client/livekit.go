package client

import (
	"context"

	"github.com/aravindanve/livemeet-server/src/config"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
)

type LiveKitClientDeps interface {
	config.LiveKitConfigProvider
}

type LiveKitClientProvider interface {
	LiveKitClient() LiveKitClient
}

type LiveKitClient interface {
	SendData(ctx context.Context, req *livekit.SendDataRequest) (*livekit.SendDataResponse, error)
}

type liveKitClient struct {
	config     config.LiveKitConfig
	roomClient *lksdk.RoomServiceClient
}

func NewLiveKitClient(ds LiveKitClientDeps) LiveKitClient {
	cf := ds.LiveKitConfig()
	roomClient := lksdk.NewRoomServiceClient(cf.APIURL, cf.APIKey, cf.APISecret)

	return &liveKitClient{
		config:     cf,
		roomClient: roomClient,
	}
}

func (l *liveKitClient) SendData(ctx context.Context, req *livekit.SendDataRequest) (*livekit.SendDataResponse, error) {
	return l.roomClient.SendData(ctx, req)
}
