package dbconn

import "os"

type dbConfig struct {
	host     string
	port     string
	username string
	password string
	dbname   string
}

func newDbConfig() *dbConfig {
	return &dbConfig{
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		username: os.Getenv("DB_USERNAME"),
		password: os.Getenv("DB_PASSWORD"),
		dbname:   os.Getenv("DB_NAME"),
	}
}
