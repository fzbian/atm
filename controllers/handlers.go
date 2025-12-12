package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"atm/models"
	"atm/notify"
	"atm/odoo"
)

// -------------------- CATEGORIAS --------------------
// GetCategorias godoc
// @Summary Listar categorias
// @Produce json
// @Success 200 {array} models.Categoria
// @Failure 500 {object} map[string]interface{}
// @Router /api/categorias [get]
func GetCategorias(c *gin.Context) {
	var categorias []models.Categoria
	if err := DB.Find(&categorias).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categorias)
}

// CreateCategoria godoc
// @Summary Crear categoria
// @Accept json
// @Produce json
// @Param categoria body models.CategoriaCreateInput true "Categoria"
// @Success 201 {object} models.Categoria
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/categorias [post]
func CreateCategoria(c *gin.Context) {
	var input models.CategoriaCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// validar tipo
	tipo := strings.ToUpper(strings.TrimSpace(input.Tipo))
	if tipo != "INGRESO" && tipo != "EGRESO" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tipo inválido: debe ser 'INGRESO' o 'EGRESO'"})
		return
	}
	categoria := models.Categoria{
		Nombre: input.Nombre,
		Tipo:   tipo,
	}
	if err := DB.Create(&categoria).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, categoria)
}

// GetCategoria godoc
// @Summary Obtener categoria por id
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} models.Categoria
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/categorias/{id} [get]
func GetCategoria(c *gin.Context) {
	id := c.Param("id")
	var categoria models.Categoria
	if err := DB.First(&categoria, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "categoria no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categoria)
}

// UpdateCategoria godoc
// @Summary Actualizar categoria (parcial)
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param categoria body models.CategoriaUpdateInput false "Campos a actualizar (parcial)"
// @Success 200 {object} models.Categoria
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/categorias/{id} [put]
func UpdateCategoria(c *gin.Context) {
	id := c.Param("id")
	var categoria models.Categoria
	if err := DB.First(&categoria, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "categoria no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var input models.CategoriaUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Nombre == nil && input.Tipo == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no hay campos para actualizar"})
		return
	}
	if input.Nombre != nil {
		categoria.Nombre = *input.Nombre
	}
	if input.Tipo != nil {
		tipo := strings.ToUpper(strings.TrimSpace(*input.Tipo))
		if tipo != "INGRESO" && tipo != "EGRESO" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tipo inválido: debe ser 'INGRESO' o 'EGRESO'"})
			return
		}
		categoria.Tipo = tipo
	}
	if err := DB.Save(&categoria).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categoria)
}

