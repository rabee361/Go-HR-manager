package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite" // CGO-free SQLite driver
)

func connectToDB(ctx context.Context) (*sql.DB, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" {
		return nil, fmt.Errorf("database connection details are missing. Please set DB_HOST, DB_PORT, DB_USER, DB_NAME")
	}

	dbPath := fmt.Sprintf("%s.db", dbName)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}

	fmt.Printf("✅ Successfully connected to the SQLite database: %s\n", dbPath)

	// Initialize database schema from .sql file
	sqlFile := "db.sql"
	schema, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Printf("Warning: Could not read schema file %s: %v", sqlFile, err)
	} else {
		_, err = db.ExecContext(ctx, string(schema))
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("executing schema: %w", err)
		}
		fmt.Println("🚀 Database schema initialized successfully!")
	}

	return db, nil
}
