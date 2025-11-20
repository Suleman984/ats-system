package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AddCandidateNoteRequest for adding notes
type AddCandidateNoteRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
	Note          string `json:"note" binding:"required"`
	IsPrivate     bool   `json:"is_private"`
}

// UpdateCandidateNoteRequest for updating notes
type UpdateCandidateNoteRequest struct {
	Note      string `json:"note" binding:"required"`
	IsPrivate bool   `json:"is_private"`
}

// AddTalentPoolRequest for adding to talent pool
type AddTalentPoolRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
}

// UpdateReferralRequest for updating referral info
type UpdateReferralRequest struct {
	ReferralSource  string `json:"referral_source"`
	ReferredByName  string `json:"referred_by_name"`
	ReferredByEmail string `json:"referred_by_email"`
	ReferredByPhone string `json:"referred_by_phone"`
}

// AddCandidateNote adds a note to a candidate's application
func AddCandidateNote(c *gin.Context) {
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

	var req AddCandidateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify application belongs to company
	var application models.Application
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (applications.company_id = ? OR jobs.company_id = ?)", req.ApplicationID, companyID, companyID).
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Get admin ID
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)

	// Create note
	note := models.CandidateNote{
		ApplicationID: uuid.MustParse(req.ApplicationID),
		AdminID:       adminUUID,
		Note:          req.Note,
		IsPrivate:     req.IsPrivate,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := config.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add note"})
		return
	}

	// Load relations
	config.DB.Preload("Admin").Preload("Application").First(&note, note.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Note added successfully",
		"note":    note,
	})
}

// GetCandidateNotes retrieves all notes for an application
func GetCandidateNotes(c *gin.Context) {
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

	// Verify application belongs to company
	var application models.Application
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (applications.company_id = ? OR jobs.company_id = ?)", applicationID, companyID, companyID).
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Get admin ID to filter private notes
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)

	// Get notes (public notes + private notes created by this admin)
	var notes []models.CandidateNote
	query := config.DB.Where("application_id = ? AND (is_private = false OR admin_id = ?)", applicationID, adminUUID).
		Preload("Admin").
		Order("created_at DESC")

	if err := query.Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notes": notes,
		"count": len(notes),
	})
}

// UpdateCandidateNote updates a note
func UpdateCandidateNote(c *gin.Context) {
	noteID := c.Param("id")
	// Note: We verify ownership through admin_id, so company_id check is not needed here

	var req UpdateCandidateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get admin ID
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)

	// Find note and verify ownership
	var note models.CandidateNote
	if err := config.DB.Where("id = ? AND admin_id = ?", noteID, adminUUID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found or you don't have permission to edit it"})
		return
	}

	// Update note
	note.Note = req.Note
	note.IsPrivate = req.IsPrivate
	note.UpdatedAt = time.Now()

	if err := config.DB.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update note"})
		return
	}

	config.DB.Preload("Admin").Preload("Application").First(&note, note.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Note updated successfully",
		"note":    note,
	})
}

// DeleteCandidateNote deletes a note
func DeleteCandidateNote(c *gin.Context) {
	noteID := c.Param("id")

	// Get admin ID
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)

	// Find note and verify ownership
	var note models.CandidateNote
	if err := config.DB.Where("id = ? AND admin_id = ?", noteID, adminUUID).First(&note).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found or you don't have permission to delete it"})
		return
	}

	if err := config.DB.Delete(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

// AddToTalentPool adds a candidate to the talent pool
func AddToTalentPool(c *gin.Context) {
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

	var req AddTalentPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify application belongs to company
	var application models.Application
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (applications.company_id = ? OR jobs.company_id = ?)", req.ApplicationID, companyID, companyID).
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Get admin ID
	adminIDVal, _ := c.Get("admin_id")
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, _ := uuid.Parse(adminIDStr)

	now := time.Now()
	application.InTalentPool = true
	application.TalentPoolAddedAt = &now
	application.TalentPoolAddedBy = &adminUUID

	if err := config.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to talent pool"})
		return
	}

	// Log activity
	companyUUID, _ := uuid.Parse(companyID)
	applicationUUID, _ := uuid.Parse(req.ApplicationID)
	services.LogActivity(
		&companyUUID,
		&adminUUID,
		"candidate_added_to_talent_pool",
		"application",
		&applicationUUID,
		"Candidate added to talent pool: "+application.FullName,
		map[string]interface{}{
			"candidate_name": application.FullName,
			"candidate_email": application.Email,
		},
	)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Candidate added to talent pool",
		"application": application,
	})
}

// RemoveFromTalentPool removes a candidate from the talent pool
func RemoveFromTalentPool(c *gin.Context) {
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

	// Verify application belongs to company (even if job is deleted)
	var application models.Application
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (applications.company_id = ? OR jobs.company_id = ?)", applicationID, companyID, companyID).
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	application.InTalentPool = false
	application.TalentPoolAddedAt = nil
	application.TalentPoolAddedBy = nil

	if err := config.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from talent pool"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Candidate removed from talent pool",
		"application": application,
	})
}

