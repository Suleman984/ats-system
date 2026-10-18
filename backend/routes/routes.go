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
	auth.POST("/auth/logout", controllers.Logout)
	auth.POST("/super-admin/logout", controllers.SuperAdminLogout)
	auth.POST("/auth/forgot-password", controllers.ForgotPassword)
	auth.POST("/auth/reset-password", controllers.ResetPassword)
	auth.POST("/auth/verify-email", controllers.VerifyEmail)
	auth.POST("/auth/verify-email/resend", controllers.ResendVerification)
	auth.GET("/auth/invites", controllers.GetInvite)
	auth.POST("/auth/invites/accept", controllers.AcceptInvite)

	// Signature-checked, and only triggers a status re-check with Safepay.
	api.POST("/webhooks/safepay", middleware.RateLimit(120, time.Minute), controllers.SafepayWebhook)

	public := api.Group("", middleware.RateLimit(60, time.Minute))
	public.GET("/jobs/public/:companyId", controllers.GetPublicJobs)
	public.GET("/public/jobs/:id", controllers.GetPublicJob)
	public.POST("/applications", controllers.SubmitApplication)
	public.POST("/upload/cv", controllers.UploadCV)
	public.POST("/upload/portfolio", controllers.UploadPortfolio)
	public.POST("/candidate/status", controllers.GetApplicationStatus)
	public.POST("/candidate/messages/send", controllers.SendCandidateMessage)
	public.POST("/candidate/messages", controllers.GetCandidateMessages)
	public.POST("/candidate/stream-token", controllers.CandidateStreamToken)
	public.GET("/candidate/events", controllers.CandidateChatEvents)

	// Generous per-IP cap; several protected routes fan out to CV downloads.
	protected := api.Group("", middleware.AuthMiddleware(), middleware.RateLimit(300, time.Minute))
	protected.POST("/jobs", controllers.CreateJob)
	protected.GET("/jobs", controllers.GetJobs)
	protected.GET("/jobs/:id", controllers.GetJob)
	protected.PUT("/jobs/:id", controllers.UpdateJob)
	protected.GET("/auth/me", controllers.GetMe)

	owner := protected.Group("", middleware.RequireOwner())
	owner.DELETE("/jobs/:id", controllers.DeleteJob)
	owner.POST("/applications/bulk-delete", controllers.BulkDeleteApplications)
	owner.GET("/team", controllers.GetTeam)
	owner.POST("/team/invites", controllers.InviteTeamMember)
	owner.DELETE("/team/invites/:id", controllers.RevokeTeamInvite)
	owner.PUT("/team/members/:id/role", controllers.UpdateTeamMemberRole)
	owner.DELETE("/team/members/:id", controllers.RemoveTeamMember)
	owner.POST("/billing/checkout", controllers.StartCheckout)
	protected.GET("/billing", controllers.GetBilling)
	protected.POST("/billing/payments/:id/confirm", controllers.ConfirmPayment)

	protected.GET("/dashboard/stats", controllers.GetDashboardStats)
	protected.GET("/messages/unread-summary", controllers.GetUnreadSummary)
	protected.GET("/events", controllers.ChatEvents)
	protected.GET("/applications", controllers.GetApplications)
	protected.PUT("/applications/:id/shortlist", controllers.ShortlistApplication)
	protected.PUT("/applications/:id/reject", controllers.RejectApplication)
	protected.PUT("/applications/:id/advance", controllers.AdvanceApplication)
	protected.PUT("/applications/:id/reopen", controllers.ReopenApplication)
	protected.POST("/applications/:id/track-cv-view", controllers.TrackCVView)
	protected.GET("/applications/:id/files/:kind", controllers.GetApplicationFile)
	protected.DELETE("/applications/:id", controllers.DeleteApplication)
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
	superAdmin.PUT("/companies/:id", controllers.UpdateCompanySubscription)
	superAdmin.GET("/activity-logs", controllers.GetSuperAdminActivityLogs)
}
