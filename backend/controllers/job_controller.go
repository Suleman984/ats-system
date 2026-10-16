package controllers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// jobRequest uses pointers so updates only touch fields the client sent.
type jobRequest struct {
	Title             *string `json:"title" binding:"omitempty,max=255"`
	Description       *string `json:"description"`
	Requirements      *string `json:"requirements"`
	Location          *string `json:"location" binding:"omitempty,max=255"`
	JobType           *string `json:"job_type" binding:"omitempty,max=50"`
	SalaryRange       *string `json:"salary_range" binding:"omitempty,max=100"`
	Deadline          *string `json:"deadline"`
	Status            *string `json:"status" binding:"omitempty,oneof=open closed archived"`
	AutoShortlist     *bool   `json:"auto_shortlist"`
	ShortlistCriteria *string `json:"shortlist_criteria"`
}

func (r *jobRequest) apply(job *models.Job) error {
	set := func(dst *string, src *string) {
		if src != nil && *src != "" {
			*dst = *src
		}
	}
	set(&job.Title, r.Title)
	set(&job.Description, r.Description)
	set(&job.Requirements, r.Requirements)
	set(&job.Location, r.Location)
	set(&job.JobType, r.JobType)
	set(&job.SalaryRange, r.SalaryRange)
	set(&job.Status, r.Status)

	if r.Deadline != nil && *r.Deadline != "" {
		deadline, err := models.ParseDateOnly(strings.TrimSpace(*r.Deadline))
		if err != nil {
			return err
		}
		job.Deadline = deadline
	}
	if r.AutoShortlist != nil {
		job.AutoShortlist = *r.AutoShortlist
	}
	if r.ShortlistCriteria != nil {
		job.ShortlistCriteria = nil
		if *r.ShortlistCriteria != "" {
			job.ShortlistCriteria = r.ShortlistCriteria
		}
	}
	return nil
}

func CreateJob(c *gin.Context) {
	var req jobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if req.Title == nil || *req.Title == "" || req.Description == nil || *req.Description == "" || req.Deadline == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title, description and deadline are required"})
		return
	}

	job := models.Job{CompanyID: companyID(c), Status: "open", AutoShortlist: true}
	if err := req.apply(&job); err != nil {
		badRequest(c, err)
		return
	}
	if err := config.DB.Create(&job).Error; err != nil {
		serverError(c, "Failed to create job")
		return
	}

	services.LogJobCreated(job.CompanyID, adminID(c), job.ID, job.Title)
	c.JSON(http.StatusCreated, gin.H{"message": "Job created successfully", "job": job})
}

func GetJobs(c *gin.Context) {
	company := companyID(c)
	today := time.Now().UTC().Format("2006-01-02")
	if err := config.DB.Model(&models.Job{}).
		Where("company_id = ? AND status = 'open' AND deadline < ?", company, today).
		Update("status", "closed").Error; err != nil {
		log.Printf("auto-close expired jobs for %s: %v", company, err)
	}

	query := config.DB.Where("company_id = ?", company).Order("created_at desc")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var jobs []models.Job
	if err := query.Find(&jobs).Error; err != nil {
		serverError(c, "Failed to fetch jobs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

func GetJob(c *gin.Context) {
	job, ok := findCompanyJob(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"job": job})
}

// GetPublicJobs lists open jobs for a company's public careers page.
func GetPublicJobs(c *gin.Context) {
	company, ok := uuidParam(c, "companyId")
	if !ok {
		return
	}
	var jobs []models.Job
	err := config.DB.
		Omit("shortlist_criteria").
		Where("company_id = ? AND status = 'open' AND deadline >= ?", company, time.Now().UTC().Format("2006-01-02")).
		Order("created_at desc").
		Find(&jobs).Error
	if err != nil {
		serverError(c, "Failed to fetch jobs")
		return
	}
	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

// GetPublicJob returns one job for the public apply page, including whether it still accepts applications.
func GetPublicJob(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	var job models.Job
	err := config.DB.Omit("shortlist_criteria").First(&job, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	if err != nil {
		serverError(c, "Failed to load job")
		return
	}
	c.JSON(http.StatusOK, gin.H{"job": job, "accepting_applications": !job.Closed(time.Now().UTC())})
}

func UpdateJob(c *gin.Context) {
	job, ok := findCompanyJob(c)
	if !ok {
		return
	}
	var req jobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}

	oldStatus, oldTitle := job.Status, job.Title
	if err := req.apply(job); err != nil {
		badRequest(c, err)
		return
	}
	if err := config.DB.Save(job).Error; err != nil {
		serverError(c, "Failed to update job")
		return
	}

	if job.Status != oldStatus {
		services.LogJobStatusChanged(job.CompanyID, adminID(c), job.ID, job.Title, oldStatus, job.Status)
	} else {
		changes := map[string]any{"updated": true}
		if job.Title != oldTitle {
			changes = map[string]any{"title": map[string]string{"old": oldTitle, "new": job.Title}}
		}
		services.LogJobUpdated(job.CompanyID, adminID(c), job.ID, job.Title, changes)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Job updated successfully", "job": job})
}

// DeleteJob keeps the job's applications; the FK sets their job_id to null.
func DeleteJob(c *gin.Context) {
	job, ok := findCompanyJob(c)
	if !ok {
		return
	}
	if err := config.DB.Delete(job).Error; err != nil {
		serverError(c, "Failed to delete job")
		return
	}
	services.LogJobDeleted(job.CompanyID, adminID(c), job.ID, job.Title)
	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

func findCompanyJob(c *gin.Context) (*models.Job, bool) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return nil, false
	}
	return loadCompanyJob(c, id)
}

func loadCompanyJob(c *gin.Context, id uuid.UUID) (*models.Job, bool) {
	var job models.Job
	err := config.DB.Where("id = ? AND company_id = ?", id, companyID(c)).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return nil, false
	}
	if err != nil {
		serverError(c, "Failed to load job")
		return nil, false
	}
	return &job, true
}
