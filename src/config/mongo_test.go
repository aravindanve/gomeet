package config

import "testing"

func TestNewMongoConfigProvider(t *testing.T) {
	t.Parallel()
	var _ = NewMongoConfigProvider()
}
