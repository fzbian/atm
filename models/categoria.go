package models

import "time"

// Categoria representa la tabla categorias
type Categoria struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Nombre    string    `gorm:"size:100;not null" json:"nombre"`
	Tipo      string    `gorm:"type:enum('INGRESO','EGRESO');not null" json:"tipo"`
	CreatedAt time.Time `json:"created_at" gorm:"-"`
}

func (Categoria) TableName() string { return "categorias" }