// DeleteCategoria godoc
// @Summary Eliminar categoria
// @Produce json
// @Param id path int true "ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/categorias/{id} [delete]
func DeleteCategoria(c *gin.Context) {
	id := c.Param("id")
	var categoria models.Categoria
	if err := DB.First(&categoria, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "categoria no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := DB.Delete(&categoria).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Notificación removida para categorías
	c.Status(http.StatusNoContent)
}

// -------------------- TRANSACCIONES --------------------
// GetTransacciones godoc
// @Summary Listar transacciones
// @Produce json
// @Param limit query int false "Número máximo de movimientos a devolver"
// @Param from query string false "Fecha inicio (RFC3339 o YYYY-MM-DD)"
// @Param to query string false "Fecha fin (RFC3339 o YYYY-MM-DD)"
// @Param tipo query string false "Tipo de movimiento: INGRESO|EGRESO"
// @Param caja_id query int false "Filtrar por ID de caja"
// @Param descripcion query string false "Buscar por descripcion (texto parcial)"
// @Success 200 {array} models.Transaccion
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transacciones [get]
func GetTransacciones(c *gin.Context) {
	// Parámetros opcionales
	limitStr := c.Query("limit")
	fromStr := c.Query("from")            // fecha inicio (inclusive) RFC3339 o YYYY-MM-DD
	toStr := c.Query("to")                // fecha fin (inclusive) RFC3339 o YYYY-MM-DD
	tipo := c.Query("tipo")               // INGRESO o EGRESO (filtrado por categoria)
	cajaIDStr := c.Query("caja_id")       // Filtro por caja
	descripcion := c.Query("descripcion") // Filtro por descripción
	usuario := c.Query("usuario")         // Nuevo filtro por usuario

	var transacciones []models.Transaccion
	query := DB.Model(&models.Transaccion{})

	// Si se filtra por tipo, hacemos JOIN con categorias
	if tipo != "" {
		query = query.Joins("JOIN categorias c ON c.id = transacciones.categoria_id").Where("c.tipo = ?", tipo)
	}

	// Filtros de descripcion
	if descripcion != "" {
		query = query.Where("transacciones.descripcion LIKE ?", "%"+descripcion+"%")
	}

	// Filtro por usuario
	if usuario != "" {
		query = query.Where("transacciones.usuario = ?", usuario)
	}
	// Filtro por caja
	if cajaIDStr != "" {
		if cajaID, err := strconv.Atoi(cajaIDStr); err == nil && cajaID > 0 {
			query = query.Where("transacciones.caja_id = ?", cajaID)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parametro 'caja_id' inválido"})
			return
		}
	}

	// Parsear y aplicar from/to
	timeLayouts := []string{time.RFC3339, "2006-01-02"}
	if fromStr != "" {
		var t time.Time
		var err error
		for _, l := range timeLayouts {
			t, err = time.Parse(l, fromStr)
			if err == nil {
				break
			}
		}
		if t.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "formato de 'from' inválido. Use RFC3339 o YYYY-MM-DD"})
			return
		}
		query = query.Where("transacciones.fecha >= ?", t)
	}
	if toStr != "" {
		var t time.Time
		var err error
		for _, l := range timeLayouts {
			t, err = time.Parse(l, toStr)
			if err == nil {
				break
			}
		}
		if t.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "formato de 'to' inválido. Use RFC3339 o YYYY-MM-DD"})
			return
		}
		// hacer la fecha inclusiva hasta el fin del día si se usó YYYY-MM-DD
		if len(toStr) == len("2006-01-02") {
			t = t.Add(24*time.Hour - time.Nanosecond)
		}
		query = query.Where("transacciones.fecha <= ?", t)
	}

	// Orden por fecha desc
	query = query.Order("transacciones.fecha desc")

	// Aplicar limit si viene
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parametro 'limit' inválido"})
			return
		}
		query = query.Limit(l)
	}

	if err := query.Find(&transacciones).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, transacciones)
}

