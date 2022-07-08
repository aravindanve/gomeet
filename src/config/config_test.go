package config

import "testing"

func TestNewConfig(t *testing.T) {
	t.Parallel()
	var _ = NewConfig()
}
