package database

import (
	"atm/models"
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
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	// Code-First Migration
	log.Println("[DB] Ejecutando migraciones automáticas...")
	if err = db.AutoMigrate(
		&models.Caja{},
		&models.Categoria{},
		&models.Transaccion{},
		&models.TransaccionLog{},
		&models.GastoLocal{},
		&models.RoleConfig{},
		&models.User{},
		&models.NominaConfig{},
		&models.UserPayroll{},
		&models.NominaPayment{},
	); err != nil {
		return nil, fmt.Errorf("fallo migracion automatica: %w", err)
	}

	// Seeding
	go func() {
		if err := Seed(db); err != nil {
			log.Printf("[DB] Error en seeding: %v", err)
		}
	}()

	return db, nil
}

// Seed puebla la base de datos con datos iniciales si está vacía
func Seed(db *gorm.DB) error {
	// 1. Cajas
	cajas := []models.Caja{
		{ID: 1, Saldo: 0},
		{ID: 2, Saldo: 0},
	}
	for _, c := range cajas {
		if err := db.FirstOrCreate(&c, models.Caja{ID: c.ID}).Error; err != nil {
			return err
		}
	}

	// 2. Categorias
	var count int64
	db.Model(&models.Categoria{}).Count(&count)
	if count == 0 {
		log.Println("[DB] Seeding categorias...")
		cats := []models.Categoria{
			{Nombre: "Venta", Tipo: "INGRESO"},
			{Nombre: "Recarga", Tipo: "INGRESO"},
			{Nombre: "Aporte de Capital", Tipo: "INGRESO"},
			{Nombre: "Gasto", Tipo: "EGRESO"},
			{Nombre: "Retirada", Tipo: "EGRESO"},
			{Nombre: "Pago Proveedor", Tipo: "EGRESO"},
			{Nombre: "Otro", Tipo: "EGRESO"},
			// ID 16 will be handled below to ensuring it exists even if count > 0
		}
		for _, cat := range cats {
			if err := db.Create(&cat).Error; err != nil {
				return err
			}
		}
	}

	// Ensure critical Category 16 exists (idempotent)
	cat16 := models.Categoria{ID: 16, Nombre: "Efectivos Puntos de Venta", Tipo: "INGRESO"}
	if err := db.FirstOrCreate(&cat16, models.Categoria{ID: 16}).Error; err != nil {
		log.Printf("[DB] Error asegurando categoria 16: %v", err)
	}

	return nil
}