// CreateTransaccion godoc
// @Summary Crear transaccion
// @Accept json
// @Produce json
// @Param transaccion body models.TransaccionCreateInput true "Transaccion (requiere caja_id)"
// @Success 201 {object} models.Transaccion
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transacciones [post]
func CreateTransaccion(c *gin.Context) {
	var input models.TransaccionCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Obtener tipo de la categoria
	var categoria models.Categoria
	if err := DB.First(&categoria, input.CategoriaID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "categoria no encontrada"})
		return
	}
	// Validar caja existente
	var caja models.Caja
	if err := DB.First(&caja, input.CajaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "caja no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Si es EGRESO, validar saldo suficiente en esa caja
	if categoria.Tipo == "EGRESO" {
		if caja.Saldo < input.Monto {
			c.JSON(http.StatusConflict, gin.H{"error": "Saldo insuficiente en la caja para realizar el egreso", "saldo_actual": caja.Saldo, "monto_solicitado": input.Monto, "caja_id": input.CajaID})
			return
		}
	}
	transaccion := models.Transaccion{
		CategoriaID: input.CategoriaID,
		CajaID:      input.CajaID,
		Monto:       input.Monto,
		Descripcion: input.Descripcion,
		Usuario:     input.Usuario,
	}
	if err := DB.Create(&transaccion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	emoji := tipoEmoji(categoria.Tipo)
	msg := fmt.Sprintf("*TRANSACCION CREADA*\n📪 *ID:* %d\n📄 *Descripcion:* %s\n📚 *Categoria:* %s\n🏷️ *Tipo de movimiento:* %s %s\n💲*Monto:* %s\n🧾 *Caja:* %s\n👤 *Usuario:* %s", transaccion.ID, transaccion.Descripcion, categoria.Nombre, categoria.Tipo, emoji, formatMonto(transaccion.Monto), labelCaja(transaccion.CajaID), transaccion.Usuario)
	notify.SendText(msg)
	c.JSON(http.StatusCreated, transaccion)
}

// GetTransaccion godoc
// @Summary Obtener transaccion por id
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} models.Transaccion
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transacciones/{id} [get]
func GetTransaccion(c *gin.Context) {
	id := c.Param("id")
	var t models.Transaccion
	if err := DB.First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaccion no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

// UpdateTransaccion godoc
// @Summary Actualizar transaccion parcialmente
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param transaccion body models.TransaccionUpdateInput false "Campos a actualizar (parcial). Debe incluir caja_id en body o query."
// @Param usuario query string true "Usuario que actualiza la transacción"
// @Param caja_id query int false "ID de caja (obligatorio si no se envía en el body)"
// @Success 200 {object} models.Transaccion
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transacciones/{id} [put]
func UpdateTransaccion(c *gin.Context) {
	id := c.Param("id")
	usuario := c.Query("usuario")
	if usuario == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parametro 'usuario' es requerido"})
		return
	}
	var t models.Transaccion
	if err := DB.First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaccion no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	oldCat := t.CategoriaID
	oldMonto := t.Monto
	oldDesc := t.Descripcion
	var input models.TransaccionUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// caja_id es obligatorio (ya sea en body o como query param)
	var cajaID uint
	if input.CajaID != nil {
		cajaID = *input.CajaID
	} else {
		if s := strings.TrimSpace(c.Query("caja_id")); s != "" {
			if v, err := strconv.Atoi(s); err == nil && v > 0 {
				cajaID = uint(v)
			}
		}
	}
	if cajaID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "caja_id es requerido para actualizar"})
		return
	}
	// Validar caja existente
	{
		var caja models.Caja
		if err := DB.First(&caja, cajaID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "caja no encontrada"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if input.CategoriaID == nil && input.Monto == nil && input.Descripcion == nil && input.CajaID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no hay campos para actualizar"})
		return
	}
	// Determinar nuevos valores propuestos
	newCatID := oldCat
	if input.CategoriaID != nil {
		newCatID = *input.CategoriaID
	}
	newMonto := oldMonto
	if input.Monto != nil {
		newMonto = *input.Monto
	}
	// Si nuevo tipo es EGRESO, validar saldo suficiente
	var newCat models.Categoria
	if err := DB.First(&newCat, newCatID).Error; err == nil {
		if strings.ToUpper(newCat.Tipo) == "EGRESO" {
			// Si cambia de caja, necesitamos saldo >= nuevo monto en la nueva caja
			if cajaID != t.CajaID {
				var nuevaCaja models.Caja
				if err := DB.First(&nuevaCaja, cajaID).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo validar saldo de caja", "detalle": err.Error()})
					return
				}
				if nuevaCaja.Saldo < newMonto {
					c.JSON(http.StatusConflict, gin.H{"error": "Saldo insuficiente en la caja destino para actualizar el egreso", "saldo_actual": nuevaCaja.Saldo, "monto_requerido": newMonto, "caja_id": cajaID})
					return
				}
			} else {
				// Misma caja: sólo validar delta positivo
				if newMonto > oldMonto {
					delta := newMonto - oldMonto
					var caja models.Caja
					if err := DB.First(&caja, cajaID).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo validar saldo de caja", "detalle": err.Error()})
						return
					}
					if caja.Saldo < delta {
						c.JSON(http.StatusConflict, gin.H{"error": "Saldo insuficiente en la caja para aumentar el egreso", "saldo_actual": caja.Saldo, "incremento": delta, "caja_id": cajaID})
						return
					}
				}
			}
		}
	}
	// Construir mapa de actualizaciones
	updates := map[string]interface{}{}
	if input.CategoriaID != nil {
		updates["categoria_id"] = *input.CategoriaID
	}
	if cajaID != t.CajaID {
		updates["caja_id"] = cajaID
	}
	if input.Monto != nil {
		updates["monto"] = *input.Monto
	}
	if input.Descripcion != nil {
		updates["descripcion"] = *input.Descripcion
	}
	log.Printf("[DEBUG] UpdateTransaccion updates map: %v", updates)
	res := DB.Debug().Model(&t).Updates(updates)
	if res.Error != nil {
		log.Printf("[DEBUG] UpdateTransaccion error: %v", res.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	log.Printf("[DEBUG] UpdateTransaccion RowsAffected=%d", res.RowsAffected)
	// recargar la transacción actualizada
	if err := DB.First(&t, t.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo recargar transaccion", "detalle": err.Error()})
		return
	}
	// Actualizar el usuario en el último log de UPDATE generado por el trigger
	var lastLogID int64
	err := DB.Raw("SELECT id FROM transacciones_log WHERE transaccion_id = ? AND accion = 'UPDATE' ORDER BY id DESC LIMIT 1", t.ID).Scan(&lastLogID).Error
	if err == nil && lastLogID > 0 {
		if execRes := DB.Exec("UPDATE transacciones_log SET usuario = ? WHERE id = ?", usuario, lastLogID); execRes.Error != nil {
			log.Printf("[DEBUG] UpdateTransaccion update log error: %v", execRes.Error)
		}
	}
	var lines []string
	if t.CategoriaID != oldCat {
		var cat models.Categoria
		DB.First(&cat, t.CategoriaID)
		emoji := tipoEmoji(cat.Tipo)
		lines = append(lines, fmt.Sprintf("📚 *Categoria:* %s", cat.Nombre))
		lines = append(lines, fmt.Sprintf("🏷️ *Tipo de movimiento:* %s %s", cat.Tipo, emoji))
	}
	if t.Monto != oldMonto {
		lines = append(lines, fmt.Sprintf("💲*Monto:* %s", formatMonto(t.Monto)))
	}
	if t.Descripcion != oldDesc {
		lines = append(lines, fmt.Sprintf("📄 *Descripcion:* %s", t.Descripcion))
	}
	if len(lines) > 0 {
		msg := fmt.Sprintf("*TRANSACCION ACTUALIZADA*\n📪 *ID:* %d\n%s\n🧾 *Caja:* %s\n👤 *Usuario:* %s", t.ID, strings.Join(lines, "\n"), labelCaja(cajaID), usuario)
		notify.SendText(msg)
	}
	c.JSON(http.StatusOK, t)
}

// DeleteTransaccion godoc
// @Summary Eliminar transaccion
// @Produce json
// @Param id path int true "ID"
// @Param usuario query string true "Usuario que elimina la transacción"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/transacciones/{id} [delete]
func DeleteTransaccion(c *gin.Context) {
	id := c.Param("id")
	usuario := c.Query("usuario")
	if usuario == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parametro 'usuario' es requerido"})
		return
	}
	var t models.Transaccion
	if err := DB.First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaccion no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Datos antes de borrar
	desc := t.Descripcion
	monto := t.Monto
	cajaID := t.CajaID
	// borrar
	if err := DB.Delete(&models.Transaccion{}, t.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Actualizar el usuario en el último log de DELETE generado por el trigger
	var lastLogID int64
	err2 := DB.Raw("SELECT id FROM transacciones_log WHERE transaccion_id = ? AND accion = 'DELETE' ORDER BY id DESC LIMIT 1", t.ID).Scan(&lastLogID).Error
	if err2 == nil && lastLogID > 0 {
		if execRes := DB.Exec("UPDATE transacciones_log SET usuario = ? WHERE id = ?", usuario, lastLogID); execRes.Error != nil {
			log.Printf("[DEBUG] DeleteTransaccion update log error: %v", execRes.Error)
		}
	}
	msg := fmt.Sprintf("*TRANSACCION ELIMINADA*\n📪 *ID:* %s\n📄 *Descripcion:* %s\n💲*Monto:* %s\n🧾 *Caja:* %s\n👤 *Usuario:* %s", id, desc, formatMonto(monto), labelCaja(cajaID), usuario)
	notify.SendText(msg)
	c.Status(http.StatusNoContent)
}

// -------------------- CAJA --------------------
// GetCaja godoc
// @Summary Obtener saldo en caja
// @Description Si solo_caja=true, devuelve solo los saldos locales (caja 1 y caja 2). Si no, devuelve saldos locales y saldos de Odoo (POS) combinados.
// @Produce json
// @Param solo_caja query bool false "Solo saldo de caja local (sin Odoo)"
// @Success 200 {object} models.CajaOdoo
// @Success 200 {object} models.Caja
// @Failure 500 {object} map[string]interface{}
// @Router /api/caja [get]
func GetCaja(c *gin.Context) {
	soloCaja := c.Query("solo_caja")
	var cajaLocal models.Caja
	if err := DB.First(&cajaLocal, 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "caja local no inicializada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error obteniendo caja local", "detalle": err.Error()})
		return
	}
	// Obtener también caja id 2 (cuenta bancaria) si existe
	var caja2 models.Caja
	if err := DB.First(&caja2, 2).Error; err != nil {
		// Si no existe, asumimos saldo 0 sin fallar toda la solicitud
		caja2 = models.Caja{ID: 2, Saldo: 0}
	}
	if soloCaja == "true" || soloCaja == "1" {
		c.JSON(http.StatusOK, gin.H{
			"id":                   cajaLocal.ID,
			"saldo_caja":           cajaLocal.Saldo,
			"saldo_caja2":          caja2.Saldo,
			"ultima_actualizacion": cajaLocal.UltimaActualizacion,
		})
		return
	}
	client, err := odoo.NewFromEnv()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuración Odoo incompleta", "detalle": err.Error()})
		return
	}
	if err = client.Authenticate(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo autenticar en Odoo", "detalle": err.Error()})
		return
	}
	locales, _, err := client.FetchPOSBalances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo obtener saldos filtrados de Odoo", "detalle": err.Error()})
		return
	}
	truncateToPeso := func(v float64) float64 {
		if v >= 0 {
			return math.Floor(v)
		}
		return math.Ceil(v)
	}
	localesEnteros := make(map[string]models.POSLocalDetail, len(locales))
	var totalLocalesEnteros float64
	for k, v := range locales {
		v.SaldoEnCaja = truncateToPeso(v.SaldoEnCaja)
		if v.Vendido != nil {
			val := truncateToPeso(*v.Vendido)
			v.Vendido = &val
		}
		localesEnteros[k] = v
		totalLocalesEnteros += v.SaldoEnCaja
	}
	resp := models.CajaOdoo{
		ID:                  1,
		SaldoCaja:           cajaLocal.Saldo,
		SaldoCaja2:          caja2.Saldo,
		Locales:             localesEnteros,
		TotalLocales:        totalLocalesEnteros,
		SaldoTotal:          cajaLocal.Saldo + caja2.Saldo + totalLocalesEnteros,
		UltimaActualizacion: time.Now(),
	}
	c.JSON(http.StatusOK, resp)
}

