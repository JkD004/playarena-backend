package db

import (
	"database/sql"
	"log"
	"os"
	_ "github.com/go-sql-driver/mysql" // The MySQL driver
)

var DB *sql.DB

// InitDB attempts to open and ping the database connection, returning an error on failure.
func InitDB() error {
	var err error
	
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Println("Error: DB_DSN environment variable not set")
		return os.ErrNotExist // Use a standard error type
	}

	// If DB is already initialized, just try to ping it
	if DB == nil {
		DB, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Println("Failed to open DB connection:", err)
			return err
		}
	}


	// Use Ping to check if the connection is alive and working
	err = DB.Ping()
	if err != nil {
		log.Println("Failed to ping DB:", err)
		// We do NOT close DB here, as it might just be temporary network issue
		return err
	}

	log.Println("Successfully connected to MySQL database!")
	return nil
}