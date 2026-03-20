package dbconn

import (
	"database/sql"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
)

var dbGlobal *sql.DB

func init() {
	cfg := newDbConfig()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.host, cfg.port, cfg.username, cfg.password, cfg.dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	autoMigrate(cfg)

	dbGlobal = db
}

func GetDB() *sql.DB {
	return dbGlobal
}

func autoMigrate(cfg *dbConfig) {
	sourceURl := "file://migrations"
	databaseURl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.username, cfg.password, cfg.host, cfg.port, cfg.dbname)

	m, err := migrate.New(sourceURl, databaseURl)
	if err != nil {
		panic(err)
	}

	err = m.Up()

	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
}
