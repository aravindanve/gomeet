package config

type HttpConfig struct {
	Addr string
}

type HttpConfigProvider interface {
	HttpConfig() HttpConfig
}

type httpConfigProvider struct {
	httpConfig HttpConfig
}

func (p *httpConfigProvider) HttpConfig() HttpConfig {
	return p.httpConfig
}

func NewHttpConfigProvider() HttpConfigProvider {
	return &httpConfigProvider{
		httpConfig: HttpConfig{
			Addr: GetenvStringWithDefault("HTTP_ADDR", ":8080"),
		},
	}
}