// -------------------- LOGS --------------------
// GetTransaccionesLog godoc
// @Summary Listar logs de transacciones
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/logs [get]
func GetTransaccionesLog(c *gin.Context) {
	var logs []map[string]interface{}
	if err := DB.Table("transacciones_log").Order("fecha desc").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// -------------------- RESUMEN --------------------
// GetResumenFinanciero godoc
// @Summary Obtener resumen financiero (vista resumen_financiero)
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/resumen [get]
func GetResumenFinanciero(c *gin.Context) {
	var resumen map[string]interface{}
	if err := DB.Raw("SELECT * FROM resumen_financiero").Scan(&resumen).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resumen)
}

// DeleteAllData godoc
// @Summary Eliminar todos los datos de la base de datos
// @Description Elimina todas las transacciones, logs, caja, y categorías (deja la estructura vacía)
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/limpiar [post]
func DeleteAllData(c *gin.Context) {
	errTx := DB.Exec("DELETE FROM transacciones").Error
	errLog := DB.Exec("DELETE FROM transacciones_log").Error
	errCaja := DB.Exec("UPDATE caja SET saldo = 0").Error
	errCat := DB.Exec("DELETE FROM categorias").Error

	if errTx != nil || errLog != nil || errCaja != nil || errCat != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         "Error al limpiar la base de datos",
			"transacciones": errTx,
			"logs":          errLog,
			"caja":          errCaja,
			"categorias":    errCat,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Base de datos limpiada correctamente"})
}

