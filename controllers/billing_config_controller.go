package controllers

import (
	"atm/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type billingConfigEntry struct {
	PosName        string  `json:"pos_name" binding:"required"`
	Arriendo       float64 `json:"arriendo"`
	Internet       float64 `json:"internet"`
	InternetAplica bool    `json:"internet_aplica"`
	Luz            float64 `json:"luz"`
	LuzAplica      bool    `json:"luz_aplica"`
	Gas            float64 `json:"gas"`
	GasAplica      bool    `json:"gas_aplica"`
	Agua           float64 `json:"agua"`
	AguaAplica     bool    `json:"agua_aplica"`
}

func GetBillingConfigs(c *gin.Context) {
	var cfgs []models.BillingConfig
	if err := DB.Find(&cfgs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfgs)
}

func SaveBillingConfigs(c *gin.Context) {
	var body struct {
		Entries []billingConfigEntry `json:"entries" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	toSave := make([]models.BillingConfig, 0, len(body.Entries))
	for _, e := range body.Entries {
		toSave = append(toSave, models.BillingConfig{
			PosName:        e.PosName,
			Arriendo:       e.Arriendo,
			Internet:       e.Internet,
			InternetAplica: e.InternetAplica,
			Luz:            e.Luz,
			LuzAplica:      e.LuzAplica,
			Gas:            e.Gas,
			GasAplica:      e.GasAplica,
			Agua:           e.Agua,
			AguaAplica:     e.AguaAplica,
		})
	}

	if err := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pos_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"arriendo", "internet", "internet_aplica", "luz", "luz_aplica", "gas", "gas_aplica", "agua", "agua_aplica"}),
	}).Create(&toSave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "saved": len(toSave)})
}
