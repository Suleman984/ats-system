package controllers

import (
	"errors"
	"net/http"
	"sort"
	"time"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type noteRequest struct {
	Note      string `json:"note" binding:"required,max=10000"`
	IsPrivate bool   `json:"is_private"`
}

func AddCandidateNote(c *gin.Context) {
	var req struct {
		ApplicationID uuid.UUID `json:"application_id" binding:"required"`
		noteRequest
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	app, ok := findCompanyApplication(c, req.ApplicationID)
	if !ok {
		return
	}

	note := models.CandidateNote{ApplicationID: app.ID, AdminID: adminID(c), Note: req.Note, IsPrivate: req.IsPrivate}
	if err := config.DB.Create(&note).Error; err != nil {
		serverError(c, "Failed to add note")
		return
	}
	config.DB.Preload("Admin").First(&note, "id = ?", note.ID)
	c.JSON(http.StatusCreated, gin.H{"message": "Note added successfully", "note": note})
}

// visibleNotes returns shared notes plus the caller's own private ones.
func visibleNotes(c *gin.Context, applicationID uuid.UUID) ([]models.CandidateNote, error) {
	var notes []models.CandidateNote
	err := config.DB.Preload("Admin").
		Where("application_id = ? AND (is_private = false OR admin_id = ?)", applicationID, adminID(c)).
		Order("created_at desc").
		Find(&notes).Error
	return notes, err
}

func GetCandidateNotes(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	if _, ok := findCompanyApplication(c, id); !ok {
		return
	}
	notes, err := visibleNotes(c, id)
	if err != nil {
		serverError(c, "Failed to fetch notes")
		return
	}
	c.JSON(http.StatusOK, gin.H{"notes": notes, "count": len(notes)})
}

// findOwnNote loads a note written by the caller; only authors may edit or delete.
func findOwnNote(c *gin.Context) (*models.CandidateNote, bool) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return nil, false
	}
	var note models.CandidateNote
	err := config.DB.Where("id = ? AND admin_id = ?", id, adminID(c)).First(&note).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found or you don't have permission to change it"})
		return nil, false
	}
	if err != nil {
		serverError(c, "Failed to load note")
		return nil, false
	}
	return &note, true
}