// Helpers de formato
func formatMonto(val float64) string {
	// Convertimos a enteros y decimales (2) para presentación estilo es_CO: miles con '.' y decimales con ','
	enteros := int64(val)
	dec := int64((val-float64(enteros))*100 + 0.0000001)
	intStr := strconv.FormatInt(enteros, 10)
	var grupos []string
	for len(intStr) > 3 {
		grupos = append([]string{intStr[len(intStr)-3:]}, grupos...)
		intStr = intStr[:len(intStr)-3]
	}
	if intStr != "" {
		grupos = append([]string{intStr}, grupos...)
	}
	res := strings.Join(grupos, ".")
	if dec > 0 {
		res = fmt.Sprintf("%s,%02d", res, dec)
	}
	return res
}

func tipoEmoji(tipo string) string {
	switch strings.ToUpper(tipo) {
	case "INGRESO":
		return "🟢"
	case "EGRESO":
		return "🔴"
	default:
		return "🏷️"
	}
}

// labelCaja devuelve una etiqueta amigable para la caja según su ID.
// 1 -> Efectivo, 2 -> Cuenta bancaria, cualquier otro -> "Caja <id>"
func labelCaja(id uint) string {
	switch id {
	case 1:
		return "Efectivo"
	case 2:
		return "Cuenta bancaria"
	default:
		return fmt.Sprintf("Caja %d", id)
	}
}

