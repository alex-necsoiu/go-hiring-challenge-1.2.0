package database

import (
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// New creates a new PostgreSQL database connection.
// Parameters:
//   - user: PostgreSQL username
//   - password: PostgreSQL password
//   - dbname: Database name
//   - port: PostgreSQL port
//   - host: PostgreSQL hostname (default: localhost)
func New(user, password, dbname, port, host string) (db *gorm.DB, close func() error) {
	// Default to localhost if host is empty
	if host == "" {
		host = "localhost"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %s", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %s", err)
	}

	return db, sqlDB.Close
}
