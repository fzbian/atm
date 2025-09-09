package models

import "time"

// Transaccion representa la tabla transacciones
type Transaccion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CategoriaID uint      `gorm:"column:categoria_id;not null" json:"categoria_id"`
	Monto       float64   `gorm:"type:decimal(15,2);not null" json:"monto"`
	Fecha       time.Time `gorm:"column:fecha;autoCreateTime" json:"fecha"`
	Descripcion string    `gorm:"type:text;not null" json:"descripcion"`
}

// TransaccionCreateInput limita los campos permitidos al crear una transacción (id y fecha se generan automáticamente)
type TransaccionCreateInput struct {
	CategoriaID uint    `json:"categoria_id" binding:"required"`
	Monto       float64 `json:"monto" binding:"required"`
	Descripcion string  `json:"descripcion" binding:"required"`
}

// TransaccionUpdateInput ahora usa punteros para distinguir campos omitidos
// (solo se actualizan los que lleguen no nulos)
type TransaccionUpdateInput struct {
	CategoriaID *uint    `json:"categoria_id,omitempty"`
	Monto       *float64 `json:"monto,omitempty"`
	Descripcion *string  `json:"descripcion,omitempty"`
}

func (Transaccion) TableName() string { return "transacciones" }
