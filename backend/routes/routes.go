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
			
			// AI Shortlisting routes
			protected.POST("/applications/ai-shortlist", controllers.AIShortlistApplication)
			protected.POST("/applications/ai-shortlist-batch", controllers.BatchAIShortlist)
			
			// Activity Logs routes
			protected.GET("/activity-logs", controllers.GetActivityLogs)
			
			// Candidate Search routes
			protected.POST("/candidates/search", controllers.SearchCandidates)
			protected.GET("/candidates/:id", controllers.GetCandidateDetails)
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

