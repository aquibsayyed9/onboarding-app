package controllers

import (
    "context"
    "strconv"
    "time"
	"fmt"
    "onboarding-app/models"
    "onboarding-app/services"
    "github.com/gofiber/fiber/v2"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/bson"
)

type AdminController struct {
    submissionService *services.SubmissionService
}

func NewAdminController(s3Config services.S3Config) (*AdminController, error) {
    submissionService, err := services.NewSubmissionService(s3Config)
    if err != nil {
        return nil, err
    }

    return &AdminController{
        submissionService: submissionService,
    }, nil
}

func (ac *AdminController) AdminDashboard(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    stats, err := ac.submissionService.GetStats(ctx)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to get statistics",
            "details": err.Error(),
        })
    }

    // Set up data for the dashboard
    dashboardData := fiber.Map{
        "Title": "Admin Dashboard",
        "Stats": stats,
        "RecentActivity": fiber.Map{
            "Recent24h": stats.RecentCount,
            "Pending":   stats.PendingCount,
        },
        "StatusBreakdown": stats.StatusCounts,
    }

    return c.Render("admin/dashboard", dashboardData)
}

func (ac *AdminController) GetStats(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    stats, err := ac.submissionService.GetStats(ctx)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to retrieve stats",
            "details": err.Error(),
        })
    }

    // If it's an HTMX request, return partial template
    if c.Get("HX-Request") == "true" {
        return c.Render("admin/partials/stats", fiber.Map{
            "Stats": stats,
        })
    }

    // For API requests, return JSON
    return c.JSON(fiber.Map{
        "stats": stats,
    })
}

func (ac *AdminController) GetAllSubmissions(c *fiber.Ctx) error {

	if c.Method() == "POST" {
        token := c.FormValue("token")
        if token != "" {
            cookie := new(fiber.Cookie)
            cookie.Name = "token"
            cookie.Value = token
            cookie.Expires = time.Now().Add(24 * time.Hour)
            c.Cookie(cookie)
            return c.Redirect("/admin/submissions")
        }
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Get page number, default to 1 if not provided
    pageStr := c.Query("page", "1")
    pageNum, err := strconv.Atoi(pageStr)
    if err != nil {
        pageNum = 1
    }

    // Get limit, default to 10 if not provided
    limitStr := c.Query("limit", "10")
    limitNum, err := strconv.Atoi(limitStr)
    if err != nil {
        limitNum = 10
    }

    // Get sort order, default to -1 if not provided
    orderStr := c.Query("order", "-1")
    orderNum, err := strconv.Atoi(orderStr)
    if err != nil {
        orderNum = -1
    }

    filter := services.SubmissionFilter{
        Page:      pageNum,
        Limit:     limitNum,
        SortField: c.Query("sort", "submitted_at"),
        SortOrder: orderNum,
        Status:    c.Query("status", ""),
        Search:    c.Query("search", ""),
    }

    result, err := ac.submissionService.GetSubmissions(ctx, filter)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to retrieve submissions",
            "details": err.Error(),
        })
    }

    // For HTMX partial updates
    if c.Get("HX-Request") == "true" {
        return c.Render("admin/partials/submissions-table", fiber.Map{
            "Data": result,
        })
    }

    // For full page load
    return c.Render("admin/submissions", fiber.Map{
        "Title": "Submissions",
        "Data": result,
        "Filters": fiber.Map{
            "Status": filter.Status,
            "Search": filter.Search,
            "Sort": filter.SortField,
            "Order": filter.SortOrder,
        },
        "Pagination": fiber.Map{
            "CurrentPage": filter.Page,
            "TotalPages": result.TotalPages,
            "HasNext": filter.Page < result.TotalPages,
            "HasPrev": filter.Page > 1,
        },
    })
}

func (ac *AdminController) GetSubmission(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    id, err := primitive.ObjectIDFromHex(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid submission ID",
        })
    }

    submission, err := ac.submissionService.GetSubmissionByID(ctx, id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to retrieve submission",
        })
    }

    return c.JSON(submission)
}

