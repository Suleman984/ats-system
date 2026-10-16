package services

import (
	"encoding/json"
	"log"

	"ats-backend/config"
	"ats-backend/models"

	"github.com/google/uuid"
)

// LogActivity writes an audit entry in the background so requests never wait on it.
func LogActivity(companyID, adminID *uuid.UUID, actionType, entityType string, entityID *uuid.UUID, description string, metadata map[string]any) {
	entry := models.ActivityLog{
		CompanyID:   companyID,
		AdminID:     adminID,
		ActionType:  actionType,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
	}
	if metadata != nil {
		if raw, err := json.Marshal(metadata); err == nil {
			s := string(raw)
			entry.Metadata = &s
		}
	}

	go func() {
		if err := config.DB.Create(&entry).Error; err != nil {
			log.Printf("activity log: %v", err)
		}
	}()
}

func LogCompanyRegistered(companyID, adminID uuid.UUID, companyName, adminEmail string) {
	LogActivity(&companyID, &adminID, "company_registered", "company", &companyID,
		"Company registered: "+companyName+" by "+adminEmail,
		map[string]any{"company_name": companyName, "admin_email": adminEmail})
}

func LogJobCreated(companyID, adminID, jobID uuid.UUID, jobTitle string) {
	LogActivity(&companyID, &adminID, "job_created", "job", &jobID,
		"Job created: "+jobTitle, map[string]any{"job_title": jobTitle})
}

func LogJobUpdated(companyID, adminID, jobID uuid.UUID, jobTitle string, changes map[string]any) {
	LogActivity(&companyID, &adminID, "job_updated", "job", &jobID, "Job updated: "+jobTitle, changes)
}

func LogJobDeleted(companyID, adminID, jobID uuid.UUID, jobTitle string) {
	LogActivity(&companyID, &adminID, "job_deleted", "job", &jobID,
		"Job deleted: "+jobTitle, map[string]any{"job_title": jobTitle})
}

func LogJobStatusChanged(companyID, adminID, jobID uuid.UUID, jobTitle, oldStatus, newStatus string) {
	LogActivity(&companyID, &adminID, "job_status_changed", "job", &jobID,
		"Job status changed: "+jobTitle+" from "+oldStatus+" to "+newStatus,
		map[string]any{"job_title": jobTitle, "old_status": oldStatus, "new_status": newStatus})
}

func LogApplicationStatusChanged(companyID, adminID, applicationID uuid.UUID, candidateName, jobTitle, oldStatus, newStatus string) {
	LogActivity(&companyID, &adminID, "application_status_changed", "application", &applicationID,
		"Application status changed: "+candidateName+" for "+jobTitle+" from "+oldStatus+" to "+newStatus,
		map[string]any{"candidate_name": candidateName, "job_title": jobTitle, "old_status": oldStatus, "new_status": newStatus})
}
