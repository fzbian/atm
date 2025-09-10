package database

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Convierte un URI tipo mysql://user:pass@host:port/db a DSN para GORM
func uriToDSN(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if u.Scheme != "mysql" {
		return "", fmt.Errorf("solo se soporta mysql://")
	}
	user := u.User.Username()
	pass, _ := u.User.Password()
	host := u.Host
	db := u.Path
	if len(db) > 0 && db[0] == '/' {
		db = db[1:]
	}
	// charset, parseTime y loc son recomendados para GORM
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, db), nil
}

// Connect abre la conexión a la base de datos MySQL usando solo DB_URI
func Connect() (*gorm.DB, error) {
	_ = godotenv.Load()

	dburi := os.Getenv("DB_URI")
	if dburi == "" {
		return nil, fmt.Errorf("DB_URI no está definido en el entorno")
	}
	dsn, err := uriToDSN(dburi)
	if err != nil {
		return nil, fmt.Errorf("DB_URI inválido: %w", err)
	}
	log.Printf("[DB] Conectando a %s", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	return db, nil
}
