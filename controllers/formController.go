// controllers/formController.go
package controllers

import (
    "context"
    "time"
    "onboarding-app/models"
    "onboarding-app/services"
    "github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FormController struct {
    submissionService *services.SubmissionService
}

func NewFormController(s3Config services.S3Config) (*FormController, error) {
    submissionService, err := services.NewSubmissionService(s3Config)
    if err != nil {
        return nil, err
    }

    return &FormController{
        submissionService: submissionService,
    }, nil
}

// ShowForm renders the submission form
func (fc *FormController) ShowForm(c *fiber.Ctx) error {
    return c.Render("form", fiber.Map{
        "Title": "Admission to Trade Application",
    })
}

// SubmitForm handles the form submission
func (fc *FormController) SubmitForm(c *fiber.Ctx) error {
    var submission models.FormSubmission
    if err := c.BodyParser(&submission); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid submission data",
        })
    }

    submission.Status = "pending"
    submission.SubmittedAt = time.Now()
    submission.UpdatedAt = time.Now()

    err := fc.submissionService.Create(&submission)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to save submission",
            "details": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "Form submitted successfully",
        "submission": submission,
    })
}

// HandleFileUpload processes file uploads
func (fc *FormController) HandleFileUpload(c *fiber.Ctx) error {
    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "No file uploaded",
        })
    }

    docType := c.FormValue("type", "general")
    submissionID, err := primitive.ObjectIDFromHex(c.FormValue("submissionId"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid submission ID",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    doc, err := fc.submissionService.AddDocument(ctx, submissionID, file, docType, "system")
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to upload file",
            "details": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "File uploaded successfully",
        "document": doc,
    })
}

// TestSubmission handles test form submissions
func (fc *FormController) TestSubmission(c *fiber.Ctx) error {    
    // Create a test submission
    testSubmission := &models.FormSubmission{
        Status: "draft",
        SubmittedAt: time.Now(),
        UpdatedAt: time.Now(),
        IssuerInfo: models.IssuerInformation{
            CompanyName: "Test Company Ltd",
            TradeName: "TestCo",
            Constitution: "LLC",
            IncorporationDetails: struct {
                Country    string `json:"country" bson:"country"`
                Date      string `json:"date" bson:"date"`
                RegNumber string `json:"reg_number" bson:"reg_number"`
                LEI       string `json:"lei" bson:"lei"`
            }{
                Country: "United States",
                Date: "2020-01-01",
                RegNumber: "12345",
                LEI: "123456789ABCDEFGHIJK",
            },
            Address: struct {
                Registered string `json:"registered" bson:"registered"`
                Corporate  string `json:"corporate" bson:"corporate"`
                Website    string `json:"website" bson:"website"`
            }{
                Registered: "123 Test St, Test City",
                Corporate: "456 Corp Ave, Business City",
                Website: "https://testcompany.com",
            },
            Operations: "Test company operations",
            GroupDescription: "Test group description",
            IndustrySector: "Technology",
            ContactPerson: struct {
                Name     string `json:"name" bson:"name"`
                Position string `json:"position" bson:"position"`
                Phone    string `json:"phone" bson:"phone"`
                Email    string `json:"email" bson:"email"`
            }{
                Name: "John Doe",
                Position: "CEO",
                Phone: "+1234567890",
                Email: "john@testcompany.com",
            },
        },
    }

    err := fc.submissionService.Create(testSubmission)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "Test submission created successfully",
        "submission": testSubmission,
    })
}