func UpdateCandidateNote(c *gin.Context) {
	var req noteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	note, ok := findOwnNote(c)
	if !ok {
		return
	}
	if err := config.DB.Model(note).Updates(map[string]any{"note": req.Note, "is_private": req.IsPrivate}).Error; err != nil {
		serverError(c, "Failed to update note")
		return
	}
	config.DB.Preload("Admin").First(note, "id = ?", note.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Note updated successfully", "note": note})
}

func DeleteCandidateNote(c *gin.Context) {
	note, ok := findOwnNote(c)
	if !ok {
		return
	}
	if err := config.DB.Delete(note).Error; err != nil {
		serverError(c, "Failed to delete note")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

func AddToTalentPool(c *gin.Context) {
	var req struct {
		ApplicationID uuid.UUID `json:"application_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	setTalentPool(c, req.ApplicationID, true)
}

func RemoveFromTalentPool(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	setTalentPool(c, id, false)
}

func setTalentPool(c *gin.Context, applicationID uuid.UUID, inPool bool) {
	app, ok := findCompanyApplication(c, applicationID)
	if !ok {
		return
	}

	admin := adminID(c)
	app.InTalentPool, app.TalentPoolAddedAt, app.TalentPoolAddedBy = inPool, nil, nil
	if inPool {
		app.TalentPoolAddedAt, app.TalentPoolAddedBy = ptr(time.Now()), &admin
	}
	err := config.DB.Model(app).Updates(map[string]any{
		"in_talent_pool":       app.InTalentPool,
		"talent_pool_added_at": app.TalentPoolAddedAt,
		"talent_pool_added_by": app.TalentPoolAddedBy,
	}).Error
	if err != nil {
		serverError(c, "Failed to update talent pool")
		return
	}

	message := "Candidate removed from talent pool"
	if inPool {
		message = "Candidate added to talent pool"
		company := app.CompanyID
		services.LogActivity(&company, &admin, "candidate_added_to_talent_pool", "application", &app.ID,
			message+": "+app.FullName, map[string]any{"candidate_name": app.FullName, "candidate_email": app.Email})
	}
	c.JSON(http.StatusOK, gin.H{"message": message, "application": app})
}

func GetTalentPool(c *gin.Context) {
	var applications []models.Application
	err := config.DB.Preload("Job").Omit("parsed_cv_text").
		Where("company_id = ? AND in_talent_pool", companyID(c)).
		Order("talent_pool_added_at desc").
		Find(&applications).Error
	if err != nil {
		serverError(c, "Failed to fetch talent pool")
		return
	}
	c.JSON(http.StatusOK, gin.H{"applications": applications, "count": len(applications)})
}

func UpdateReferralInfo(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		ReferralSource  string `json:"referral_source" binding:"max=255"`
		ReferredByName  string `json:"referred_by_name" binding:"max=255"`
		ReferredByEmail string `json:"referred_by_email" binding:"omitempty,email,max=255"`
		ReferredByPhone string `json:"referred_by_phone" binding:"max=50"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}

	app.ReferralSource, app.ReferredByName = req.ReferralSource, req.ReferredByName
	app.ReferredByEmail, app.ReferredByPhone = req.ReferredByEmail, req.ReferredByPhone
	err := config.DB.Model(app).Updates(map[string]any{
		"referral_source":   req.ReferralSource,
		"referred_by_name":  req.ReferredByName,
		"referred_by_email": req.ReferredByEmail,
		"referred_by_phone": req.ReferredByPhone,
	}).Error
	if err != nil {
		serverError(c, "Failed to update referral info")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Referral information updated", "application": app})
}

type timelineItem struct {
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Icon        string    `json:"icon"`
	Author      string    `json:"author,omitempty"`
	Admin       string    `json:"admin,omitempty"`
	Sender      string    `json:"sender,omitempty"`
	IsPrivate   bool      `json:"is_private,omitempty"`
}

// GetRelationshipTimeline merges status events, notes, messages and audit logs, newest first.
func GetRelationshipTimeline(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}

	timeline := []timelineItem{{
		Type: "application_submitted", Title: "Application Submitted",
		Description: "Candidate applied for the position", Timestamp: app.AppliedAt, Icon: "📝",
	}}
	if app.CVViewedAt != nil {
		timeline = append(timeline, timelineItem{Type: "cv_viewed", Title: "CV Viewed",
			Description: "Recruiter viewed the candidate's CV", Timestamp: *app.CVViewedAt, Icon: "👁️"})
	}
	if app.ReviewedAt != nil {
		timeline = append(timeline, timelineItem{Type: "status_changed", Title: "Status: " + app.Status,
			Description: "Application status changed to " + app.Status, Timestamp: *app.ReviewedAt, Icon: statusIcon(app.Status)})
	}
	if app.InTalentPool && app.TalentPoolAddedAt != nil {
		timeline = append(timeline, timelineItem{Type: "talent_pool", Title: "Added to Talent Pool",
			Description: "Candidate marked for future opportunities", Timestamp: *app.TalentPoolAddedAt, Icon: "⭐"})
	}

	notes, err := visibleNotes(c, app.ID)
	if err != nil {
		serverError(c, "Failed to load notes")
		return
	}
	for _, n := range notes {
		item := timelineItem{Type: "note", Title: "Note Added", Description: n.Note, Timestamp: n.CreatedAt, Icon: "📝", IsPrivate: n.IsPrivate}
		if n.Admin != nil {
			item.Author = n.Admin.Name
		}
		timeline = append(timeline, item)
	}

	var messages []models.Message
	config.DB.Where("application_id = ?", app.ID).Find(&messages)
	for _, m := range messages {
		timeline = append(timeline, timelineItem{Type: "message", Title: "Message: " + m.SenderType,
			Description: m.Message, Timestamp: m.CreatedAt, Icon: "💬", Sender: m.SenderEmail})
	}

	var logs []models.ActivityLog
	config.DB.Preload("Admin").Where("entity_type = 'application' AND entity_id = ?", app.ID).Find(&logs)
	for _, l := range logs {
		item := timelineItem{Type: "activity", Title: l.Description, Description: l.ActionType, Timestamp: l.CreatedAt, Icon: "📋", Admin: "System"}
		if l.Admin != nil {
			item.Admin = l.Admin.Name
		}
		timeline = append(timeline, item)
	}

	sort.SliceStable(timeline, func(i, j int) bool { return timeline[i].Timestamp.After(timeline[j].Timestamp) })
	c.JSON(http.StatusOK, gin.H{"timeline": timeline, "count": len(timeline)})
}

func statusIcon(status string) string {
	switch status {
	case models.StatusShortlisted:
		return "✅"
	case models.StatusRejected:
		return "❌"
	case models.StatusCVViewed:
		return "👁️"
	default:
		return "📋"
	}
}
