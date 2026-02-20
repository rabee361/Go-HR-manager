package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite" // CGO-free SQLite driver
)

func connectToDB() (*sql.DB, error) {
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
		log.Fatal("Database connection details are missing. Please set DB_HOST, DB_PORT, DB_USER, DB_NAME.")
	}

	dbPath := fmt.Sprintf("%s.db", dbName)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	fmt.Printf("✅ Successfully connected to the SQLite database: %s\n", dbPath)

	// Initialize database schema from .sql file
	sqlFile := "db.sql"
	schema, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Printf("Warning: Could not read schema file %s: %v", sqlFile, err)
	} else {
		_, err = db.Exec(string(schema))
		if err != nil {
			log.Fatalf("Error executing schema: %v", err)
		}
		fmt.Println("🚀 Database schema initialized successfully!")
	}


	return db, nil
}
