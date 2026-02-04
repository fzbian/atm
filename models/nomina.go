package models

import "time"

// NominaConfig guarda configuraciones globales de la nómina
type NominaConfig struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	AuxilioTransporte int64     `json:"auxilio_transporte"` // Valor mensual
	ValorDominical    int64     `json:"valor_dominical"`    // Deprecated: Use S1/S2
	ValorDominicalS1  int64     `json:"valor_dominical_s1"` // Valor Semestre 1 (Ene-Jun)
	ValorDominicalS2  int64     `json:"valor_dominical_s2"` // Valor Semestre 2 (Jul-Dic)
	ValorMadrugon     int64     `json:"valor_madrugon"`     // Valor por hora
	PorcentajeSalud   float64   `json:"porcentaje_salud"`   // % (e.g. 4.0)
	PorcentajePension float64   `json:"porcentaje_pension"` // % (e.g. 4.0)
	SalarioMinimo     int64     `json:"salario_minimo"`     // Referencia (opcional)
	CompanyName       string    `json:"company_name"`       // Nombre Empresa para recibos
	NIT               string    `json:"nit"`                // NIT Empresa para recibos
	UpdatedAt         time.Time `json:"updated_at"`
}

func (NominaConfig) TableName() string { return "nomina_configs" }

// UserPayroll guarda información específica de nómina por empleado
type UserPayroll struct {
	UserID      uint      `gorm:"primaryKey" json:"user_id"`
	BaseSalary  int64     `json:"base_salary"`  // Salario base mensual
	HasSecurity *bool     `json:"has_security"` // Si paga seguridad social (puntero para permitir false/true explícito en JSON, aunque GORM maneja bool)
	UpdatedAt   time.Time `json:"updated_at"`
}

func (UserPayroll) TableName() string { return "user_payrolls" }

// NominaPayment guarda el histórico de pagos generados
type NominaPayment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PeriodStart time.Time `gorm:"index" json:"period_start"`
	PeriodEnd   time.Time `gorm:"index" json:"period_end"`

	// Valores calculados y guardados
	BaseSalary      int64   `json:"base_salary"`   // Salario base (mensual) snapshot
	PaidBase        int64   `json:"paid_base"`     // (Base / 2)
	TransportAid    int64   `json:"transport_aid"` // (Auxilio / 2)
	SundaysQty      int     `json:"sundays_qty"`
	SundaysTotal    int64   `json:"sundays_total"`    // (Valor * Qty)
	MadrugonesQty   float64 `json:"madrugones_qty"`   // Horas
	MadrugonesTotal int64   `json:"madrugones_total"` // Valor total madrugones

	IncludesSecurity bool  `json:"includes_security"` // Si incluyó seguridad social
	Health           int64 `json:"health"`            // Deducción 4%
	Pension          int64 `json:"pension"`           // Deducción 4%
	Advance          int64 `json:"advance"`           // Adelanto descontado

	// JSON fields for extensibility
	Aditions   string `gorm:"type:json" json:"aditions"`   // []{Label, Value}
	Deductions string `gorm:"type:json" json:"deductions"` // []{Label, Value}

	TotalPaid int64 `json:"total_paid"` // El neto pagado

	Notes      string    `gorm:"size:255" json:"notes"`
	CreatedBy  string    `json:"created_by"`  // Usuario que generó el pago
	IsSigned   bool      `json:"is_signed"`   // Si el contrato está firmado
	SignedFile string    `json:"signed_file"` // Ruta al archivo del contrato (PDF)
	CreatedAt  time.Time `json:"created_at"`
}

func (NominaPayment) TableName() string { return "nomina_payments" }
