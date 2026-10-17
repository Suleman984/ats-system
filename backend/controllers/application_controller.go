package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ats-backend/config"
	"ats-backend/models"
	"ats-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// expectedResponseWindow is the response time promised to candidates after each status change.
const expectedResponseWindow = 5 * 24 * time.Hour

// candidateFields are the only fields an applicant may set.
type candidateFields struct {
	FullName          string `json:"full_name" binding:"required,max=255"`
	Email             string `json:"email" binding:"required,email,max=255"`
	Phone             string `json:"phone" binding:"omitempty,max=50"`
	ResumeURL         string `json:"resume_url" binding:"omitempty,url"`
	CoverLetter       string `json:"cover_letter" binding:"max=10000"`
	YearsOfExperience int    `json:"years_of_experience" binding:"min=0,max=60"`
	CurrentPosition   string `json:"current_position" binding:"max=255"`
	LinkedinURL       string `json:"linkedin_url" binding:"omitempty,url,max=255"`
	PortfolioURL      string `json:"portfolio_url" binding:"omitempty,url,max=255"`
	ReferralSource    string `json:"referral_source" binding:"max=255"`
	ReferredByName    string `json:"referred_by_name" binding:"max=255"`
	ReferredByEmail   string `json:"referred_by_email" binding:"omitempty,email,max=255"`
	ReferredByPhone   string `json:"referred_by_phone" binding:"max=50"`

	// Files sent through /upload/* are referenced by upload ID instead of a URL.
	ResumeUploadID    *uuid.UUID `json:"resume_upload_id"`
	PortfolioUploadID *uuid.UUID `json:"portfolio_upload_id"`
}

// validateLinks rejects non-web schemes (e.g. javascript:) since recruiters open these as links,
// and requires exactly one CV source.
func (f *candidateFields) validateLinks() error {
	if (f.ResumeURL == "") == (f.ResumeUploadID == nil) {
		return errors.New("provide either a CV link or an uploaded CV")
	}
	if f.PortfolioURL != "" && f.PortfolioUploadID != nil {
		return errors.New("provide either a portfolio link or an uploaded portfolio, not both")
	}
	for name, raw := range map[string]string{"resume_url": f.ResumeURL, "linkedin_url": f.LinkedinURL, "portfolio_url": f.PortfolioURL} {
		if raw == "" {
			continue
		}
		if u, err := url.Parse(raw); err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return fmt.Errorf("%s must be an http or https link", name)
		}
	}
	return nil
}

var errUploadUnavailable = errors.New("upload missing, expired or already claimed")

// claimUpload marks an upload as used by exactly one application; it must be fresh and unclaimed.
func claimUpload(tx *gorm.DB, id uuid.UUID, kind string) (*models.Upload, error) {
	var claimed []models.Upload
	err := tx.Raw(`update uploads set claimed_at = now()
		where id = ? and kind = ? and claimed_at is null and created_at > now() - interval '24 hours'
		returning *`, id, kind).Scan(&claimed).Error
	if err != nil {
		return nil, err
	}
	if len(claimed) == 0 {
		return nil, errUploadUnavailable
	}
	return &claimed[0], nil
}

// createApplication claims any uploaded files and inserts the application in one transaction,
// so a failed insert (e.g. a duplicate) leaves the uploads claimable. Returns the claimed CV upload.
func createApplication(app *models.Application, f *candidateFields) (*models.Upload, error) {
	var resume *models.Upload
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if f.ResumeUploadID != nil {
			up, err := claimUpload(tx, *f.ResumeUploadID, models.UploadResume)
			if err != nil {
				return err
			}
			resume, app.ResumePath, app.ResumeURL = up, &up.ObjectPath, ""
		}
		if f.PortfolioUploadID != nil {
			up, err := claimUpload(tx, *f.PortfolioUploadID, models.UploadPortfolio)
			if err != nil {
				return err
			}
			app.PortfolioPath, app.PortfolioURL = &up.ObjectPath, ""
		}
		return tx.Create(app).Error
	})
	app.SetFileFlags()
	return resume, err
}

func uploadText(up *models.Upload) *string {
	if up == nil {
		return nil
	}
	return up.ParsedText
}

