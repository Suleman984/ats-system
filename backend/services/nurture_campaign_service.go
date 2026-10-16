package services

import (
	"fmt"
	"html"
	"log"
	"time"

	"ats-backend/config"
	"ats-backend/models"

	"github.com/google/uuid"
)

// Talent-pool nurture emails. Not yet scheduled: wire ProcessMonthlyNurtureCampaigns to a cron to enable.

const nurtureInterval = 30 * 24 * time.Hour

func careersURL(companyID uuid.UUID) string {
	return config.FrontendURL() + "/jobs/" + companyID.String()
}

func SendJobAlert(app *models.Application, job *models.Job) error {
	subject := "New Opportunity: " + job.Title
	body := fmt.Sprintf(`<p>Based on your previous application, we think you might be a great fit for:</p>
    <div style="background-color: #f3f4f6; padding: 15px; border-radius: 5px; margin: 20px 0;">
      <h3 style="margin: 0; color: #2563eb;">%s</h3>
    </div>
    <p style="font-size: 12px; color: #666;">You're receiving this because you're in our talent pool.</p>`,
		html.EscapeString(job.Title))

	if err := SendEmail(app.Email, subject,
		renderEmail("Hello "+app.FullName+",", "#2563eb", body, "View Job Details", careersURL(job.CompanyID))); err != nil {
		return err
	}
	return logCampaign(app.ID, &job.ID, "job_alert", subject)
}

func SendMonthlyCheckIn(app *models.Application) error {
	subject := "Stay in Touch - New Opportunities Available"
	body := `<p>We keep you in our talent pool because we believe you have great potential.
    Check out our latest openings.</p>
    <p style="font-size: 12px; color: #666;">You're receiving this monthly update because you're in our talent pool.</p>`

	if err := SendEmail(app.Email, subject,
		renderEmail("Hello "+app.FullName+",", "#2563eb", body, "View Open Positions", careersURL(app.CompanyID))); err != nil {
		return err
	}
	return logCampaign(app.ID, nil, "check_in", subject)
}

func logCampaign(applicationID uuid.UUID, jobID *uuid.UUID, emailType, subject string) error {
	return config.DB.Create(&models.NurtureCampaign{
		ApplicationID: applicationID,
		JobID:         jobID,
		EmailSentAt:   time.Now(),
		EmailType:     emailType,
		Subject:       subject,
		Status:        "sent",
	}).Error
}

// ProcessMonthlyNurtureCampaigns emails talent-pool candidates not checked in with for 30 days.
func ProcessMonthlyNurtureCampaigns() error {
	cutoff := time.Now().Add(-nurtureInterval)
	var applications []models.Application
	err := config.DB.Omit("parsed_cv_text").
		Where("in_talent_pool AND (talent_pool_added_at IS NULL OR talent_pool_added_at < ?)", cutoff).
		Where("NOT EXISTS (SELECT 1 FROM nurture_campaigns n WHERE n.application_id = applications.id AND n.email_type = 'check_in' AND n.email_sent_at > ?)", cutoff).
		Find(&applications).Error
	if err != nil {
		return err
	}

	for i := range applications {
		if err := SendMonthlyCheckIn(&applications[i]); err != nil {
			log.Printf("nurture check-in for application %s: %v", applications[i].ID, err)
		}
	}
	return nil
}
