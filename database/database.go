package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Connect abre la conexión a la base de datos MySQL y devuelve *gorm.DB
func Connect() (*gorm.DB, error) {
	// Cargar .env (ignorar error si no existe)
	_ = godotenv.Load()

	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "root")
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	dbname := getEnv("DB_NAME", "atm")
	charset := getEnv("DB_CHARSET", "utf8mb4")
	loc := getEnv("DB_LOC", "Local")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s", user, password, host, port, dbname, charset, loc)
	log.Printf("[DB] Conectando a %s:%s/%s", host, port, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	return db, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
