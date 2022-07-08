package config

import "testing"

func TestNewHttpConfigProvider(t *testing.T) {
	t.Parallel()
	var _ = NewHttpConfigProvider()
}
