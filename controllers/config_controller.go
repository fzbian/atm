package controllers

import (
	"atm/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterConfigRoutes(r *gin.RouterGroup) {
	cfg := r.Group("/config")
	{
		cfg.GET("/roles", GetRoleConfigs)
		cfg.POST("/roles", SaveRoleConfig)
	}
}

func GetRoleConfigs(c *gin.Context) {
	var configs []models.RoleConfig
	if err := DB.Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error loading configs"})
		return
	}

	// Ensure defaults exist if empty?
	// Or frontend handles defaults. Let's return what we have.
	c.JSON(http.StatusOK, configs)
}

func SaveRoleConfig(c *gin.Context) {
	var payload models.RoleConfig
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if payload.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role required"})
		return
	}

	// Update or Create
	if err := DB.Save(&payload).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving config"})
		return
	}

	c.JSON(http.StatusOK, payload)
}
