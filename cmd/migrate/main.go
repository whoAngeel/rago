package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("failed to create migrate: %v", err)
	}

	if err := m.Up(); err != nil {
		if strings.Contains(err.Error(), "no change") {
			fmt.Println("no new migrations to apply")
			return
		}
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Println("migrations applied successfully")
}
