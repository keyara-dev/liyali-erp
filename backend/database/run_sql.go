//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Note: .env file not found, using environment variables")
	}

	// Set default values if not provided
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "5432")
	}
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "postgres")
	}
	if os.Getenv("DB_SSL_MODE") == "" {
		os.Setenv("DB_SSL_MODE", "disable")
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run database/run_sql.go <sql_file.sql>")
	}

	sqlFile := os.Args[1]
	
	// Check if file exists
	if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
		log.Fatalf("SQL file does not exist: %s", sqlFile)
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Read SQL file
	content, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("Failed to read SQL file: %v", err)
	}

	// Execute SQL
	fmt.Printf("Running SQL file: %s\n", sqlFile)
	
	rows, err := db.Query(string(content))
	if err != nil {
		log.Fatalf("SQL execution failed: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("Failed to get columns: %v", err)
	}

	// Print results if any
	if len(columns) > 0 {
		fmt.Println("Results:")
		
		// Print header
		for i, col := range columns {
			if i > 0 {
				fmt.Print("\t")
			}
			fmt.Print(col)
		}
		fmt.Println()

		// Print rows
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		for rows.Next() {
			err := rows.Scan(valuePtrs...)
			if err != nil {
				log.Fatalf("Failed to scan row: %v", err)
			}

			for i, val := range values {
				if i > 0 {
					fmt.Print("\t")
				}
				if val != nil {
					fmt.Print(val)
				} else {
					fmt.Print("NULL")
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("✅ SQL executed successfully!")
}