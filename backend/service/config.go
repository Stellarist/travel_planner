package service

type ModelConfig struct {
	ApiKey  string `json:"apikey"`
	BaseURL string `json:"baseurl"`
	Model   string `json:"model"`
}

type AppConfig struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
	Model ModelConfig `json:"model"`
}

var GlobalAppConfig AppConfig
