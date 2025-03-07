package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var DB *bun.DB

func InitDB() {
	//dsn := "postgres://postgres:12345@localhost:5432/test?sslmode=disable"
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or cannot be loaded")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := os.Getenv("POSTGRES_HOST")
		port := os.Getenv("POSTGRES_PORT")
		user := os.Getenv("POSTGRES_USER")
		password := os.Getenv("POSTGRES_PASSWORD")
		dbName := os.Getenv("POSTGRES_DB")
		fmt.Println("🔍 Debug ENV - POSTGRES_USER:", user)

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}
		if dbName == "" {
			dbName = "test"
		}
		if user == "" {
			user = "postgres"
		}
		if password == "" {
			password = "12345"
		}
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
		log.Println("DATABASE_URL is empty, using fallback", dsn)
	}

	// sqlDB, err := sql.Open("postgres", dsn)
	// if err != nil {
	// 	log.Fatalf("Failed to connnect to database: %v", err)
	// }
	var sqlDB *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		sqlDB, err = sql.Open("postgres", dsn)
		if err == nil {
			err = sqlDB.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Failed to connnect to database (attemp: %d): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connnect to database: %v", err)
	}
	DB = bun.NewDB(sqlDB, pgdialect.New())
	// if err := DB.Ping(); err != nil {
	// 	log.Fatalf("database connect 'test' failed: %v", err)
	// }
	var result int
	if err := DB.NewRaw("SELECT 1").Scan(context.Background(), &result); err != nil {
		log.Fatalf("database connect 'test' failed: %v", err)
	}
	log.Println("Database connected successfully")
}
