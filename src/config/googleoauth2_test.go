package config

import "testing"

func TestNewGoogleOAuth2ConfigProvider(t *testing.T) {
	t.Parallel()
	var _ = NewGoogleOAuth2ConfigProvider()
}
