package routes

import (
    "onboarding-app/controllers"
    "onboarding-app/middleware"
    
    "github.com/gofiber/fiber/v2"
)

func SetupAdminRoutes(app *fiber.App) {
    // Admin route group
    admin := app.Group("/admin", middleware.RequireAuth())
    
    // Dashboard
    admin.Get("/", controllers.AdminDashboard)
    
    // Submissions
    submissions := admin.Group("/submissions")
    submissions.Get("/", controllers.GetAllSubmissions)
    submissions.Get("/:id", controllers.GetSubmission)
    submissions.Put("/:id", controllers.UpdateSubmission)
}