package middleware

import (
    "fmt"
    "os"
	"time"
    "strings"
    "github.com/gofiber/fiber/v2"
    "github.com/golang-jwt/jwt/v4"
)

type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func RequireAuth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        var tokenString string

        // Check Authorization header
        authHeader := c.Get("Authorization")
        if authHeader != "" {
            tokenString = strings.TrimPrefix(authHeader, "Bearer ")
            
            // Store token in cookie if it came from header
            cookie := new(fiber.Cookie)
            cookie.Name = "auth_token"
            cookie.Value = tokenString
            cookie.Expires = time.Now().Add(24 * time.Hour)
            cookie.HTTPOnly = true
            cookie.Secure = true // for HTTPS
            cookie.SameSite = "Lax"
            c.Cookie(cookie)
        } else {
            // Try to get token from cookie
            tokenString = c.Cookies("auth_token")
        }

        if tokenString == "" {
            fmt.Println("No token found, redirecting to login")
            return c.Redirect("/login")
        }

        // Validate token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil || !token.Valid {
            fmt.Printf("Token validation failed: %v\n", err)
            // Clear invalid cookie
            c.ClearCookie("auth_token")
            return c.Redirect("/login")
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            return c.Redirect("/login")
        }

        fmt.Printf("Auth successful for user: %v\n", claims["user_id"])
        c.Locals("user", claims)
        return c.Next()
    }
}

func RequireAdmin() fiber.Handler {
    return func(c *fiber.Ctx) error {
        claims, ok := c.Locals("user").(jwt.MapClaims)
        if !ok {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "User not authenticated",
            })
        }

        if claims["role"] != "admin" {
            return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
                "error": "Admin access required",
            })
        }

        return c.Next()
    }
}