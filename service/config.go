package service

var cacheConfig *Config

type Config struct {
	Port    int         `json:"port"`
	CrtFile string      `json:"crtFile"`
	KeyFile string      `json:"keyFile"`
	Static  StaticItem  `json:"static"`
	Proxies []ProxyItem `json:"proxies"`
	Limit   LimitItem   `json:"limit"`
	Debug   bool        `json:"debug"`
}

type LimitItem struct {
	Enable     bool     `json:"enable"`
	Delay      int64    `json:"delay"`
	Period     int64    `json:"period"`
	Times      int      `json:"times"`
	WhiteIps   []string `json:"whiteIps"`
	StatusCode int      `json:"statusCode"`
	Message    string   `json:"message"`
}

type StaticItem struct {
	Dir     string `json:"dir"`
	History bool   `json:"history"`
}

type ProxyItem struct {
	Uri     string `json:"uri"`
	Target  string `json:"target"`
	Forward bool   `json:"forward"`
}
