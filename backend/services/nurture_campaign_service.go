package services

import (
	"ats-backend/config"
	"ats-backend/models"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

// SendJobAlert sends a job alert email to a candidate in the talent pool
func SendJobAlert(applicationID, jobID string, candidateEmail, candidateName, jobTitle string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	subject := "New Opportunity: " + jobTitle
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #2563eb;">Hello ` + candidateName + `,</h2>
				<p>We have a new opportunity that might interest you!</p>
				<p>Based on your previous application, we think you might be a great fit for:</p>
				<div style="background-color: #f3f4f6; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<h3 style="margin: 0; color: #2563eb;">` + jobTitle + `</h3>
				</div>
				<p>If you're interested, you can apply directly through our portal:</p>
				<p style="text-align: center; margin: 20px 0;">
					<a href="` + frontendURL + `/jobs/public" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">View Job Details</a>
				</p>
				<p style="font-size: 12px; color: #666;">
					You're receiving this because you're in our talent pool. 
					<a href="` + frontendURL + `/application-status">Update your preferences</a> to customize these alerts.
				</p>
				<br>
				<p>Best regards,<br>The Hiring Team</p>
			</div>
		</body>
		</html>
	`

	if err := SendCustomEmail(candidateEmail, subject, html); err != nil {
		return err
	}

	// Log the campaign
	appUUID, _ := uuid.Parse(applicationID)
	jobUUID, _ := uuid.Parse(jobID)
	
	campaign := models.NurtureCampaign{
		ApplicationID: appUUID,
		JobID:         &jobUUID,
		EmailSentAt:   time.Now(),
		EmailType:     "job_alert",
		Subject:       subject,
		Status:        "sent",
		CreatedAt:     time.Now(),
	}

	if err := config.DB.Create(&campaign).Error; err != nil {
		log.Printf("ERROR: Failed to log nurture campaign: %v", err)
	}

	return nil
}

// SendMonthlyCheckIn sends a monthly check-in email to candidates in talent pool
func SendMonthlyCheckIn(applicationID, candidateEmail, candidateName string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	subject := "Stay in Touch - New Opportunities Available"
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h2 style="color: #2563eb;">Hello ` + candidateName + `,</h2>
				<p>We wanted to check in and let you know that we have new opportunities available!</p>
				<p>We keep you in our talent pool because we believe you have great potential. Check out our latest job openings:</p>
				<p style="text-align: center; margin: 20px 0;">
					<a href="` + frontendURL + `/jobs/public" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">View Open Positions</a>
				</p>
				<p style="font-size: 12px; color: #666;">
					You're receiving this monthly update because you're in our talent pool. 
					<a href="` + frontendURL + `/application-status">Update your preferences</a> to customize these alerts.
				</p>
				<br>
				<p>Best regards,<br>The Hiring Team</p>
			</div>
		</body>
		</html>
	`

	if err := SendCustomEmail(candidateEmail, subject, html); err != nil {
		return err
	}

	// Log the campaign
	appUUID, _ := uuid.Parse(applicationID)
	
	campaign := models.NurtureCampaign{
		ApplicationID: appUUID,
		EmailSentAt:   time.Now(),
		EmailType:     "check_in",
		Subject:       subject,
		Status:        "sent",
		CreatedAt:     time.Now(),
	}

	if err := config.DB.Create(&campaign).Error; err != nil {
		log.Printf("ERROR: Failed to log nurture campaign: %v", err)
	}

	return nil
}

// ProcessMonthlyNurtureCampaigns processes all talent pool candidates and sends monthly check-ins
// This should be run as a scheduled job (cron)
func ProcessMonthlyNurtureCampaigns() error {
	// Get all candidates in talent pool who haven't been contacted in the last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var applications []models.Application
	err := config.DB.Where("in_talent_pool = true").
		Where("talent_pool_added_at < ? OR talent_pool_added_at IS NULL", thirtyDaysAgo).
		Find(&applications).Error

	if err != nil {
		return err
	}

	log.Printf("Processing monthly nurture campaigns for %d candidates", len(applications))

	for _, app := range applications {
		// Check if we've sent a check-in in the last 30 days
		var recentCampaign models.NurtureCampaign
		err := config.DB.Where("application_id = ? AND email_type = 'check_in' AND email_sent_at > ?", 
			app.ID, thirtyDaysAgo).First(&recentCampaign).Error

		// If no recent campaign, send check-in
		if err != nil {
			if err := SendMonthlyCheckIn(app.ID.String(), app.Email, app.FullName); err != nil {
				log.Printf("ERROR: Failed to send monthly check-in to %s: %v", app.Email, err)
			} else {
				log.Printf("SUCCESS: Monthly check-in sent to %s", app.Email)
			}
		}
	}

	return nil
}

