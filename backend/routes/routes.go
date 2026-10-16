package routes

import (
	"net/http"
	"time"

	"ats-backend/controllers"
	"ats-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := router.Group("/api")

	auth := api.Group("", middleware.RateLimit(10, time.Minute))
	auth.POST("/auth/register", controllers.Register)
	auth.POST("/auth/login", controllers.Login)
	auth.POST("/super-admin/login", controllers.SuperAdminLogin)

	public := api.Group("", middleware.RateLimit(60, time.Minute))
	public.GET("/jobs/public/:companyId", controllers.GetPublicJobs)
	public.GET("/public/jobs/:id", controllers.GetPublicJob)
	public.POST("/applications", controllers.SubmitApplication)
	public.POST("/upload/cv", controllers.UploadCV)
	public.POST("/upload/portfolio", controllers.UploadPortfolio)
	public.POST("/candidate/status", controllers.GetApplicationStatus)
	public.POST("/candidate/messages/send", controllers.SendCandidateMessage)
	public.GET("/candidate/messages", controllers.GetCandidateMessages)

	protected := api.Group("", middleware.AuthMiddleware())
	protected.POST("/jobs", controllers.CreateJob)
	protected.GET("/jobs", controllers.GetJobs)
	protected.GET("/jobs/:id", controllers.GetJob)
	protected.PUT("/jobs/:id", controllers.UpdateJob)
	protected.DELETE("/jobs/:id", controllers.DeleteJob)

	protected.GET("/applications", controllers.GetApplications)
	protected.PUT("/applications/:id/shortlist", controllers.ShortlistApplication)
	protected.PUT("/applications/:id/reject", controllers.RejectApplication)
	protected.POST("/applications/:id/track-cv-view", controllers.TrackCVView)
	protected.DELETE("/applications/:id", controllers.DeleteApplication)
	protected.POST("/applications/bulk-delete", controllers.BulkDeleteApplications)
	protected.POST("/applications/:id/messages", controllers.SendRecruiterMessage)
	protected.GET("/applications/:id/messages", controllers.GetRecruiterMessages)
	protected.POST("/applications/ai-shortlist", controllers.AIShortlistApplication)
	protected.POST("/applications/ai-shortlist-batch", controllers.BatchAIShortlist)

	protected.GET("/activity-logs", controllers.GetActivityLogs)

	protected.POST("/candidates/search", controllers.SearchCandidates)
	protected.GET("/candidates/:id", controllers.GetCandidateDetails)
	protected.POST("/candidates/manual", controllers.AddManualCandidate)
	protected.POST("/candidates/reparse-all", controllers.ReparseAllCVs)
	protected.POST("/candidates/:id/reparse", controllers.ReparseSingleCV)

	protected.POST("/crm/notes", controllers.AddCandidateNote)
	protected.GET("/crm/applications/:id/notes", controllers.GetCandidateNotes)
	protected.PUT("/crm/notes/:id", controllers.UpdateCandidateNote)
	protected.DELETE("/crm/notes/:id", controllers.DeleteCandidateNote)
	protected.GET("/crm/talent-pool", controllers.GetTalentPool)
	protected.POST("/crm/talent-pool", controllers.AddToTalentPool)
	protected.DELETE("/crm/talent-pool/:id", controllers.RemoveFromTalentPool)
	protected.PUT("/crm/applications/:id/referral", controllers.UpdateReferralInfo)
	protected.GET("/crm/applications/:id/timeline", controllers.GetRelationshipTimeline)

	superAdmin := api.Group("/super-admin", middleware.SuperAdminAuthMiddleware())
	superAdmin.GET("/stats", controllers.GetSuperAdminStats)
	superAdmin.GET("/companies", controllers.GetAllCompanies)
	superAdmin.GET("/activity-logs", controllers.GetSuperAdminActivityLogs)
}
