package client

import (
	"testing"

	"github.com/aravindanve/livemeet-server/src/config"
)

func TestNewLiveKitClient(t *testing.T) {
	t.Parallel()
	p := config.NewLiveKitConfigProvider()
	var _ = NewLiveKitClient(p)
}