// -------------------- NOTIFY --------------------
// NotifyTest godoc
// @Summary Enviar mensaje de prueba de notificación
// @Description Envía un mensaje de prueba al chat fijo 'atm' usando NOTIFY_URL (nuevo formato /whatsapp/send-text). Por defecto envía "ping".
// @Produce json
// @Param text query string false "Texto a enviar"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/notify/test [get]
func NotifyTest(c *gin.Context) {
	text := c.Query("text")
	if strings.TrimSpace(text) == "" {
		text = "ping"
	}
	// Enviar y confirmar con 200 (la función interna ya hace logging de errores)
	notify.SendText(text)
	c.JSON(http.StatusOK, gin.H{"status": "sent", "text": text})
}

// -------------------- ODOO --------------------
type cashOutRequest struct {
	POSName      string  `json:"pos_name" binding:"required"`
	Amount       float64 `json:"amount" binding:"required"`
	CategoryName string  `json:"category_name" binding:"required"`
	Reason       string  `json:"reason" binding:"required"`
	Usuario      string  `json:"usuario"`
	CategoriaID  uint    `json:"categoria_id,omitempty"`
}

// OdooCashOut godoc
// @Summary Cash out en Odoo POS
// @Description Realiza una retirada de caja (cash out) en una sesión de POS abierta en Odoo. Usa credenciales de entorno ODOO_*.
// @Accept json
// @Produce json
// @Param payload body cashOutRequest true "Cash out request (usuario puede ir en body o como query param). Si categoria_id == 16 (Efectivos Puntos de Venta), crea ingreso interno en caja_id=1."
// @Param usuario query string false "Usuario que realiza el cashout (alternativo a body.usuario)"
// @Success 200 {object} odoo.CashOutResult
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/odoo/cashout [post]
func OdooCashOut(c *gin.Context) {
	var req cashOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount debe ser > 0"})
		return
	}
	// Validar categoría
	cat := strings.ToUpper(strings.TrimSpace(req.CategoryName))
	if cat != "GASTO" && cat != "RETIRADA" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category_name debe ser 'GASTO' o 'RETIRADA'"})
		return
	}
	// Validar razón no vacía
	if strings.TrimSpace(req.Reason) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason no puede estar en blanco"})
		return
	}
	// Determinar usuario (body o query)
	usuario := strings.TrimSpace(req.Usuario)
	if usuario == "" {
		usuario = strings.TrimSpace(c.Query("usuario"))
	}
	if usuario == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "usuario es requerido (en body.usuario o query ?usuario=)"})
		return
	}
	// Cargar env Odoo
	odooURL := strings.TrimSpace(os.Getenv("ODOO_URL"))
	db := strings.TrimSpace(os.Getenv("ODOO_DB"))
	user := strings.TrimSpace(os.Getenv("ODOO_USER"))
	pass := strings.TrimSpace(os.Getenv("ODOO_PASSWORD"))
	mysqlURI := strings.TrimSpace(os.Getenv("MYSQL_URI"))
	if odooURL == "" || db == "" || user == "" || pass == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "variables ODOO_* faltantes"})
		return
	}
	// Timeout de operación
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	// Validar existencia de POS; si no existe, devolver lista
	names, err := odoo.ListPOSNames(ctx, odooURL, db, user, pass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	found := false
	for _, n := range names {
		if n == req.POSName {
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pos_name no encontrado", "disponibles": names})
		return
	}
	res, err := odoo.CashOutPOS(ctx, odooURL, db, user, pass, req.POSName, req.Amount, req.Reason, cat, mysqlURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Notificaciones y transacción interna según reglas
	if res.OK {
		// Siempre enviar mensaje al chat "gastos" indicando la retirada en el POS
		gastoMsg := fmt.Sprintf("*RETIRADA EN CAJA*\n🏬 *PUNTO DE VENTA:* %s\n💲 *Monto:* %s\n📝 *Razón:* %s", req.POSName, formatMonto(req.Amount), req.Reason)
		notify.SendTo("retiradas", gastoMsg)

		// Si categoria_id == 16 (Efectivos Puntos de Venta), crear ingreso interno en caja_id=1
		if req.CategoriaID == 16 {
			const transferCatID uint = 16
			var catRec models.Categoria
			if err := DB.First(&catRec, transferCatID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "categoria de transferencia (ID 16) no existe", "detalle": err.Error()})
				return
			}
			desc := fmt.Sprintf("%s: %s", strings.ReplaceAll(req.POSName, "_", " "), req.Reason)
			t := models.Transaccion{
				CategoriaID: transferCatID,
				CajaID:      1,
				Monto:       req.Amount,
				Descripcion: desc,
				Usuario:     usuario,
			}
			if err := DB.Create(&t).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo crear transaccion interna", "detalle": err.Error()})
				return
			}
			// Notificar transacción al chat "atm" (SendText ya usa 'atm')
			emoji := tipoEmoji(catRec.Tipo)
			msg := fmt.Sprintf("*TRANSACCION CREADA*\n📪 *ID:* %d\n📄 *Descripcion:* %s\n📚 *Categoria:* %s\n🏷️ *Tipo de movimiento:* %s %s\n💲*Monto:* %s\n🧾 *Caja:* %s\n👤 *Usuario:* %s", t.ID, t.Descripcion, catRec.Nombre, catRec.Tipo, emoji, formatMonto(t.Monto), labelCaja(t.CajaID), t.Usuario)
			notify.SendText(msg)
		}
	}
	c.JSON(http.StatusOK, res)
}

