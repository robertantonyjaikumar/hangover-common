package config

type HTTPConfig struct {
	Port string
}

// LoadHTTPConfig returns server config
func LoadHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Port: CFG.V.GetString("server.port"),
	}
}
