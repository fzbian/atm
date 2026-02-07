package controllers

import (
	"atm/models"
	"atm/odoo"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type billingMonthlyEntry struct {
	PosName string  `json:"pos_name" binding:"required"`
	Nomina  float64 `json:"nomina"`
}

// SaveBillingMonthly upserts gastos y margen por local/mes.
func SaveBillingMonthly(c *gin.Context) {
	var body struct {
		Year    int                   `json:"year" binding:"required"`
		Month   int                   `json:"month" binding:"required,min=1,max=12"`
		Entries []billingMonthlyEntry `json:"entries" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	toSave := make([]models.BillingMonthly, 0, len(body.Entries))
	for _, e := range body.Entries {
		toSave = append(toSave, models.BillingMonthly{
			PosName:   e.PosName,
			Year:      body.Year,
			Month:     body.Month,
			Nomina:    e.Nomina,
			UpdatedAt: now,
		})
	}

	if err := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pos_name"}, {Name: "year"}, {Name: "month"}},
		DoUpdates: clause.AssignmentColumns([]string{"nomina", "updated_at"}),
	}).Create(&toSave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "saved": len(toSave)})
}

// GetBillingMonthly devuelve ventas Odoo + gastos/margen guardados para un mes/año.
func GetBillingMonthly(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Año inválido"})
		return
	}
	monthStr := c.Query("month")
	var month int
	if monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Mes inválido"})
			return
		}
	}

	// Ventas y margen por local/mes desde Odoo
	ventas, margenOdoo, err := odoo.GetMonthlyBillingWithMargin(context.Background(),
		os.Getenv("ODOO_URL"), os.Getenv("ODOO_DB"), os.Getenv("ODOO_USER"), os.Getenv("ODOO_PASSWORD"), year)
	if err != nil {
		// No bloqueamos, devolvemos error pero continuamos con DB para que el front decida.
		fmt.Printf("[billing] error obteniendo ventas/margen Odoo: %v\n", err)
	}

	// Gastos/margen guardados
	var rows []models.BillingMonthly
	tx := DB.Where("year = ?", year)
	if month > 0 {
		tx = tx.Where("month = ?", month)
	}
	if err := tx.Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Unir datos
	type respEntry struct {
		PosName       string  `json:"pos_name"`
		Year          int     `json:"year"`
		Month         int     `json:"month"`
		Venta         float64 `json:"venta"`
		Margen        float64 `json:"margen"`
		GastosComunes float64 `json:"gastos_comunes"`
		Servicios     float64 `json:"servicios"`
		Nomina        float64 `json:"nomina"`
		Arriendo      float64 `json:"arriendo"`
		UtilidadBruta float64 `json:"utilidad_bruta"`
		ComisionAdmin float64 `json:"comision_admin"`
		UtilidadNeta  float64 `json:"utilidad_neta"`
	}

	entries := []respEntry{}

	// Helper to pick venta by month number
	getVenta := func(pos string, m int) float64 {
		if ventas == nil {
			return 0
		}
		for monthName, val := range ventas[pos] {
			if monthNumberFromLabel(monthName) == m {
				return val
			}
		}
		return 0
	}

	getMargen := func(pos string, m int) float64 {
		if margenOdoo == nil {
			return 0
		}
		for monthName, val := range margenOdoo[pos] {
			if monthNumberFromLabel(monthName) == m {
				return val
			}
		}
		return 0
	}

	// Config fijos (servicios/arriendo)
	var cfgs []models.BillingConfig
	if err := DB.Find(&cfgs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cfgMap := make(map[string]models.BillingConfig)
	for _, cfg := range cfgs {
		cfgMap[cfg.PosName] = cfg
	}

	// Gastos comunes desde gastos_locales (suma mensual)
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	type gastoSum struct {
		Local string
		Month int
		Total float64
	}
	var sums []gastoSum
	gastosMap := make(map[string]map[int]float64) // pos -> month -> total
	if err := DB.Model(&models.GastoLocal{}).
		Select("local as local, MONTH(fecha) as month, SUM(monto) as total").
		Where("fecha >= ? AND fecha < ?", start, end).
		Group("local, MONTH(fecha)").Scan(&sums).Error; err == nil {
		for _, s := range sums {
			if gastosMap[s.Local] == nil {
				gastosMap[s.Local] = make(map[int]float64)
			}
			gastosMap[s.Local][s.Month] = s.Total
		}
	}

	// index existing rows
	rowKey := func(pos string, y, m int) string { return fmt.Sprintf("%s-%d-%d", pos, y, m) }
	rowMap := make(map[string]models.BillingMonthly)
	for _, r := range rows {
		rowMap[rowKey(r.PosName, r.Year, r.Month)] = r
	}

	// gather all pos/month present either in ventas or rows
	posMonths := make(map[string]struct{})
	for pos, months := range ventas {
		for label := range months {
			mn := monthNumberFromLabel(label)
			if month > 0 && mn != month {
				continue
			}
			posMonths[rowKey(pos, year, mn)] = struct{}{}
		}
	}
	for _, r := range rows {
		if month > 0 && r.Month != month {
			continue
		}
		posMonths[rowKey(r.PosName, r.Year, r.Month)] = struct{}{}
	}

	for key := range posMonths {
		parts := strings.Split(key, "-")
		if len(parts) != 3 {
			continue
		}
		pos := parts[0]
		m, _ := strconv.Atoi(parts[2])

		venta := getVenta(pos, m)
		row := rowMap[key]
		rowMargen := row.Margen
		if rowMargen == 0 {
			rowMargen = getMargen(pos, m)
		}
		cfg := cfgMap[pos]
		serviciosTot := 0.0
		if cfg.InternetAplica {
			serviciosTot += cfg.Internet
		}
		if cfg.LuzAplica {
			serviciosTot += cfg.Luz
		}
		if cfg.GasAplica {
			serviciosTot += cfg.Gas
		}
		if cfg.AguaAplica {
			serviciosTot += cfg.Agua
		}
		arriendo := cfg.Arriendo
		gastosComunes := gastosMap[pos][m]
		gastosTot := gastosComunes + serviciosTot + row.Nomina + arriendo
		utilidadBruta := rowMargen - gastosTot
		comision := 0.05 * utilidadBruta
		if comision < 0 {
			comision = 0
		}
		utilidadNeta := utilidadBruta - comision

		entries = append(entries, respEntry{
			PosName:       pos,
			Year:          year,
			Month:         m,
			Venta:         venta,
			Margen:        rowMargen,
			GastosComunes: gastosComunes,
			Servicios:     serviciosTot,
			Nomina:        row.Nomina,
			Arriendo:      arriendo,
			UtilidadBruta: utilidadBruta,
			ComisionAdmin: comision,
			UtilidadNeta:  utilidadNeta,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"year":   year,
		"month":  month,
		"data":   entries,
		"source": "db+odoo",
	})
}

// GetBillingGastos lista gastos_locales filtrados por local/mes.
func GetBillingGastos(c *gin.Context) {
	local := c.Query("pos")
	if local == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pos requerido"})
		return
	}
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Año inválido"})
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mes inválido"})
		return
	}

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	var gastos []models.GastoLocal
	if err := DB.Where("local = ? AND fecha >= ? AND fecha < ?", local, start, end).Order("fecha asc").Find(&gastos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gastos)
}

// CreateBillingGasto crea un gasto común sin imagen para un local/mes.
func CreateBillingGasto(c *gin.Context) {
	var body struct {
		Pos    string  `json:"pos" binding:"required"`
		Year   int     `json:"year" binding:"required"`
		Month  int     `json:"month" binding:"required,min=1,max=12"`
		Motivo string  `json:"motivo" binding:"required"`
		Monto  float64 `json:"monto" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fecha := time.Date(body.Year, time.Month(body.Month), 1, 0, 0, 0, 0, time.UTC)
	gasto := models.GastoLocal{
		Local:   body.Pos,
		Fecha:   fecha,
		Tipo:    "GASTO_COMUN",
		Motivo:  body.Motivo,
		Monto:   body.Monto,
		Usuario: "sistema",
	}
	if err := DB.Create(&gasto).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gasto)
}

// monthNumberFromLabel acepta "January", "Enero", "January 2024", etc.
func monthNumberFromLabel(label string) int {
	l := strings.ToLower(label)
	names := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
	namesEn := []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}
	for i, n := range names {
		if strings.HasPrefix(l, n) {
			return i + 1
		}
	}
	for i, n := range namesEn {
		if strings.HasPrefix(l, n) {
			return i + 1
		}
	}
	return 0
}