func (f *candidateFields) toApplication(job *models.Job) models.Application {
	return models.Application{
		JobID:             &job.ID,
		CompanyID:         job.CompanyID,
		FullName:          f.FullName,
		Email:             strings.ToLower(f.Email),
		Phone:             f.Phone,
		ResumeURL:         f.ResumeURL,
		CoverLetter:       f.CoverLetter,
		YearsOfExperience: f.YearsOfExperience,
		CurrentPosition:   f.CurrentPosition,
		LinkedinURL:       f.LinkedinURL,
		PortfolioURL:      f.PortfolioURL,
		ReferralSource:    f.ReferralSource,
		ReferredByName:    f.ReferredByName,
		ReferredByEmail:   f.ReferredByEmail,
		ReferredByPhone:   f.ReferredByPhone,
		Status:            models.StatusPending,
		AppliedAt:         time.Now(),
	}
}

type submitApplicationRequest struct {
	JobID uuid.UUID `json:"job_id" binding:"required"`
	candidateFields
}

func SubmitApplication(c *gin.Context) {
	var req submitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}
	if err := req.validateLinks(); err != nil {
		badRequest(c, err)
		return
	}

	var job models.Job
	err := config.DB.First(&job, "id = ?", req.JobID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	if err != nil {
		serverError(c, "Failed to load job")
		return
	}
	// A company over its monthly application quota stops accepting, without exposing plan details.
	if job.Closed(time.Now().UTC()) || publicCompanyName(job.CompanyID) == "" ||
		services.CheckCanAcceptApplication(job.CompanyID) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Applications closed",
			"message": "This job is no longer accepting applications.",
		})
		return
	}

	app := req.toApplication(&job)
	resumeUpload, err := createApplication(&app, &req.candidateFields)
	if errors.Is(err, errUploadUnavailable) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Your uploaded file has expired or was already used. Please upload it again."})
		return
	}
	if err != nil {
		// The ID is deliberately withheld: email + ID is the candidate's portal credential.
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Already applied",
				"message": "You've already applied for this job. Your application ID is in your confirmation email.",
			})
			return
		}
		serverError(c, "Failed to submit application")
		return
	}

	go services.ProcessApplicationCV(&app, uploadText(resumeUpload), scoringJob(&job))
	go func() {
		if err := services.SendConfirmationEmail(app.Email, app.FullName, job.Title, app.ID.String()); err != nil {
			log.Printf("confirmation email to application %s: %v", app.ID, err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Application submitted successfully!",
		"application": gin.H{"id": app.ID, "status": app.Status, "applied_at": app.AppliedAt},
	})
}

// GetDashboardStats returns the dashboard counts in one query each instead of shipping every row.
func GetDashboardStats(c *gin.Context) {
	company := companyID(c)
	today := time.Now().UTC().Format("2006-01-02")
	var stats struct {
		TotalJobs         int64 `json:"total_jobs"`
		OpenJobs          int64 `json:"open_jobs"`
		TotalApplications int64 `json:"total_applications"`
		Shortlisted       int64 `json:"shortlisted"`
	}
	err := config.DB.Raw(`select
		(select count(*) from jobs where company_id = @c) as total_jobs,
		(select count(*) from jobs where company_id = @c and status = 'open' and deadline >= @today) as open_jobs,
		(select count(*) from applications where company_id = @c and job_id is not null) as total_applications,
		(select count(*) from applications where company_id = @c and job_id is not null and status = 'shortlisted') as shortlisted`,
		map[string]any{"c": company, "today": today}).Scan(&stats).Error
	if err != nil {
		serverError(c, "Failed to load dashboard stats")
		return
	}

	var recent []models.Application
	if err := config.DB.Select("id", "full_name", "email", "status", "applied_at", "job_id").
		Where("company_id = ? AND job_id IS NOT NULL", company).
		Order("applied_at desc").Limit(5).Find(&recent).Error; err != nil {
		serverError(c, "Failed to load recent applications")
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats, "recent_applications": recent})
}

