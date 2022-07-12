package config

import "testing"

func TestNewLiveKitConfigProvider(t *testing.T) {
	t.Parallel()
	var _ = NewLiveKitConfigProvider()
}
