package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetActivityLogs returns activity logs for a company (admin view)
func GetActivityLogs(c *gin.Context) {
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}

	companyIDStr, ok := companyIDVal.(string)
	if !ok || companyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID format"})
		return
	}

	var logs []models.ActivityLog
	query := config.DB.Where("company_id = ?", companyID).
		Preload("Admin").
		Order("created_at DESC")

	// Filter by action type if provided
	if actionType := c.Query("action_type"); actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}

	// Filter by entity type if provided
	if entityType := c.Query("entity_type"); entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}

	// Filter by date range
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if dateFromTime, err := time.Parse("2006-01-02", dateFrom); err == nil {
			query = query.Where("DATE(created_at) >= ?", dateFromTime.Format("2006-01-02"))
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if dateToTime, err := time.Parse("2006-01-02", dateTo); err == nil {
			query = query.Where("DATE(created_at) <= ?", dateToTime.Format("2006-01-02"))
		}
	}

	// Pagination
	limit := 100 // Default limit
	query = query.Limit(limit)

	// Execute query
	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch activity logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"count": len(logs),
	})
}

// GetSuperAdminActivityLogs returns all activity logs across all companies (super admin view)
func GetSuperAdminActivityLogs(c *gin.Context) {
	var logs []models.ActivityLog
	query := config.DB.Preload("Admin").
		Preload("Company").
		Order("created_at DESC")

	// Filter by company if provided
	if companyID := c.Query("company_id"); companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	// Filter by action type if provided
	if actionType := c.Query("action_type"); actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}

	// Filter by entity type if provided
	if entityType := c.Query("entity_type"); entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}

	// Filter by date range
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if dateFromTime, err := time.Parse("2006-01-02", dateFrom); err == nil {
			query = query.Where("DATE(created_at) >= ?", dateFromTime.Format("2006-01-02"))
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if dateToTime, err := time.Parse("2006-01-02", dateTo); err == nil {
			query = query.Where("DATE(created_at) <= ?", dateToTime.Format("2006-01-02"))
		}
	}

	// Pagination
	limit := 200 // Default limit for super admin
	query = query.Limit(limit)

	// Execute query
	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch activity logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"count": len(logs),
	})
}

