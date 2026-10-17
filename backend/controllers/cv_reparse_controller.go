package controllers

import (
	"log"
	"net/http"
	"sync"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// reparsing tracks companies with a background reparse running so repeat clicks don't stack jobs.
var reparsing sync.Map

// ReparseAllCVs extracts text in the background for every application still missing it.
func ReparseAllCVs(c *gin.Context) {
	company := companyID(c)
	if _, running := reparsing.LoadOrStore(company, true); running {
		c.JSON(http.StatusConflict, gin.H{"error": "CV parsing is already running for your company"})
		return
	}
	started := false
	defer func() {
		if !started {
			reparsing.Delete(company)
		}
	}()

	var applications []models.Application
	err := config.DB.Select("id", "resume_url", "resume_path").
		Where("company_id = ? AND (parsed_cv_text IS NULL OR parsed_cv_text = '') AND (resume_url <> '' OR resume_path IS NOT NULL)", company).
		Find(&applications).Error
	if err != nil {
		serverError(c, "Failed to fetch applications")
		return
	}
	if len(applications) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No applications need CV parsing", "total": 0})
		return
	}

	started = true
	go func(company uuid.UUID) {
		defer reparsing.Delete(company)
		failed := 0
		for i := range applications {
			app := &applications[i]
			text, err := services.ExtractApplicationCVText(app)
			if err == nil {
				err = services.SaveAnalysis(app.ID, text, nil)
			}
			if err != nil {
				failed++
				log.Printf("reparse application %s: %v", app.ID, err)
			}
		}
		log.Printf("CV reparse finished: %d of %d failed", failed, len(applications))
	}(company)

	c.JSON(http.StatusAccepted, gin.H{"message": "CV reparsing started in background", "total": len(applications)})
}

func ReparseSingleCV(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}

	text, err := services.ExtractApplicationCVText(app)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to parse CV", "details": err.Error()})
		return
	}
	if err := services.SaveAnalysis(app.ID, text, nil); err != nil {
		serverError(c, "Failed to save parsed CV text")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "CV parsed successfully", "characters": len(text)})
}
