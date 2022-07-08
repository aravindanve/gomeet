package config

type GoogleOAuth2Config struct {
	ClientID string
}

type GoogleOAuth2ConfigProvider interface {
	GoogleOAuth2Config() GoogleOAuth2Config
}

type googleOAuth2ConfigProvider struct {
	googleOAuth2Config GoogleOAuth2Config
}

func (p *googleOAuth2ConfigProvider) GoogleOAuth2Config() GoogleOAuth2Config {
	return p.googleOAuth2Config
}

func NewGoogleOAuth2ConfigProvider() GoogleOAuth2ConfigProvider {
	return &googleOAuth2ConfigProvider{
		googleOAuth2Config: GoogleOAuth2Config{
			ClientID: GetenvString("GOOGLE_OAUTH2_CLIENT_ID"),
		},
	}
}
