package controllers

import (
	"atm/models"
	atmOdoo "atm/odoo"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// --- CONFIGURATION ---

// GetNominaConfig obtiene la configuración global
func GetNominaConfig(c *gin.Context) {
	var config []models.NominaConfig
	// Asumimos ID=1 para la configuración global
	DB.Limit(1).Find(&config, 1)

	if len(config) == 0 {
		// Retornar defaults si no existe
		c.JSON(http.StatusOK, models.NominaConfig{
			AuxilioTransporte: 162000,
			ValorDominical:    100000,
			ValorDominicalS1:  100000,
			ValorDominicalS2:  100000,
			PorcentajeSalud:   4.0,
			PorcentajePension: 4.0,
		})
		return
	}
	c.JSON(http.StatusOK, config[0])
}

// UpdateNominaConfig actualiza valores globales
func UpdateNominaConfig(c *gin.Context) {
	var input models.NominaConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var config models.NominaConfig
	if err := DB.First(&config, 1).Error; err != nil {
		config.ID = 1
	}

	config.AuxilioTransporte = input.AuxilioTransporte
	config.ValorDominical = input.ValorDominical
	config.ValorDominicalS1 = input.ValorDominicalS1
	config.ValorDominicalS2 = input.ValorDominicalS2
	config.ValorMadrugon = input.ValorMadrugon
	config.PorcentajeSalud = input.PorcentajeSalud
	config.PorcentajePension = input.PorcentajePension
	config.SalarioMinimo = input.SalarioMinimo
	config.CompanyName = input.CompanyName
	config.NIT = input.NIT
	config.UpdatedAt = time.Now()

	DB.Save(&config)
	c.JSON(http.StatusOK, config)
}

// --- EMPLOYEES ---

// GetNominaEmployees lista usuarios con su info de nómina optimizado
func GetNominaEmployees(c *gin.Context) {
	var users []models.User
	if err := DB.Order("name asc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Batch fetch all payroll records to avoid N+1 problem
	var payrolls []models.UserPayroll
	DB.Find(&payrolls)
	payrollMap := make(map[uint]*models.UserPayroll)
	for i := range payrolls {
		payrollMap[payrolls[i].UserID] = &payrolls[i]
	}

	type EmployeePayrollDTO struct {
		models.User
		Payroll *models.UserPayroll `json:"payroll"`
	}

	var result []EmployeePayrollDTO
	for _, u := range users {
		result = append(result, EmployeePayrollDTO{
			User:    u,
			Payroll: payrollMap[u.ID],
		})
	}

	c.JSON(http.StatusOK, result)
}

// UpdateEmployeeDetails actualiza detalles del empleado (Salario, Info Legal)
func UpdateEmployeeDetails(c *gin.Context) {
	id := c.Param("id")
	userID, _ := strconv.Atoi(id)

	var input struct {
		BaseSalary  *int64  `json:"base_salary"` // Pointer to check nil if not updating
		HasSecurity *bool   `json:"has_security"`
		FullName    *string `json:"full_name"`
		Cedula      *string `json:"cedula"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Update Payroll Info (Base Salary & Security)
	if input.BaseSalary != nil || input.HasSecurity != nil {
		var p models.UserPayroll
		if err := DB.First(&p, userID).Error; err != nil {
			p.UserID = uint(userID)
		}
		if input.BaseSalary != nil {
			p.BaseSalary = *input.BaseSalary
		}
		if input.HasSecurity != nil {
			p.HasSecurity = input.HasSecurity
		}
		p.UpdatedAt = time.Now()
		DB.Save(&p)
	}

	// 2. Update Personal Info (User)
	if input.FullName != nil || input.Cedula != nil {
		var u models.User
		if err := DB.First(&u, userID).Error; err == nil {
			if input.FullName != nil {
				u.FullName = *input.FullName
			}
			if input.Cedula != nil {
				u.Cedula = *input.Cedula
			}
			DB.Save(&u)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// --- PAYMENTS ---

// GeneratePayment calcula y guarda un pago de nómina
func GeneratePayment(c *gin.Context) {
	var input struct {
		UserID           uint      `json:"user_id" binding:"required"`
		PeriodStart      time.Time `json:"period_start"`
		PeriodEnd        time.Time `json:"period_end"`
		SundaysQty       int       `json:"sundays_qty"`
		MadrugonesQty    float64   `json:"madrugones_qty"`
		Advance          int64     `json:"advance"`
		IncludesSecurity bool      `json:"includes_security"`
		Aditions         string    `json:"aditions"`   // JSON string
		Deductions       string    `json:"deductions"` // JSON string
		Notes            string    `json:"notes"`
		CreatedBy        string    `json:"created_by"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Obtener Config Global
	var global models.NominaConfig
	if err := DB.First(&global, 1).Error; err != nil {
		// Defaults si falla config
		global.AuxilioTransporte = 0
		global.ValorDominical = 0
		global.ValorMadrugon = 10000
	}
	if global.ValorMadrugon == 0 {
		global.ValorMadrugon = 10000
	}

	// 2. Obtener Info Empleado
	var payroll models.UserPayroll
	if err := DB.First(&payroll, input.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El empleado no tiene salario base configurado"})
		return
	}

	// 3. Cálculos
	// Salario Base Quincenal = Base / 2
	paidBase := payroll.BaseSalary / 2

	// Auxilio Transporte Quincenal = Auxilio / 2
	transport := global.AuxilioTransporte / 2

	// Defaults if missing
	if global.PorcentajeSalud == 0 {
		global.PorcentajeSalud = 4.0
	}
	if global.PorcentajePension == 0 {
		global.PorcentajePension = 4.0
	}

	// Salud and Pension (Dynamic %)
	var health, pension int64
	if input.IncludesSecurity {
		health = int64(float64(paidBase) * (global.PorcentajeSalud / 100))
		pension = int64(float64(paidBase) * (global.PorcentajePension / 100))
	} else {
		health = 0
		pension = 0
	}

	// Dominicales
	sundaysTotal := int64(input.SundaysQty) * global.ValorDominical

	// Madrugones
	madrugonesTotal := int64(input.MadrugonesQty * float64(global.ValorMadrugon))

	// Total
	// Base + Transporte + Dominicales + Madrugones - Salud - Pension - Adelantos
	totalPaid := paidBase + transport + sundaysTotal + madrugonesTotal - health - pension - input.Advance

	// 4. Guardar Pago
	payment := models.NominaPayment{
		UserID:           input.UserID,
		PeriodStart:      input.PeriodStart,
		PeriodEnd:        input.PeriodEnd,
		BaseSalary:       payroll.BaseSalary,
		PaidBase:         paidBase,
		TransportAid:     transport,
		SundaysQty:       input.SundaysQty,
		SundaysTotal:     sundaysTotal,
		MadrugonesQty:    input.MadrugonesQty,
		MadrugonesTotal:  madrugonesTotal,
		IncludesSecurity: input.IncludesSecurity,
		Health:           health,
		Pension:          pension,
		Advance:          input.Advance,
		Aditions:         input.Aditions,
		Deductions:       input.Deductions,
		TotalPaid:        totalPaid,
		Notes:            input.Notes,
		CreatedBy:        input.CreatedBy,
		CreatedAt:        time.Now(),
	}

	if err := DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error guardando pago"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// UploadSignedContract sube un contrato firmado
func UploadSignedContract(c *gin.Context) {
	id := c.Param("id")

	// Validar que existe el pago
	var payment models.NominaPayment
	if err := DB.First(&payment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pago no encontrado"})
		return
	}

	// Recibir archivo
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Archivo no enviado"})
		return
	}

	// Guardar en disco
	// Asegurar que existe uploads/
	// (Debe crearse al inicio o aqui mismo)

	// Nombre único: payment_ID_timestamp.pdf
	filename := fmt.Sprintf("payment_%s_%d.pdf", id, time.Now().Unix())
	dst := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error guardando archivo"})
		return
	}

	// Actualizar BD
	payment.IsSigned = true
	payment.SignedFile = "/uploads/" + filename
	DB.Save(&payment)

	c.JSON(http.StatusOK, payment)
}

// DeleteNominaPayment elimina un pago de nómina
func DeleteNominaPayment(c *gin.Context) {
	id := c.Param("id")
	if err := DB.Delete(&models.NominaPayment{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando pago"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pago eliminado"})
}

// --- ODOO INTEGRATION ---

// GetOdooPOSConfigs lista los POS disponibles
func GetOdooPOSConfigs(c *gin.Context) {
	client, err := atmOdoo.NewFromEnv()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := client.Authenticate(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Odoo Auth: " + err.Error()})
		return
	}
	configs, err := client.ListPOSConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, configs)
}

// GetOdooSessions obtiene sesiones filtradas
func GetOdooSessions(c *gin.Context) {
	posIDStr := c.Query("pos_id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	posID, _ := strconv.Atoi(posIDStr)
	start, err1 := time.Parse(time.RFC3339, startStr)
	end, err2 := time.Parse(time.RFC3339, endStr)

	if posID == 0 || err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
		return
	}

	client, err := atmOdoo.NewFromEnv()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := client.Authenticate(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Odoo Auth: " + err.Error()})
		return
	}

	sessions, err := client.GetPOSSessions(posID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

// GetNominaHistory obtiene historial de pagos
func GetNominaHistory(c *gin.Context) {
	var history []models.NominaPayment
	DB.Preload("User").Order("created_at desc").Limit(50).Find(&history)
	c.JSON(http.StatusOK, history)
}

// --- MATRIX REPORT ---

// NominaMatrixDTO estructura de respuesta para la matriz
type NominaMatrixDTO struct {
	Users    interface{}              `json:"users"` // []EmployeePayrollDTO
	Payments []models.NominaPayment   `json:"payments"`
	Stats    map[string]map[int]int64 `json:"stats"` // [month][period] -> total_paid
}

// GetNominaMatrix retorna datos para el grid anual de pagos
func GetNominaMatrix(c *gin.Context) {
	yearStr := c.Query("year")
	year, _ := strconv.Atoi(yearStr)
	if year == 0 {
		year = time.Now().Year()
	}

	// 1. Obtener empleados activos
	// 1. Obtener empleados activos con su info de nómina
	type EmployeePayrollDTO struct {
		models.User
		Payroll *models.UserPayroll `json:"payroll"`
	}
	var users []models.User
	if err := DB.Order("name asc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cargando empleados"})
		return
	}

	// Batch fetch payrolls
	var payrolls []models.UserPayroll
	DB.Find(&payrolls)
	payrollMap := make(map[uint]*models.UserPayroll)
	for i := range payrolls {
		payrollMap[payrolls[i].UserID] = &payrolls[i]
	}

	var userDTOs []EmployeePayrollDTO
	for _, u := range users {
		userDTOs = append(userDTOs, EmployeePayrollDTO{
			User:    u,
			Payroll: payrollMap[u.ID], // Will be nil if not found, which is fine
		})
	}

	// 2. Obtener pagos del año
	startYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	var payments []models.NominaPayment
	if err := DB.Preload("User").
		Where("period_start >= ? AND period_start < ?", startYear, endYear).
		Order("period_start asc").
		Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cargando pagos"})
		return
	}

	// 3. Generar estadísticas simples para el grid (Frontend puede calcular detalle)
	// Pero ayudamos con totales por quincena
	stats := make(map[string]map[int]int64)

	for _, p := range payments {
		month := p.PeriodStart.Month().String() // "January", etc
		day := p.PeriodStart.Day()
		period := 1
		if day > 15 {
			period = 2
		}

		if stats[month] == nil {
			stats[month] = make(map[int]int64)
		}
		stats[month][period] += p.TotalPaid
	}

	c.JSON(http.StatusOK, NominaMatrixDTO{
		Users:    userDTOs,
		Payments: payments,
		Stats:    stats,
	})
}
