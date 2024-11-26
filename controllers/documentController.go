package controllers

import (
    "context"
    "fmt"
    "time"
    "path/filepath"
    "onboarding-app/services"
    "github.com/gofiber/fiber/v2"
)

type DocumentController struct {
    fileService *services.S3FileService
}

func NewDocumentController(s3Config services.S3Config) (*DocumentController, error) {
    fileService, err := services.NewS3FileService(s3Config)
    if err != nil {
        return nil, err
    }

    return &DocumentController{
        fileService: fileService,
    }, nil
}

// HandleDownload handles document downloads
func (dc *DocumentController) HandleDownload(c *fiber.Ctx) error {
    key := c.Params("key")
    if key == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "No document key provided",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Check if file exists first
    exists, err := dc.fileService.CheckIfFileExists(ctx, key)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Error checking file",
            "details": err.Error(),
        })
    }

    if !exists {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "File not found",
        })
    }

    // If redirect parameter is true, return presigned URL
    if c.Query("redirect") == "true" {
        presignedURL, err := dc.fileService.GetPresignedURL(ctx, key)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Failed to generate download URL",
                "details": err.Error(),
            })
        }
        return c.Redirect(presignedURL)
    }

    // Otherwise stream the file
    fileStream, contentType, err := dc.fileService.DownloadFile(ctx, key)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to download file",
            "details": err.Error(),
        })
    }
    defer fileStream.Close()

    // Set appropriate headers for download
    filename := filepath.Base(key)
    c.Set("Content-Type", contentType)
    c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

    return c.SendStream(fileStream)
}