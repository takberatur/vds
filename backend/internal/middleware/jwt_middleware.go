package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/response"
)

func JWTMiddleware(tokenService service.TokenService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized, "Authorization header missing", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid authorization header format", nil)
		}
		tokenString := parts[1]

		token, err := tokenService.ValidateToken(tokenString)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid or expired token", err.Error())
		}
		if !token.Valid {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid token", nil)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid token claims", nil)
		}

		if sub, ok := claims["sub"].(string); ok {
			if userID, err := uuid.Parse(sub); err == nil {
				c.Locals("user_id", userID)
			} else {
				return response.Error(c, fiber.StatusUnauthorized, "Invalid token subject format", nil)
			}
		} else {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid token subject", nil)
		}

		if email, ok := claims["email"].(string); ok {
			c.Locals("email", email)
		}

		if role, ok := claims["role"].(string); ok {
			c.Locals("role_id", role)
		}

		if roleName, ok := claims["role_name"].(string); ok {
			c.Locals("role_name", roleName)
		}

		return c.Next()
	}
}
