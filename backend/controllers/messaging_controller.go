package controllers

import (
	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SendMessageRequest for sending messages
type SendMessageRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
	Message       string `json:"message" binding:"required"`
	SenderEmail   string `json:"sender_email" binding:"required,email"`
}

// GetMessagesRequest for retrieving messages
type GetMessagesRequest struct {
	ApplicationID string `json:"application_id" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
}

// SendMessage handles sending messages (both candidate and recruiter)
func SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if application exists
	var application models.Application
	if err := config.DB.Where("id = ?", req.ApplicationID).First(&application).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}

	// Determine sender type and ID
	senderType := "candidate"
	var senderID *uuid.UUID

	// Check if sender is a recruiter (admin)
	_, exists := c.Get("company_id")
	if exists {
		// This is a protected route, so sender is a recruiter
		senderType = "recruiter"
		adminIDVal, _ := c.Get("admin_id")
		adminIDStr, _ := adminIDVal.(string)
		if adminIDStr != "" {
			adminUUID, _ := uuid.Parse(adminIDStr)
			senderID = &adminUUID
		}
	} else {
		// Public route - verify email matches application email
		if req.SenderEmail != application.Email {
			c.JSON(http.StatusForbidden, gin.H{"error": "Email does not match application"})
			return
		}
	}

	// Create message
	message := models.Message{
		ApplicationID: uuid.MustParse(req.ApplicationID),
		SenderType:    senderType,
		SenderID:      senderID,
		SenderEmail:   req.SenderEmail,
		Message:       req.Message,
		IsRead:        false,
		CreatedAt:     time.Now(),
	}

	if err := config.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// If recruiter sent message, notify candidate via email (async)
	if senderType == "recruiter" {
		go func() {
			jobTitle := "Unknown Job"
			if application.JobID != nil {
				var job models.Job
				if err := config.DB.First(&job, "id = ?", application.JobID).Error; err == nil {
					jobTitle = job.Title
				}
			}

			// Send email notification to candidate
			emailSubject := "New Message About Your Application - " + jobTitle
			emailHTML := `
				<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
				</head>
				<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
					<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
						<h2 style="color: #2563eb;">Hello ` + application.FullName + `,</h2>
						<p>You have received a new message regarding your application for <strong>` + jobTitle + `</strong>.</p>
						<div style="background-color: #f3f4f6; padding: 15px; border-radius: 5px; margin: 20px 0;">
							<p style="margin: 0;">` + req.Message + `</p>
						</div>
						<p>You can reply to this message by visiting your application status portal.</p>
						<p style="text-align: center; margin: 20px 0;">
							<a href="` + getFrontendURL() + `/application-status" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">View Application Status</a>
						</p>
						<br>
						<p>Best regards,<br>The Hiring Team</p>
					</div>
				</body>
				</html>
			`

			if err := services.SendCustomEmail(application.Email, emailSubject, emailHTML); err != nil {
				log.Printf("ERROR: Failed to send message notification email to %s: %v", application.Email, err)
			}

			// Send SMS if phone number available
			if application.Phone != "" {
				smsMessage := "Hi " + application.FullName + ", you have a new message about your application for " + jobTitle + ". Check your email or application portal."
				if err := services.SendSMS(application.Phone, smsMessage); err != nil {
					log.Printf("ERROR: Failed to send message notification SMS to %s: %v", application.Phone, err)
				}
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message sent successfully",
		"data":    message,
	})
}

// GetMessages retrieves messages for an application
func GetMessages(c *gin.Context) {
	applicationID := c.Query("application_id")
	email := c.Query("email")

	if applicationID == "" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "application_id and email are required"})
		return
	}

	// Verify application exists and email matches
	var application models.Application
	if err := config.DB.Where("id = ? AND email = ?", applicationID, email).First(&application).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found or email does not match"})
		return
	}

	// Get messages for this application
	var messages []models.Message
	if err := config.DB.Where("application_id = ?", applicationID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	// Mark messages as read if viewed by candidate
	now := time.Now()
	for i := range messages {
		if !messages[i].IsRead && messages[i].SenderType == "recruiter" {
			messages[i].IsRead = true
			readAt := now
			messages[i].ReadAt = &readAt
			config.DB.Save(&messages[i])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetMessagesForRecruiter retrieves messages for an application (protected route)
func GetMessagesForRecruiter(c *gin.Context) {
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

	// Get messages for this application
	var messages []models.Message
	if err := config.DB.Where("application_id = ?", applicationID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	// Mark messages as read if viewed by recruiter
	now := time.Now()
	for i := range messages {
		if !messages[i].IsRead && messages[i].SenderType == "candidate" {
			messages[i].IsRead = true
			readAt := now
			messages[i].ReadAt = &readAt
			config.DB.Save(&messages[i])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// Helper function to get frontend URL
func getFrontendURL() string {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		return "http://localhost:3000"
	}
	return frontendURL
}


