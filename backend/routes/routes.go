package routes

import (
	"ats-backend/controllers"
	"ats-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Public routes (no auth required)
		api.POST("/auth/register", controllers.Register)
		api.POST("/auth/login", controllers.Login)
		api.GET("/jobs/public/:companyId", controllers.GetPublicJobs)
		api.POST("/applications", controllers.SubmitApplication)
		
		// Candidate Portal routes (public)
		api.POST("/candidate/status", controllers.GetApplicationStatus)
		api.GET("/candidate/applications", controllers.GetApplicationStatusByEmail)
		api.POST("/candidate/messages/send", controllers.SendMessage)
		api.GET("/candidate/messages", controllers.GetMessages)
		
		// File upload routes (public for application submission)
		api.POST("/upload/cv", controllers.UploadCV)
		api.POST("/upload/portfolio", controllers.UploadPortfolio)

		// Super Admin routes (login only - no public registration)
		api.POST("/super-admin/login", controllers.SuperAdminLogin)

		// Protected routes (company admin auth required)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Job routes
			protected.POST("/jobs", controllers.CreateJob)
			protected.GET("/jobs", controllers.GetJobs)
			protected.GET("/jobs/:id", controllers.GetJob)
			protected.PUT("/jobs/:id", controllers.UpdateJob)
			protected.DELETE("/jobs/:id", controllers.DeleteJob)

			// Application routes
			protected.GET("/applications", controllers.GetApplications)
			protected.PUT("/applications/:id/shortlist", controllers.ShortlistApplication)
			protected.PUT("/applications/:id/reject", controllers.RejectApplication)
			protected.POST("/applications/:id/track-cv-view", controllers.TrackCVView)
			protected.DELETE("/applications/:id", controllers.DeleteApplication)
			protected.POST("/applications/bulk-delete", controllers.BulkDeleteApplications)
			
			// Messaging routes (protected)
			protected.POST("/applications/:id/messages", controllers.SendMessage)
			protected.GET("/applications/:id/messages", controllers.GetMessagesForRecruiter)
			
			// AI Shortlisting routes
			protected.POST("/applications/ai-shortlist", controllers.AIShortlistApplication)
			protected.POST("/applications/ai-shortlist-batch", controllers.BatchAIShortlist)
			
			// Activity Logs routes
			protected.GET("/activity-logs", controllers.GetActivityLogs)
			
			// Candidate Search routes
			protected.POST("/candidates/search", controllers.SearchCandidates)
			protected.GET("/candidates/:id", controllers.GetCandidateDetails)
			
			// Manual Candidate routes
			protected.POST("/candidates/manual", controllers.AddManualCandidate)
			
			// CV Reparsing routes (for fixing existing applications)
			protected.POST("/candidates/reparse-all", controllers.ReparseAllCVs)
			protected.POST("/candidates/:id/reparse", controllers.ReparseSingleCV)
			
			// CRM routes
			protected.POST("/crm/notes", controllers.AddCandidateNote)
			protected.GET("/crm/applications/:id/notes", controllers.GetCandidateNotes)
			protected.PUT("/crm/notes/:id", controllers.UpdateCandidateNote)
			protected.DELETE("/crm/notes/:id", controllers.DeleteCandidateNote)
			// Talent pool routes - GET must come before DELETE with :id parameter
			protected.GET("/crm/talent-pool", controllers.GetTalentPool)
			protected.POST("/crm/talent-pool", controllers.AddToTalentPool)
			protected.DELETE("/crm/talent-pool/:id", controllers.RemoveFromTalentPool)
			protected.PUT("/crm/applications/:id/referral", controllers.UpdateReferralInfo)
			protected.GET("/crm/applications/:id/timeline", controllers.GetRelationshipTimeline)
		}

		// Super Admin protected routes
		superAdmin := api.Group("/super-admin")
		superAdmin.Use(middleware.SuperAdminAuthMiddleware())
		{
			superAdmin.GET("/stats", controllers.GetSuperAdminStats)
			superAdmin.GET("/companies", controllers.GetAllCompanies)
			superAdmin.GET("/activity-logs", controllers.GetSuperAdminActivityLogs)
		}
	}
}

