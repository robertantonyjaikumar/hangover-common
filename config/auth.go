package config

type AuthConfig struct {
	AccessSecret  string
	RefreshSecret string
}

func LoadAuthConfig() AuthConfig {
	return AuthConfig{
		AccessSecret:  CFG.V.Get("ACCESS_SECRET").(string),
		RefreshSecret: CFG.V.Get("REFRESH_SECRET").(string),
	}
}
