package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func dsnFromEnv() string {
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}
	host := getenv("DB_HOST", "localhost")
	user := getenv("DB_USER", "app")
	pass := getenv("DB_PASSWORD", "app")
	name := getenv("DB_NAME", "videosdb")
	return fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", user, pass, host, name)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func MustOpen() *sql.DB {
	dsn := dsnFromEnv()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		panic(err)
	}
	return db
}
