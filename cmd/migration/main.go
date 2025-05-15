package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	dbConnString := os.Getenv("DB")
	if dbConnString == "" {
		log.Fatal("DB environment variable not set")
	}

	migrationPath := os.Getenv("MIGRATION_PATH")
	if migrationPath == "" {
		migrationPath = "./migrations"
	}

	absPath, err := filepath.Abs(migrationPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	log.Printf("Using migrations from: %s", absPath)

	cmd := exec.Command(
		"go", "run", "-mod=mod",
		"github.com/golang-migrate/migrate/v4/cmd/migrate",
		"-path", absPath,
		"-database", dbConnString,
		"up",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		log.Fatalf("Migration failed: %v", err)
	}

	log.Printf("Migration output: %s", string(output))
	log.Println("Migration completed successfully")
}
