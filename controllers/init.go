package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

// SetDB asigna la conexión de base de datos para los controladores
func SetDB(db *gorm.DB) {
	DB = db
}

// RegisterCategoriaRoutes registra rutas de categorias
func RegisterCategoriaRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/categorias")
	{
		r.GET("", GetCategorias)
		r.POST("", CreateCategoria)
		r.GET(":id", GetCategoria)
		r.PUT(":id", UpdateCategoria)
		r.DELETE(":id", DeleteCategoria)
	}
}

// RegisterTransaccionRoutes registra rutas de transacciones
func RegisterTransaccionRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/transacciones")
	{
		r.GET("", GetTransacciones)
		r.POST("", CreateTransaccion)
		r.GET(":id", GetTransaccion)
		r.PUT(":id", UpdateTransaccion)
		r.DELETE(":id", DeleteTransaccion)
	}
}

// RegisterCajaRoutes registra rutas de caja
func RegisterCajaRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/caja")
	{
		r.GET("", GetCaja)
	}
}

// RegisterLogsRoutes registra rutas de logs
func RegisterLogsRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/logs")
	{
		r.GET("", GetTransaccionesLog)
	}
}

// RegisterResumenRoutes registra ruta de resumen
func RegisterResumenRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/resumen")
	{
		r.GET("", GetResumenFinanciero)
	}
}
