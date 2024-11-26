// middleware/file.go

package middleware

import (
    "github.com/gofiber/fiber/v2"
)

func LimitFileSize(sizeMB int64) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Get the content length from the header
        size := int64(c.Request().Header.ContentLength())
        
        // Convert MB to bytes
        maxSize := sizeMB * 1024 * 1024
        
        if size > maxSize {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "File size exceeds the limit",
            })
        }
        
        return c.Next()
    }
}