// GetTalentPool retrieves all candidates in the talent pool
func GetTalentPool(c *gin.Context) {
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

	var applications []models.Application
	err := config.DB.Where("company_id = ? AND in_talent_pool = true", companyID).
		Preload("Job").
		Order("talent_pool_added_at DESC").
		Find(&applications).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch talent pool"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"applications": applications,
		"count":        len(applications),
	})
}

// UpdateReferralInfo updates referral information for an application
func UpdateReferralInfo(c *gin.Context) {
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

	var req UpdateReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify application belongs to company
	var application models.Application
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (applications.company_id = ? OR jobs.company_id = ?)", applicationID, companyID, companyID).
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Update referral info
	application.ReferralSource = req.ReferralSource
	application.ReferredByName = req.ReferredByName
	application.ReferredByEmail = req.ReferredByEmail
	application.ReferredByPhone = req.ReferredByPhone

	if err := config.DB.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update referral info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Referral information updated",
		"application": application,
	})
}

// GetRelationshipTimeline retrieves all interactions with a candidate
func GetRelationshipTimeline(c *gin.Context) {
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

	// Verify application belongs to company
	var application models.Application
	err := config.DB.Table("applications").
		Select("applications.*").
		Joins("LEFT JOIN jobs ON jobs.id = applications.job_id").
		Where("applications.id = ? AND (applications.company_id = ? OR jobs.company_id = ?)", applicationID, companyID, companyID).
		First(&application).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	timeline := []gin.H{}

	// Application submitted
	timeline = append(timeline, gin.H{
		"type":        "application_submitted",
		"title":       "Application Submitted",
		"description": "Candidate applied for the position",
		"timestamp":   application.AppliedAt,
		"icon":        "📝",
	})

	// CV viewed
	if application.CVViewedAt != nil {
		timeline = append(timeline, gin.H{
			"type":        "cv_viewed",
			"title":       "CV Viewed",
			"description": "Recruiter viewed the candidate's CV",
			"timestamp":   application.CVViewedAt,
			"icon":        "👁️",
		})
	}

	// Status changes
	if application.ReviewedAt != nil {
		timeline = append(timeline, gin.H{
			"type":        "status_changed",
			"title":       "Status: " + application.Status,
			"description": "Application status changed to " + application.Status,
			"timestamp":   application.ReviewedAt,
			"icon":        getStatusIcon(application.Status),
		})
	}

	// Talent pool
	if application.InTalentPool && application.TalentPoolAddedAt != nil {
		timeline = append(timeline, gin.H{
			"type":        "talent_pool",
			"title":       "Added to Talent Pool",
			"description": "Candidate marked for future opportunities",
			"timestamp":   application.TalentPoolAddedAt,
			"icon":        "⭐",
		})
	}

	// Get notes
	var notes []models.CandidateNote
	config.DB.Where("application_id = ?", applicationID).
		Preload("Admin").
		Order("created_at DESC").
		Find(&notes)

	for _, note := range notes {
		timeline = append(timeline, gin.H{
			"type":        "note",
			"title":       "Note Added",
			"description": note.Note,
			"timestamp":   note.CreatedAt,
			"icon":        "📝",
			"author":      note.Admin.Name,
			"is_private":  note.IsPrivate,
		})
	}

	// Get messages
	var messages []models.Message
	config.DB.Where("application_id = ?", applicationID).
		Order("created_at DESC").
		Find(&messages)

	for _, msg := range messages {
		timeline = append(timeline, gin.H{
			"type":        "message",
			"title":       "Message: " + msg.SenderType,
			"description": msg.Message,
			"timestamp":   msg.CreatedAt,
			"icon":        "💬",
			"sender":      msg.SenderEmail,
		})
	}

	// Get activity logs
	var activityLogs []models.ActivityLog
	config.DB.Where("entity_type = 'application' AND entity_id = ?", applicationID).
		Preload("Admin").
		Order("created_at DESC").
		Find(&activityLogs)

	for _, log := range activityLogs {
		adminName := "System"
		if log.Admin != nil {
			adminName = log.Admin.Name
		}
		timeline = append(timeline, gin.H{
			"type":        "activity",
			"title":       log.Description,
			"description": log.ActionType,
			"timestamp":   log.CreatedAt,
			"icon":        "📋",
			"admin":       adminName,
		})
	}

	// Sort by timestamp (newest first)
	// Note: In a real implementation, you'd want to sort this properly
	// For now, we'll return as-is since we're ordering each query by DESC

	c.JSON(http.StatusOK, gin.H{
		"timeline": timeline,
		"count":    len(timeline),
	})
}

// Helper function to get status icon
func getStatusIcon(status string) string {
	switch status {
	case "shortlisted":
		return "✅"
	case "rejected":
		return "❌"
	case "cv_viewed":
		return "👁️"
	default:
		return "📋"
	}
}

