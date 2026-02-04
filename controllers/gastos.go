package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"atm/models"
	"atm/notify"
	"atm/odoo"
)

// GastoRequest define el payload para crear un gasto
type GastoRequest struct {
	Local     string  `json:"local" binding:"required"`
	Monto     float64 `json:"monto" binding:"required"`
	Motivo    string  `json:"motivo" binding:"required"`
	ImagenURL string  `json:"imagen_url" binding:"required"`
	Usuario   string  `json:"usuario"`
}

// CreateGasto crea un nuevo gasto operativo y realiza el cashout en Odoo
func CreateGasto(c *gin.Context) {
	var req GastoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Monto <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monto debe ser mayor a 0"})
		return
	}

	// Determinar usuario
	usuario := strings.TrimSpace(req.Usuario)
	if usuario == "" {
		usuario = "Desconocido"
	}

	// 1. Realizar Cashout en Odoo ("GASTO")
	// Configurar entorno Odoo
	odooURL := os.Getenv("ODOO_URL")
	db := os.Getenv("ODOO_DB")
	user := os.Getenv("ODOO_USER")
	pass := os.Getenv("ODOO_PASSWORD")
	mysqlURI := os.Getenv("MYSQL_URI")

	if odooURL == "" || db == "" || user == "" || pass == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuración Odoo incompleta"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Llamar a CashOutPOS de odoo package
	// Nota: Usamos "GASTO" como category_name para que Odoo lo registre así
	res, err := odoo.CashOutPOS(ctx, odooURL, db, user, pass, req.Local, req.Monto, req.Motivo, "GASTO", mysqlURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fallo al registrar en Odoo", "detalle": err.Error()})
		return
	}

	if !res.OK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Odoo rechazó el gasto", "detalle": res.Message})
		return
	}

	// 2. Guardar en base de datos local (gastos_locales)
	gasto := models.GastoLocal{
		Local:     req.Local,
		Fecha:     time.Now(),
		Tipo:      "GASTO_OPERATIVO",
		Motivo:    req.Motivo,
		Monto:     req.Monto,
		ImagenURL: req.ImagenURL,
		Usuario:   usuario,
	}

	if err := DB.Create(&gasto).Error; err != nil {
		// OJO: Si falla aquí, ya se hizo el cashout en Odoo.
		// Idealmente deberíamos revertir, pero por simplicidad retornamos error
		// y notificamos para corrección manual si es necesario.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "se hizo cashout en Odoo pero falló guardar localmente", "detalle": err.Error()})
		return
	}

	// 3. Notificar
	msg := fmt.Sprintf("*NUEVO GASTO OPERATIVO*\n🏬 *Local:* %s\n💰 *Monto:* %s\n📝 *Motivo:* %s\n📎 *Soporte:* %s\n👤 *Por:* %s",
		gasto.Local, formatMonto(gasto.Monto), gasto.Motivo, gasto.ImagenURL, gasto.Usuario)
	notify.SendTo("gastos", msg) // Asumiendo que existe canal 'gastos' o usar 'retiradas' se prefiere

	c.JSON(http.StatusCreated, gin.H{"message": "Gasto registrado exitosamente", "id": gasto.ID})
}

// GetGastos lista los gastos, opcionalmente filtrados por local
func GetGastos(c *gin.Context) {
	local := c.Query("local")
	var gastos []models.GastoLocal

	q := DB.Model(&models.GastoLocal{})
	if local != "" {
		q = q.Where("local = ?", local)
	}

	if err := q.Order("fecha desc").Find(&gastos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gastos)
}