// OdooListPOS godoc
// @Summary Listar nombres de POS en Odoo
// @Description Devuelve los nombres de los puntos de venta disponibles en Odoo usando las credenciales ODOO_* del entorno.
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} map[string]interface{}
// @Router /api/odoo/pos [get]
func OdooListPOS(c *gin.Context) {
	odooURL := strings.TrimSpace(os.Getenv("ODOO_URL"))
	db := strings.TrimSpace(os.Getenv("ODOO_DB"))
	user := strings.TrimSpace(os.Getenv("ODOO_USER"))
	pass := strings.TrimSpace(os.Getenv("ODOO_PASSWORD"))
	if odooURL == "" || db == "" || user == "" || pass == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "variables ODOO_* faltantes"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	names, err := odoo.ListPOSNames(ctx, odooURL, db, user, pass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, names)
}

// -------------------- CUENTA BANCARIA --------------------
type RetiroCuentaRequest struct {
	Monto       float64 `json:"monto" binding:"required"`
	Descripcion string  `json:"descripcion" binding:"required"`
	Usuario     string  `json:"usuario" binding:"required"`
}

// RetiroCuenta godoc
// @Summary Retiro en cuenta bancaria (caja_2) con ingreso automático en efectivo (caja_1)
// @Description Crea un EGRESO en caja_id=2 por el monto indicado y, en la misma operación, crea un INGRESO en caja_id=1 por el mismo monto. La operación es atómica.
// @Accept json
// @Produce json
// @Param payload body RetiroCuentaRequest true "Datos del retiro desde la cuenta bancaria (usa categorias fijas: egreso=30, ingreso=20)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/cuenta/retiro [post]
func RetiroCuenta(c *gin.Context) {
	var req RetiroCuentaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Monto <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "monto debe ser > 0"})
		return
	}
	// Categorías fijas: egreso=30, ingreso=20
	const egresoCatID uint = 30
	const ingresoCatID uint = 20
	var catEg models.Categoria
	if err := DB.First(&catEg, egresoCatID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "categoria de egreso (ID 30) no existe", "detalle": err.Error()})
		return
	}
	if strings.ToUpper(catEg.Tipo) != "EGRESO" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "categoria de egreso (ID 30) no es de tipo EGRESO"})
		return
	}
	var catIn models.Categoria
	if err := DB.First(&catIn, ingresoCatID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "categoria de ingreso (ID 20) no existe", "detalle": err.Error()})
		return
	}
	if strings.ToUpper(catIn.Tipo) != "INGRESO" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "categoria de ingreso (ID 20) no es de tipo INGRESO"})
		return
	}
	// Validar caja 2 y saldo suficiente
	var caja2 models.Caja
	if err := DB.First(&caja2, 2).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "caja_2 no existe"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if caja2.Saldo < req.Monto {
		c.JSON(http.StatusConflict, gin.H{"error": "Saldo insuficiente en la cuenta bancaria (caja_2)", "saldo_actual": caja2.Saldo, "monto_solicitado": req.Monto})
		return
	}
	// Transacción atómica
	var tEgreso, tIngreso models.Transaccion
	err := DB.Transaction(func(tx *gorm.DB) error {
		tEgreso = models.Transaccion{
			CategoriaID: egresoCatID,
			CajaID:      2,
			Monto:       req.Monto,
			Descripcion: req.Descripcion,
			Usuario:     req.Usuario,
		}
		if err := tx.Create(&tEgreso).Error; err != nil {
			return err
		}
		tIngreso = models.Transaccion{
			CategoriaID: ingresoCatID,
			CajaID:      1,
			Monto:       req.Monto,
			Descripcion: "Ingreso por retiro desde Cuenta bancaria",
			Usuario:     req.Usuario,
		}
		if err := tx.Create(&tIngreso).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo completar el retiro/ingreso", "detalle": err.Error()})
		return
	}
	// Notificaciones
	emojiEg := tipoEmoji(catEg.Tipo)
	msgEg := fmt.Sprintf("*TRANSACCION CREADA*\n📪 *ID:* %d\n📄 *Descripcion:* %s\n📚 *Categoria:* %s\n🏷️ *Tipo de movimiento:* %s %s\n💲*Monto:* %s\n🧾 *Caja:* %s\n👤 *Usuario:* %s", tEgreso.ID, tEgreso.Descripcion, catEg.Nombre, catEg.Tipo, emojiEg, formatMonto(tEgreso.Monto), labelCaja(tEgreso.CajaID), tEgreso.Usuario)
	notify.SendText(msgEg)
	emojiIn := tipoEmoji(catIn.Tipo)
	msgIn := fmt.Sprintf("*TRANSACCION CREADA*\n📪 *ID:* %d\n📄 *Descripcion:* %s\n📚 *Categoria:* %s\n🏷️ *Tipo de movimiento:* %s %s\n💲*Monto:* %s\n🧾 *Caja:* %s\n👤 *Usuario:* %s", tIngreso.ID, tIngreso.Descripcion, catIn.Nombre, catIn.Tipo, emojiIn, formatMonto(tIngreso.Monto), labelCaja(tIngreso.CajaID), tIngreso.Usuario)
	notify.SendText(msgIn)

	c.JSON(http.StatusOK, gin.H{
		"egreso":  tEgreso,
		"ingreso": tIngreso,
	})
}
