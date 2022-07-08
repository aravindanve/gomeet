package config

import "testing"

func TestNewAuthConfigProvider(t *testing.T) {
	t.Parallel()
	var _ = NewAuthConfigProvider()
}
