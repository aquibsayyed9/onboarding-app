// controllers/authController.go

package controllers

import (
    "os"
	"time"
    "fmt"
    "github.com/gofiber/fiber/v2"
    "github.com/golang-jwt/jwt/v4"
)

type AuthController struct {}

func NewAuthController() *AuthController {
    return &AuthController{}
}

func (ac *AuthController) ShowLoginPage(c *fiber.Ctx) error {
    return c.Render("login", fiber.Map{})
}

func (ac *AuthController) HandleLogin(c *fiber.Ctx) error {
    username := c.FormValue("username")
    password := c.FormValue("password")

    fmt.Printf("Login attempt - Username: %s\n", username)

    // Get credentials from environment variables
    adminUser := os.Getenv("ADMIN_USERNAME")
    adminPass := os.Getenv("ADMIN_PASSWORD")

    if adminUser == "" || adminPass == "" {
        fmt.Println("Error: Admin credentials not configured")
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Admin credentials not configured",
        })
    }

    // Validate credentials
	if username == adminUser && password == adminPass {
        // Create token
        token := jwt.New(jwt.SigningMethodHS256)
        claims := token.Claims.(jwt.MapClaims)
        claims["user_id"] = username
        claims["role"] = "admin"
        claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

        tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Could not generate token",
            })
        }

        // Set cookie
        cookie := new(fiber.Cookie)
        cookie.Name = "auth_token"
        cookie.Value = tokenString
        cookie.Expires = time.Now().Add(24 * time.Hour)
        cookie.HTTPOnly = true
        cookie.Secure = true // for HTTPS
        cookie.SameSite = "Lax"
        c.Cookie(cookie)

        return c.JSON(fiber.Map{
            "success": true,
            "token": tokenString,
            "redirect": "/admin/submissions",
        })
    }

    fmt.Println("Login failed: Invalid credentials")
    return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
        "error": "Invalid credentials",
    })
}