func (ac *AdminController) UpdateSubmission(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    id, err := primitive.ObjectIDFromHex(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid ID format",
        })
    }

    var updateData struct {
        Status string `json:"status"`
        Notes  string `json:"notes"`
    }

    if err := c.BodyParser(&updateData); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid update data",
        })
    }

    // Create update fields
    update := bson.M{
        "$set": bson.M{
            "status": updateData.Status,
            "updated_at": time.Now(),
        },
    }

    // Add note if provided
    if updateData.Notes != "" {
        note := models.Note{
            Content:   updateData.Notes,
            CreatedAt: time.Now(),
            CreatedBy: c.Get("X-User-ID", "system"),
        }
        update["$push"] = bson.M{"notes": note}
    }

    err = ac.submissionService.Update(ctx, id, update)
    if err != nil {
        if err == services.ErrSubmissionNotFound {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Submission not found",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to update submission",
            "details": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "Submission updated successfully",
    })
}

func (ac *AdminController) UpdateSubmissionStatus(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    id, err := primitive.ObjectIDFromHex(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid ID format",
        })
    }

    var statusUpdate struct {
        Status string `json:"status"`
    }

    if err := c.BodyParser(&statusUpdate); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid status data",
        })
    }

    update := bson.M{
        "$set": bson.M{
            "status":     statusUpdate.Status,
            "updated_at": time.Now(),
        },
    }

    err = ac.submissionService.Update(ctx, id, update)
    if err != nil {
        if err == services.ErrSubmissionNotFound {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Submission not found",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to update status",
            "details": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "Status updated successfully",
    })
}

func (ac *AdminController) AddSubmissionNote(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    id, err := primitive.ObjectIDFromHex(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid ID format",
        })
    }

    var noteData struct {
        Content string `json:"content"`
    }

    if err := c.BodyParser(&noteData); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid note data",
        })
    }

    note := models.Note{
        Content:   noteData.Content,
        CreatedAt: time.Now(),
        CreatedBy: c.Get("X-User-ID", "system"),
    }

    update := bson.M{
        "$push": bson.M{"notes": note},
        "$set":  bson.M{"updated_at": time.Now()},
    }

    err = ac.submissionService.Update(ctx, id, update)
    if err != nil {
        if err == services.ErrSubmissionNotFound {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
                "error": "Submission not found",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to add note",
            "details": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "message": "Note added successfully",
        "note":    note,
    })
}

func (ac *AdminController) ViewSubmission(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    fmt.Println("ViewSubmission handler called")

    id, err := primitive.ObjectIDFromHex(c.Params("id"))
    if err != nil {
        fmt.Printf("Invalid ID: %v\n", err)
        return c.Status(fiber.StatusBadRequest).Render("admin/error", fiber.Map{
            "Title": "Error",
            "Message": "Invalid submission ID",
        })
    }

    submission, err := ac.submissionService.GetSubmissionByID(ctx, id)
    if err != nil {
        fmt.Printf("Error fetching submission: %v\n", err)
        if err == services.ErrSubmissionNotFound {
            return c.Status(fiber.StatusNotFound).Render("admin/error", fiber.Map{
                "Title": "Not Found",
                "Message": "Submission not found",
            })
        }
        return c.Status(fiber.StatusInternalServerError).Render("admin/error", fiber.Map{
            "Title": "Error",
            "Message": "Failed to retrieve submission",
        })
    }

    fmt.Printf("Rendering template for submission ID: %s\n", id.Hex())

    // Force content type to be HTML
    c.Set("Content-Type", "text/html")
    
    return c.Render("admin/submission-view", fiber.Map{
        "Title": "View Submission",
        "Submission": submission,
    })
}

func (ac *AdminController) GetSubmissionJSON(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    id, err := primitive.ObjectIDFromHex(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid submission ID",
        })
    }

    submission, err := ac.submissionService.GetSubmissionByID(ctx, id)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to retrieve submission",
        })
    }

    return c.JSON(submission)
}