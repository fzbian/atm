package controllers

import (
	"fmt"
	"log"
	"net/http"
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
// @Param transaccion body models.TransaccionCreateInput true "Transaccion"
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
	transaccion := models.Transaccion{
		CategoriaID: input.CategoriaID,
		Monto:       input.Monto,
		Descripcion: input.Descripcion,
		Usuario:     input.Usuario,
	}
	if err := DB.Create(&transaccion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Obtener datos de la categoria
	var categoria models.Categoria
	DB.First(&categoria, transaccion.CategoriaID)
	emoji := tipoEmoji(categoria.Tipo)
	msg := fmt.Sprintf("*TRANSACCION CREADA*\n📪 *ID:* %d\n📄 *Descripcion:* %s\n📚 *Categoria:* %s\n🏷️ *Tipo de movimiento:* %s %s\n💲*Monto:* %s", transaccion.ID, transaccion.Descripcion, categoria.Nombre, categoria.Tipo, emoji, formatMonto(transaccion.Monto))
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
// @Param transaccion body models.TransaccionUpdateInput false "Campos a actualizar (parcial)"
// @Param usuario query string true "Usuario que actualiza la transacción"
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
	if input.CategoriaID == nil && input.Monto == nil && input.Descripcion == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no hay campos para actualizar"})
		return
	}
	updates := map[string]interface{}{}
	if input.CategoriaID != nil {
		updates["categoria_id"] = *input.CategoriaID
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
		msg := fmt.Sprintf("*TRANSACCION ACTUALIZADA*\n📪 *ID:* %d\n%s\n👤 *Usuario:* %s", t.ID, strings.Join(lines, "\n"), usuario)
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
	msg := fmt.Sprintf("*TRANSACCION ELIMINADA*\n📪 *ID:* %s\n📄 *Descripcion:* %s\n💲*Monto:* %s\n👤 *Usuario:* %s", id, desc, formatMonto(monto), usuario)
	notify.SendText(msg)
	c.Status(http.StatusNoContent)
}

// -------------------- CAJA --------------------
// GetCaja godoc
// @Summary Obtener saldo en caja (desde Odoo)
// @Produce json
// @Success 200 {object} models.CajaOdoo
// @Failure 500 {object} map[string]interface{}
// @Router /api/caja [get]
func GetCaja(c *gin.Context) {
	var cajaLocal models.Caja
	if err := DB.First(&cajaLocal, 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "caja local no inicializada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error obteniendo caja local", "detalle": err.Error()})
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
	locales, totalLocales, err := client.FetchPOSBalances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo obtener saldos filtrados de Odoo", "detalle": err.Error()})
		return
	}
	resp := models.CajaOdoo{
		ID:                  1,
		SaldoCaja:           cajaLocal.Saldo,
		Locales:             locales,
		TotalLocales:        totalLocales,
		SaldoTotal:          cajaLocal.Saldo + totalLocales,
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
