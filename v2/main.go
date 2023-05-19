package main

import (
	log2ch "flomesh.io/log4pipy/clickhouse"
	log2pg "flomesh.io/log4pipy/postgresql"

	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	// DB_ENV_CLICKHOUSE string = "clickhouse"
	DB_ENV_POSTGRESQL string = "postgresql"
)

func main() {
	// load ENV
	err := godotenv.Load()
	if err != nil {
		fmt.Println(".env file is missed")
	}
	dbType := os.Getenv("DB_TYPE")
	switch dbType {
	case DB_ENV_POSTGRESQL:
		fmt.Println("Loading postgresql driver...")
		log2pg.Run()
	default:
		fmt.Println("Loading clickhouse driver...")
		log2ch.Run()
	}
}
