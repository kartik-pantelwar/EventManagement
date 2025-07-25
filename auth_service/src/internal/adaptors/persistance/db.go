package persistance

import (
	"authservice/src/internal/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	config, err := config.Loadconfig()
	if err != nil {
		log.Fatalf("Failed to Load Config :%v", err)
		return nil, err
	}

	dbUrl := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", config.DB_USER, config.DB_PASS, config.DB_HOST, config.DB_PORT, config.DB_NAME, config.DB_SSLMODE)

	OpenDb, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Failed to Open Database :%v", err)
	}

	return &Database{db: OpenDb}, nil
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}
