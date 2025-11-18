package services

import (
	"ats-backend/config"
	"ats-backend/models"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
)

// LogActivity creates an activity log entry
func LogActivity(companyID *uuid.UUID, adminID *uuid.UUID, actionType, entityType string, entityID *uuid.UUID, description string, metadata map[string]interface{}) {
	activityLog := models.ActivityLog{
		CompanyID:   companyID,
		AdminID:     adminID,
		ActionType:  actionType,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		CreatedAt:   time.Now(),
	}

	// Convert metadata to JSON string if provided
	if metadata != nil {
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			log.Printf("Failed to marshal activity log metadata: %v", err)
		} else {
			metadataStr := string(metadataJSON)
			activityLog.Metadata = &metadataStr
		}
	}

	// Save to database (async to avoid blocking)
	go func() {
		if err := config.DB.Create(&activityLog).Error; err != nil {
			log.Printf("Failed to create activity log: %v", err)
		}
	}()
}

// Helper functions for common actions

// LogCompanyRegistered logs when a company registers
func LogCompanyRegistered(companyID uuid.UUID, companyName, adminEmail string) {
	LogActivity(
		&companyID,
		nil, // No admin yet during registration
		"company_registered",
		"company",
		&companyID,
		"Company registered: "+companyName+" by "+adminEmail,
		map[string]interface{}{
			"company_name": companyName,
			"admin_email":  adminEmail,
		},
	)
}

// LogJobCreated logs when a job is created
func LogJobCreated(companyID, adminID uuid.UUID, jobID uuid.UUID, jobTitle string) {
	LogActivity(
		&companyID,
		&adminID,
		"job_created",
		"job",
		&jobID,
		"Job created: "+jobTitle,
		map[string]interface{}{
			"job_title": jobTitle,
		},
	)
}

// LogJobUpdated logs when a job is updated
func LogJobUpdated(companyID, adminID uuid.UUID, jobID uuid.UUID, jobTitle string, changes map[string]interface{}) {
	LogActivity(
		&companyID,
		&adminID,
		"job_updated",
		"job",
		&jobID,
		"Job updated: "+jobTitle,
		changes,
	)
}

// LogJobDeleted logs when a job is deleted
func LogJobDeleted(companyID, adminID uuid.UUID, jobID uuid.UUID, jobTitle string) {
	LogActivity(
		&companyID,
		&adminID,
		"job_deleted",
		"job",
		&jobID,
		"Job deleted: "+jobTitle,
		map[string]interface{}{
			"job_title": jobTitle,
		},
	)
}

// LogJobStatusChanged logs when a job status changes
func LogJobStatusChanged(companyID, adminID uuid.UUID, jobID uuid.UUID, jobTitle, oldStatus, newStatus string) {
	LogActivity(
		&companyID,
		&adminID,
		"job_status_changed",
		"job",
		&jobID,
		"Job status changed: "+jobTitle+" from "+oldStatus+" to "+newStatus,
		map[string]interface{}{
			"job_title":  jobTitle,
			"old_status": oldStatus,
			"new_status": newStatus,
		},
	)
}

// LogApplicationShortlisted logs when an application is shortlisted
func LogApplicationShortlisted(companyID, adminID uuid.UUID, applicationID uuid.UUID, candidateName, jobTitle string) {
	LogActivity(
		&companyID,
		&adminID,
		"application_shortlisted",
		"application",
		&applicationID,
		"Application shortlisted: "+candidateName+" for "+jobTitle,
		map[string]interface{}{
			"candidate_name": candidateName,
			"job_title":      jobTitle,
		},
	)
}

// LogApplicationRejected logs when an application is rejected
func LogApplicationRejected(companyID, adminID uuid.UUID, applicationID uuid.UUID, candidateName, jobTitle string) {
	LogActivity(
		&companyID,
		&adminID,
		"application_rejected",
		"application",
		&applicationID,
		"Application rejected: "+candidateName+" for "+jobTitle,
		map[string]interface{}{
			"candidate_name": candidateName,
			"job_title":      jobTitle,
		},
	)
}

// LogApplicationStatusChanged logs when an application status changes
func LogApplicationStatusChanged(companyID, adminID uuid.UUID, applicationID uuid.UUID, candidateName, jobTitle, oldStatus, newStatus string) {
	LogActivity(
		&companyID,
		&adminID,
		"application_status_changed",
		"application",
		&applicationID,
		"Application status changed: "+candidateName+" for "+jobTitle+" from "+oldStatus+" to "+newStatus,
		map[string]interface{}{
			"candidate_name": candidateName,
			"job_title":      jobTitle,
			"old_status":     oldStatus,
			"new_status":     newStatus,
		},
	)
}

