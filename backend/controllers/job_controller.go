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
	// Required fields ignore blanks; optional ones can be cleared by sending "".
	setRequired := func(dst *string, src *string) {
		if src != nil && *src != "" {
			*dst = *src
		}
	}
	setOptional := func(dst *string, src *string) {
		if src != nil {
			*dst = strings.TrimSpace(*src)
		}
	}
	setRequired(&job.Title, r.Title)
	setRequired(&job.Description, r.Description)
	setOptional(&job.Requirements, r.Requirements)
	setOptional(&job.Location, r.Location)
	setOptional(&job.JobType, r.JobType)
	setOptional(&job.SalaryRange, r.SalaryRange)
	setRequired(&job.Status, r.Status)

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

	if !planCheck(c, services.CheckCanOpenJob(companyID(c))) {
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

// publicJob is what candidates see; internal fields like company_id and screening settings stay private.
type publicJob struct {
	ID           uuid.UUID       `json:"id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Requirements string          `json:"requirements"`
	Location     string          `json:"location"`
	JobType      string          `json:"job_type"`
	SalaryRange  string          `json:"salary_range"`
	Deadline     models.DateOnly `json:"deadline"`
	CreatedAt    time.Time       `json:"created_at"`
}

func toPublicJob(j *models.Job) publicJob {
	return publicJob{j.ID, j.Title, j.Description, j.Requirements, j.Location, j.JobType, j.SalaryRange, j.Deadline, j.CreatedAt}
}

// publicCompanyName returns "" for unknown or suspended companies, which takes their public pages offline.
func publicCompanyName(id uuid.UUID) string {
	var company models.Company
	err := config.DB.Select("company_name", "subscription_status").First(&company, "id = ?", id).Error
	if err != nil || company.SubscriptionStatus == models.CompanySuspended {
		return ""
	}
	return company.CompanyName
}

// GetPublicJobs lists open jobs for a company's public careers page.
func GetPublicJobs(c *gin.Context) {
	company, ok := uuidParam(c, "companyId")
	if !ok {
		return
	}
	companyName := publicCompanyName(company)
	if companyName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
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
	out := make([]publicJob, len(jobs))
	for i := range jobs {
		out[i] = toPublicJob(&jobs[i])
	}
	c.JSON(http.StatusOK, gin.H{"jobs": out, "company_name": companyName})
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
	companyName := publicCompanyName(job.CompanyID)
	if companyName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"job":                    toPublicJob(&job),
		"company_id":             job.CompanyID,
		"company_name":           companyName,
		"accepting_applications": !job.Closed(time.Now().UTC()),
	})
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

	oldStatus, oldTitle, oldDeadline := job.Status, job.Title, job.Deadline.Time
	if err := req.apply(job); err != nil {
		badRequest(c, err)
		return
	}
	wasLive := oldStatus == "open" && !oldDeadline.Before(models.Today())
	if job.Status == "open" && !wasLive && !planCheck(c, services.CheckCanOpenJob(job.CompanyID)) {
		return
	}
	// GetJobs auto-closes expired jobs, so reopening one without a new deadline would silently revert.
	if job.Status == "open" && job.Closed(time.Now().UTC()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Deadline has passed",
			"message": "Extend the deadline to a future date before reopening this job.",
		})
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
