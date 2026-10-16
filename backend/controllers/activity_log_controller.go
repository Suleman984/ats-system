package controllers

import (
	"net/http"
	"time"

	"ats-backend/config"
	"ats-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func filterActivityLogs(c *gin.Context, query *gorm.DB) *gorm.DB {
	if v := c.Query("action_type"); v != "" {
		query = query.Where("action_type = ?", v)
	}
	if v := c.Query("entity_type"); v != "" {
		query = query.Where("entity_type = ?", v)
	}
	if from, err := time.Parse("2006-01-02", c.Query("date_from")); err == nil {
		query = query.Where("created_at >= ?", from)
	}
	if to, err := time.Parse("2006-01-02", c.Query("date_to")); err == nil {
		query = query.Where("created_at < ?", to.AddDate(0, 0, 1))
	}
	return query.Order("created_at desc")
}

func GetActivityLogs(c *gin.Context) {
	var logs []models.ActivityLog
	query := filterActivityLogs(c, config.DB.Preload("Admin").Where("company_id = ?", companyID(c)))
	if err := query.Limit(100).Find(&logs).Error; err != nil {
		serverError(c, "Failed to fetch activity logs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs, "count": len(logs)})
}

func GetSuperAdminActivityLogs(c *gin.Context) {
	query := config.DB.Preload("Admin").Preload("Company")
	if v := c.Query("company_id"); v != "" {
		query = query.Where("company_id = ?", v)
	}

	var logs []models.ActivityLog
	if err := filterActivityLogs(c, query).Limit(200).Find(&logs).Error; err != nil {
		serverError(c, "Failed to fetch activity logs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs, "count": len(logs)})
}