// GetApplications lists applications for the company's existing jobs; orphaned ones live in Find Candidates.
func GetApplications(c *gin.Context) {
	query := config.DB.Preload("Job").
		Omit("parsed_cv_text").
		Where("company_id = ? AND job_id IS NOT NULL", companyID(c)).
		Order("applied_at desc")

	if jobID := c.Query("job_id"); jobID != "" {
		query = query.Where("job_id = ?", jobID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if c.Query("unread") == "true" {
		query = query.Where(`exists (select 1 from messages m where m.application_id = applications.id
			and m.sender_type = ? and not m.is_read)`, models.SenderCandidate)
	}
	if from, err := time.Parse("2006-01-02", c.Query("date_from")); err == nil {
		query = query.Where("applied_at >= ?", from)
	}
	if to, err := time.Parse("2006-01-02", c.Query("date_to")); err == nil {
		query = query.Where("applied_at < ?", to.AddDate(0, 0, 1))
	}

	// Pagination is opt-in (limit/offset) so existing callers still get the full list.
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		var applications []models.Application
		if err := query.Find(&applications).Error; err != nil {
			serverError(c, "Failed to fetch applications")
			return
		}
		attachUnreadCounts(applications)
		c.JSON(http.StatusOK, gin.H{"applications": applications, "total": len(applications)})
		return
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	var total int64
	if err := query.Session(&gorm.Session{}).Model(&models.Application{}).Count(&total).Error; err != nil {
		serverError(c, "Failed to count applications")
		return
	}
	var applications []models.Application
	if err := query.Limit(min(limit, 200)).Offset(max(offset, 0)).Find(&applications).Error; err != nil {
		serverError(c, "Failed to fetch applications")
		return
	}
	attachUnreadCounts(applications)
	c.JSON(http.StatusOK, gin.H{"applications": applications, "total": total})
}

// attachUnreadCounts fills UnreadMessages for a page of applications with one grouped query.
func attachUnreadCounts(apps []models.Application) {
	if len(apps) == 0 {
		return
	}
	ids := make([]uuid.UUID, len(apps))
	for i := range apps {
		ids[i] = apps[i].ID
	}
	var rows []struct {
		ApplicationID uuid.UUID
		Count         int
	}
	config.DB.Model(&models.Message{}).Select("application_id, count(*) as count").
		Where("application_id IN ? AND sender_type = ? AND NOT is_read", ids, models.SenderCandidate).
		Group("application_id").Scan(&rows)
	counts := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		counts[r.ApplicationID] = r.Count
	}
	for i := range apps {
		apps[i].UnreadMessages = counts[apps[i].ID]
	}
}

// GetUnreadSummary powers the dashboard bell: unread candidate messages in total and per job.
func GetUnreadSummary(c *gin.Context) {
	var rows []struct {
		JobID uuid.UUID
		Count int
	}
	err := config.DB.Table("messages m").
		Select("a.job_id as job_id, count(*) as count").
		Joins("join applications a on a.id = m.application_id").
		Where("a.company_id = ? AND a.job_id IS NOT NULL AND m.sender_type = ? AND NOT m.is_read", companyID(c), models.SenderCandidate).
		Group("a.job_id").Scan(&rows).Error
	if err != nil {
		serverError(c, "Failed to load unread messages")
		return
	}
	byJob, total := make(map[string]int, len(rows)), 0
	for _, r := range rows {
		byJob[r.JobID.String()] = r.Count
		total += r.Count
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "by_job": byJob})
}

func ShortlistApplication(c *gin.Context) {
	app, ok := loadApplicationParam(c)
	if !ok {
		return
	}
	if !models.Undecided(app.Status) {
		c.JSON(http.StatusConflict, gin.H{"error": "Only undecided applications can be shortlisted"})
		return
	}
	changeStatus(c, app, models.StatusShortlisted, services.SendShortlistEmail)
}

func RejectApplication(c *gin.Context) {
	app, ok := loadApplicationParam(c)
	if !ok {
		return
	}
	// Repeating a decision would re-send the candidate email and SMS.
	if app.Status == models.StatusRejected || app.Status == models.StatusHired {
		c.JSON(http.StatusConflict, gin.H{"error": "Application is already " + app.Status})
		return
	}
	changeStatus(c, app, models.StatusRejected, services.SendRejectionEmail)
}

