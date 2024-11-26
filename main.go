// main.go
package main

import (
    "log"
    "os"
    "onboarding-app/config"
    "onboarding-app/controllers"
    "onboarding-app/middleware"
    "onboarding-app/utils"
    "onboarding-app/services"  // Add services import

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/gofiber/template/html/v2"
    "github.com/joho/godotenv"
)

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: Error loading .env file")
    }

    // Connect to database
    config.ConnectDB()
    defer config.DisconnectDB()

    // Initialize template engine
    engine := html.New("./templates", ".html")
    engine.Debug(true)
    // Register template functions
    engine.AddFunc("statusBadgeClass", utils.GetStatusBadgeClass)
    engine.AddFunc("subtract", utils.Subtract)
    engine.AddFunc("add", utils.Add)
    engine.AddFunc("formatDate", utils.FormatDate)
    engine.AddFunc("formatNumber", utils.FormatNumber)
    engine.AddFunc("formatMoney", utils.FormatMoney)    
    engine.AddFunc("truncate", utils.Truncate)
    
    // For development, disable template caching
    engine.Reload(true)

    // Initialize Fiber
    app := fiber.New(fiber.Config{
        Views: engine,
    })

    // Add middleware
    setupMiddleware(app)

    // S3 Configuration
    s3Config := services.S3Config{
        AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
        SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
        Region:         os.Getenv("AWS_REGION"),
        Bucket:         os.Getenv("AWS_BUCKET_NAME"),
        Endpoint:       os.Getenv("AWS_ENDPOINT"),
    }

    // Setup routes with S3 config
    if err := setupRoutes(app, s3Config); err != nil {
        log.Fatalf("Failed to setup routes: %v", err)
    }

    // Get port from environment variable or use default
    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }

    log.Fatal(app.Listen(":" + port))
}

// setupMiddleware adds global middleware to the app
func setupMiddleware(app *fiber.App) {
    app.Use(logger.New())
    app.Use(recover.New())
}

func setupRoutes(app *fiber.App, s3Config services.S3Config) error {
    // Initialize controllers
    authController := controllers.NewAuthController()
    
    formController, err := controllers.NewFormController(s3Config)
    if err != nil {
        return err
    }
    
    adminController, err := controllers.NewAdminController(s3Config)
    if err != nil {
        return err
    }

    documentController, err := controllers.NewDocumentController(s3Config)
    if err != nil {
        return err
    }

    // Auth routes (unprotected)
    app.Get("/login", authController.ShowLoginPage)
    app.Post("/auth/login", authController.HandleLogin)

    // Form routes
    app.Get("/", formController.ShowForm)
    app.Post("/submit", formController.SubmitForm)
    // app.Post("/validate/lei", formController.ValidateLEI)
    app.Post("/upload", middleware.LimitFileSize(10), formController.HandleFileUpload)

    // Admin routes protected by auth middleware
    admin := app.Group("/admin", middleware.RequireAuth())
    admin.Get("/", adminController.AdminDashboard)
    admin.Get("/submissions", adminController.GetAllSubmissions)
    admin.Get("/submission/:id/json", adminController.GetSubmissionJSON)  // API endpoint
    admin.Get("/submission/:id", adminController.ViewSubmission) 
    admin.Get("/submission/:id", adminController.GetSubmission)
    admin.Put("/submission/:id", adminController.UpdateSubmission)
    admin.Put("/submission/:id/status", adminController.UpdateSubmissionStatus)
    admin.Post("/submission/:id/notes", adminController.AddSubmissionNote)

    // Document routes
    admin.Get("/document/:key", documentController.HandleDownload)
    admin.Get("/document/:key/download", documentController.HandleDownload)

    return nil
}