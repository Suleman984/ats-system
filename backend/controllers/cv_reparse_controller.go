package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"fmt"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// ReparseAllCVs re-parses all CVs that don't have parsed_cv_text
// This is useful for fixing existing applications
func ReparseAllCVs(c *gin.Context) {
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}
	companyID, ok := companyIDVal.(string)
	if !ok || companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	// Get all applications without parsed_cv_text for this company
	var applications []models.Application
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("(jobs.company_id = ? OR applications.company_id = ?) AND (applications.parsed_cv_text IS NULL OR applications.parsed_cv_text = '') AND applications.resume_url != ''", companyID, companyID).
		Find(&applications).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	if len(applications) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No applications found that need CV parsing",
			"count":   0,
		})
		return
	}

	// Parse CVs in background
	go func() {
		successCount := 0
		failCount := 0

		for _, app := range applications {
			log.Printf("Reparsing CV for application %s (Email: %s, URL: %s)", 
				app.ID.String(), app.Email, app.ResumeURL)

			cvText, err := services.ExtractTextFromURL(app.ResumeURL)
			if err != nil {
				log.Printf("ERROR: Failed to parse CV for %s: %v", app.Email, err)
				failCount++
				continue
			}

			if len(cvText) < 50 {
				log.Printf("WARNING: CV text too short for %s (%d characters)", app.Email, len(cvText))
				failCount++
				continue
			}

			// Validate UTF-8 before saving
			if !utf8.ValidString(cvText) {
				log.Printf("ERROR: Extracted CV text is not valid UTF-8 for %s. Skipping.", app.Email)
				failCount++
				continue
			}
			
			// Update parsed_cv_text
			updateErr := config.DB.Model(&models.Application{}).
				Where("id = ?", app.ID).
				Update("parsed_cv_text", cvText).Error

			if updateErr != nil {
				log.Printf("ERROR: Failed to save parsed CV text for %s: %v", app.Email, updateErr)
				failCount++
			} else {
				log.Printf("SUCCESS: CV parsed for %s (%d characters)", app.Email, len(cvText))
				successCount++
			}
		}

		log.Printf("CV Reparsing Complete: %d succeeded, %d failed out of %d total", 
			successCount, failCount, len(applications))
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "CV reparsing started in background",
		"total":   len(applications),
		"note":    "Check server logs for progress. This may take a few minutes.",
	})
}

// ReparseSingleCV re-parses a single CV by application ID
func ReparseSingleCV(c *gin.Context) {
	applicationID := c.Param("id")
	companyIDVal, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Company ID not found in token"})
		return
	}
	companyID, ok := companyIDVal.(string)
	if !ok || companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var application models.Application
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (jobs.company_id = ? OR applications.company_id = ?)", applicationID, companyID, companyID).
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	if application.ResumeURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application has no resume URL"})
		return
	}

	// Parse CV
	log.Printf("Reparsing CV for application %s (URL: %s)", applicationID, application.ResumeURL)
	cvText, err := services.ExtractTextFromURL(application.ResumeURL)
	if err != nil {
		log.Printf("ERROR: Failed to parse CV: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse CV",
			"details": err.Error(),
		})
		return
	}

	if len(cvText) < 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "CV text too short",
			"details": fmt.Sprintf("Extracted only %d characters. CV might be an image or corrupted.", len(cvText)),
		})
		return
	}

	// Update parsed_cv_text
	updateErr := config.DB.Model(&models.Application{}).
		Where("id = ?", application.ID).
		Update("parsed_cv_text", cvText).Error

	if updateErr != nil {
		log.Printf("ERROR: Failed to save parsed CV text: %v", updateErr)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save parsed CV text",
			"details": updateErr.Error(),
		})
		return
	}

	log.Printf("SUCCESS: CV parsed for %s (%d characters)", application.Email, len(cvText))

	c.JSON(http.StatusOK, gin.H{
		"message":     "CV parsed successfully",
		"characters": len(cvText),
	})
}

