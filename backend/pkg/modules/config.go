package modules

import "time"

type PostgreConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	DBName      string
	SSLMode     string
	ExecTimeout time.Duration
}

type AppConfig struct {
	Env             string
	HTTPPort        string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	WorkerInterval  time.Duration
}
