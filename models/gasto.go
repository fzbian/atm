package models

import "time"

type GastoLocal struct {
	ID        int32     `json:"id" gorm:"primaryKey;autoIncrement"`
	Local     string    `json:"local"`
	Fecha     time.Time `json:"fecha"`
	Tipo      string    `json:"tipo"` // e.g., "GASTO_OPERATIVO"
	Motivo    string    `json:"motivo"`
	Monto     float64   `json:"monto"`
	ImagenURL string    `json:"imagen_url"`
	Usuario   string    `json:"usuario"`
}

func (GastoLocal) TableName() string {
	return "gastos_locales"
}
