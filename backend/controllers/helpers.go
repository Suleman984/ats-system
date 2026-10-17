package controllers

import (
	"errors"
	"net/http"
	"strings"

	"ats-backend/config"
	"ats-backend/middleware"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// companyID and adminID are always present behind AuthMiddleware.
func companyID(c *gin.Context) uuid.UUID {
	return c.MustGet(middleware.CtxCompanyID).(uuid.UUID)
}

func adminID(c *gin.Context) uuid.UUID {
	return c.MustGet(middleware.CtxAdminID).(uuid.UUID)
}

// uuidParam writes a 400 and returns false when the path parameter is not a UUID.
func uuidParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + name})
		return uuid.Nil, false
	}
	return id, true
}

// findCompanyApplication loads an application owned by the caller's company, or writes a 404.
func findCompanyApplication(c *gin.Context, applicationID uuid.UUID) (*models.Application, bool) {
	var app models.Application
	err := config.DB.Preload("Job").
		Where("id = ? AND company_id = ?", applicationID, companyID(c)).
		First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return nil, false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load application"})
		return nil, false
	}
	return &app, true
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func serverError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

func ptr[T any](v T) *T { return &v }

// planCheck writes a 402 for plan-limit refusals (or a 500 for lookup failures) and returns false.
func planCheck(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, services.ErrPlanLimit) {
		msg := strings.TrimPrefix(err.Error(), services.ErrPlanLimit.Error()+": ")
		c.JSON(http.StatusPaymentRequired, gin.H{"error": msg, "code": "plan_limit"})
		return false
	}
	serverError(c, "Failed to check plan limits")
	return false
}

// scoringJob returns the job to score CVs against, or nil when the plan excludes CV scoring
// (the CV text is still extracted so the candidate stays searchable).
func scoringJob(job *models.Job) *models.Job {
	if services.CheckCanScoreCVs(job.CompanyID) != nil {
		return nil
	}
	return job
}
