package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/video-downloader-backend/pkg/response"
)

func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleName, ok := c.Locals("role_name").(string)
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, "Unauthorized: Role information missing", nil)
		}

		if roleName != "admin" {
			return response.Error(c, fiber.StatusForbidden, "Forbidden: Admin access required", nil)
		}

		return c.Next()
	}
}