// AdvanceApplication moves a shortlisted candidate one stage on: interview, then offer, then hired.
func AdvanceApplication(c *gin.Context) {
	app, ok := loadApplicationParam(c)
	if !ok {
		return
	}
	next, ok := models.NextStage[app.Status]
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Application can't be advanced from " + app.Status})
		return
	}
	changeStatus(c, app, next, func(to, name, jobTitle, applicationID string) error {
		return services.SendStageEmail(to, name, jobTitle, applicationID, next)
	})
}

func loadApplicationParam(c *gin.Context) (*models.Application, bool) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return nil, false
	}
	return findCompanyApplication(c, id)
}

// changeStatus records the move, logs it and notifies the candidate. reviewed_at marks the
// shortlist/reject decision; later stages only bump last_status_update.
func changeStatus(c *gin.Context, app *models.Application, status string, notify func(to, name, jobTitle, applicationID string) error) {
	now := time.Now()
	reviewer := adminID(c)
	oldStatus := app.Status
	app.Status = status
	app.LastStatusUpdate = &now
	updates := map[string]any{"status": status, "last_status_update": now}
	if status == models.StatusShortlisted || status == models.StatusRejected {
		app.ReviewedAt, app.ReviewedBy = &now, &reviewer
		updates["reviewed_at"], updates["reviewed_by"] = now, reviewer
	}
	if status != models.StatusRejected && status != models.StatusHired {
		app.ExpectedResponseDate = ptr(now.Add(expectedResponseWindow))
		updates["expected_response_date"] = app.ExpectedResponseDate
	}
	if err := config.DB.Model(app).Updates(updates).Error; err != nil {
		serverError(c, "Failed to update application")
		return
	}

	jobTitle := app.JobTitle()
	services.LogApplicationStatusChanged(app.CompanyID, reviewer, app.ID, app.FullName, jobTitle, oldStatus, status)
	go func() {
		if err := notify(app.Email, app.FullName, jobTitle, app.ID.String()); err != nil {
			log.Printf("%s email for application %s: %v", status, app.ID, err)
		}
		if err := services.SendStatusUpdateSMS(app.Phone, app.FullName, jobTitle, status); err != nil {
			log.Printf("%s SMS for application %s: %v", status, app.ID, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Application " + status, "application": app})
}

// ReopenApplication undoes any decision or stage move without notifying the candidate.
func ReopenApplication(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}
	if models.Undecided(app.Status) {
		c.JSON(http.StatusConflict, gin.H{"error": "Application is already under review"})
		return
	}

	now := time.Now()
	oldStatus := app.Status
	app.Status = models.StatusPending
	if app.CVViewedAt != nil {
		app.Status = models.StatusCVViewed
	}
	app.ReviewedAt, app.ReviewedBy, app.LastStatusUpdate = nil, nil, &now
	app.ExpectedResponseDate = ptr(now.Add(expectedResponseWindow))
	updates := map[string]any{
		"status": app.Status, "reviewed_at": nil, "reviewed_by": nil,
		"last_status_update": now, "expected_response_date": app.ExpectedResponseDate,
	}
	if err := config.DB.Model(app).Updates(updates).Error; err != nil {
		serverError(c, "Failed to reopen application")
		return
	}

	services.LogApplicationStatusChanged(app.CompanyID, adminID(c), app.ID, app.FullName, app.JobTitle(), oldStatus, app.Status)
	c.JSON(http.StatusOK, gin.H{"message": "Application reopened", "application": app})
}

// GetApplicationFile returns a 5-minute signed link to an application's stored CV or portfolio.
// The link is only issued after the company check, and expires before it can usefully leak.
func GetApplicationFile(c *gin.Context) {
	app, ok := loadApplicationParam(c)
	if !ok {
		return
	}
	bucket, path := services.BucketResumes, app.ResumePath
	label := "CV"
	if c.Param("kind") == "portfolio" {
		bucket, path, label = services.BucketPortfolios, app.PortfolioPath, "Portfolio"
	} else if c.Param("kind") != "resume" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unknown file"})
		return
	}
	if path == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No uploaded file"})
		return
	}
	// PDFs open in the browser's viewer; anything else downloads under a readable name.
	ext := strings.ToLower(filepath.Ext(*path))
	download := ""
	if ext != ".pdf" {
		download = safeFileName(strings.ReplaceAll(app.FullName, " ", "_") + "_" + label + ext)
	}
	link, err := services.SignedURL(bucket, *path, 5*time.Minute, download)
	if err != nil {
		log.Printf("sign %s for application %s: %v", label, app.ID, err)
		serverError(c, "Failed to open file")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"url": link, "expires_in": 300})
}

