package models

import "time"

// BillingMonthly almacena gastos y margen por local y mes.
type BillingMonthly struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	PosName       string    `json:"pos_name" gorm:"size:191;index:idx_pos_month,unique"`
	Year          int       `json:"year" gorm:"index:idx_pos_month,unique"`
	Month         int       `json:"month" gorm:"index:idx_pos_month,unique"` // 1-12
	GastosComunes float64   `json:"gastos_comunes"`
	Servicios     float64   `json:"servicios"`
	Nomina        float64   `json:"nomina"`
	Arriendo      float64   `json:"arriendo"`
	Margen        float64   `json:"margen"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (BillingMonthly) TableName() string {
	return "billing_monthlies"
}
