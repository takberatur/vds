package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/user/video-downloader-backend/internal/service"
)

// OptionalJWTMiddleware attempts to extract user info but proceeds even if token is missing/invalid
func OptionalJWTMiddleware(tokenService service.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}
		tokenString := parts[1]

		token, err := tokenService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			return c.Next() // Invalid token, treat as guest
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if ok {
			if sub, ok := claims["sub"].(string); ok {
				if userID, err := uuid.Parse(sub); err == nil {
					c.Locals("user_id", userID)
				}
			}
		}

		return c.Next()
	}
}