func DeleteApplication(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}
	if err := config.DB.Delete(app).Error; err != nil {
		serverError(c, "Failed to delete application")
		return
	}
	go services.DeleteApplicationFiles(*app)

	company, admin, jobTitle := app.CompanyID, adminID(c), app.JobTitle()
	services.LogActivity(&company, &admin, "application_deleted", "application", &app.ID,
		"Application deleted: "+app.FullName+" for job: "+jobTitle,
		map[string]any{"applicant_name": app.FullName, "applicant_email": app.Email, "job_title": jobTitle, "status": app.Status})

	c.JSON(http.StatusOK, gin.H{"message": "Application deleted successfully"})
}

func BulkDeleteApplications(c *gin.Context) {
	var req struct {
		Status string     `json:"status" binding:"required,oneof=pending cv_viewed shortlisted interview offer hired rejected"`
		JobID  *uuid.UUID `json:"job_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, err)
		return
	}

	// Scope matches the applications list: live jobs only, narrowed to one job when filtered.
	company, admin := companyID(c), adminID(c)
	query := config.DB.Clauses(clause.Returning{Columns: []clause.Column{
		{Name: "id"}, {Name: "full_name"}, {Name: "email"}, {Name: "resume_path"}, {Name: "portfolio_path"},
	}}).
		Where("company_id = ? AND status = ? AND job_id IS NOT NULL", company, req.Status)
	if req.JobID != nil {
		query = query.Where("job_id = ?", *req.JobID)
	}
	var deleted []models.Application
	err := query.Delete(&deleted).Error
	if err != nil {
		serverError(c, "Failed to delete applications")
		return
	}

	if len(deleted) > 0 {
		go services.DeleteApplicationFiles(deleted...)
		names := make([]string, len(deleted))
		for i, app := range deleted {
			names[i] = app.FullName
		}
		services.LogActivity(&company, &admin, "applications_bulk_deleted", "application", nil,
			fmt.Sprintf("Bulk deleted %d %s application(s)", len(deleted), req.Status),
			map[string]any{"status": req.Status, "count": len(deleted), "applicants": names})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Deleted %d application(s) with status '%s'", len(deleted), req.Status),
		"deleted_count": len(deleted),
		"total_found":   len(deleted),
	})
}

// TrackCVView records the first CV view and moves pending applications to cv_viewed.
func TrackCVView(c *gin.Context) {
	id, ok := uuidParam(c, "id")
	if !ok {
		return
	}
	app, ok := findCompanyApplication(c, id)
	if !ok {
		return
	}

	now := time.Now()
	viewer := adminID(c)
	updates := map[string]any{}
	if app.CVViewedAt == nil {
		app.CVViewedAt, app.CVViewedBy = &now, &viewer
		updates["cv_viewed_at"], updates["cv_viewed_by"] = now, viewer
	}
	oldStatus := app.Status
	if app.Status == models.StatusPending {
		app.Status = models.StatusCVViewed
		app.LastStatusUpdate = &now
		app.ExpectedResponseDate = ptr(now.Add(expectedResponseWindow))
		updates["status"], updates["last_status_update"], updates["expected_response_date"] = app.Status, now, app.ExpectedResponseDate
	}

	if len(updates) > 0 {
		if err := config.DB.Model(app).Updates(updates).Error; err != nil {
			serverError(c, "Failed to update application")
			return
		}
	}

	if oldStatus != app.Status {
		jobTitle := app.JobTitle()
		services.LogApplicationStatusChanged(app.CompanyID, viewer, app.ID, app.FullName, jobTitle, oldStatus, app.Status)
		go func() {
			if err := services.SendStatusUpdateSMS(app.Phone, app.FullName, jobTitle, app.Status); err != nil {
				log.Printf("cv_viewed SMS for application %s: %v", app.ID, err)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{"message": "CV view tracked", "application": app})
}
