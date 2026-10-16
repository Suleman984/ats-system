package controllers

import (
	"log"
	"net/http"
	"time"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SendCandidateMessage is the public endpoint; the candidate proves ownership with their email.
func SendCandidateMessage(c *gin.Context) {
	var req struct {
		ApplicationID uuid.UUID `json:"application_id" binding:"required"`
		SenderEmail   string    `json:"sender_email" binding:"required,email"`
		Message       string    `json:"message" binding:"required,max=5000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	app, ok := findCandidateApplication(c, req.ApplicationID, req.SenderEmail)
	if !ok {
		return
	}

	msg := models.Message{ApplicationID: app.ID, SenderType: models.SenderCandidate, SenderEmail: app.Email, Message: req.Message}
	if err := config.DB.Create(&msg).Error; err != nil {
		serverError(c, "Failed to send message")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Message sent successfully", "data": msg})
}

func SendRecruiterMessage(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		Message string `json:"message" binding:"required,max=5000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}

	var admin models.Admin
	if err := config.DB.Select("id", "email").First(&admin, "id = ?", adminID(c)).Error; err != nil {
		serverError(c, "Failed to load sender")
		return
	}

	msg := models.Message{
		ApplicationID: app.ID,
		SenderType:    models.SenderRecruiter,
		SenderID:      &admin.ID,
		SenderEmail:   admin.Email,
		Message:       req.Message,
	}
	if err := config.DB.Create(&msg).Error; err != nil {
		serverError(c, "Failed to send message")
		return
	}

	jobTitle := app.JobTitle()
	go func() {
		if err := services.SendNewMessageEmail(app.Email, app.FullName, jobTitle, req.Message); err != nil {
			log.Printf("message email for application %s: %v", app.ID, err)
		}
		sms := "Hi " + app.FullName + ", you have a new message about your application for " + jobTitle + ". Check your email or application portal."
		if err := services.SendSMS(app.Phone, sms); err != nil {
			log.Printf("message SMS for application %s: %v", app.ID, err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"message": "Message sent successfully", "data": msg})
}

func GetCandidateMessages(c *gin.Context) {
	id, err := uuid.Parse(c.Query("application_id"))
	if err != nil || c.Query("email") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application_id and email are required"})
		return
	}
	app, ok := findCandidateApplication(c, id, c.Query("email"))
	if !ok {
		return
	}
	respondWithThread(c, app.ID, models.SenderRecruiter)
}

func GetRecruiterMessages(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}
	respondWithThread(c, app.ID, models.SenderCandidate)
}

// respondWithThread returns the conversation and marks the other side's messages read.
func respondWithThread(c *gin.Context, applicationID uuid.UUID, markReadFrom string) {
	var messages []models.Message
	if err := config.DB.Where("application_id = ?", applicationID).Order("created_at asc").Find(&messages).Error; err != nil {
		serverError(c, "Failed to fetch messages")
		return
	}

	now := time.Now()
	config.DB.Model(&models.Message{}).
		Where("application_id = ? AND sender_type = ? AND is_read = false", applicationID, markReadFrom).
		Updates(map[string]any{"is_read": true, "read_at": now})

	for i := range messages {
		if messages[i].SenderType == markReadFrom && !messages[i].IsRead {
			messages[i].IsRead, messages[i].ReadAt = true, &now
		}
	}
	c.JSON(http.StatusOK, gin.H{"messages": messages, "count": len(messages)})
}
