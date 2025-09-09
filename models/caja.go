package models

import "time"

// Caja representa la tabla caja
type Caja struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Saldo               float64   `gorm:"type:decimal(15,2);not null" json:"saldo"`
	UltimaActualizacion time.Time `gorm:"column:ultima_actualizacion;autoUpdateTime" json:"ultima_actualizacion"`
}

// CajaOdoo representa la respuesta agregada de saldos obtenidos desde Odoo
// swagger:model
type CajaOdoo struct {
	ID                  uint               `json:"id"`
	SaldoCaja           float64            `json:"saldo_caja"`
	Locales             map[string]float64 `json:"locales"`
	TotalLocales        float64            `json:"total_locales"`
	SaldoTotal          float64            `json:"saldo_total"`
	UltimaActualizacion time.Time          `json:"ultima_actualizacion"`
}

func (Caja) TableName() string { return "caja" }
