package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/user/video-downloader-backend/pkg/response"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	data := fiber.Map{
		"database": "up",
		"redis":    "up",
		"time":     time.Now(),
	}

	hasError := false

	if err := h.db.Ping(ctx); err != nil {
		data["database"] = "down"
		hasError = true
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		data["redis"] = "down"
		hasError = true
	}

	if hasError {
		return response.Error(c, fiber.StatusServiceUnavailable, "System is degraded", data)
	}

	return response.Success(c, "System is healthy", data)
}

func (h *HealthHandler) GetLogger(c *fiber.Ctx) error {
	_, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	logPath := "logs/logs.json"
	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return response.SuccessWithMeta(c, "No logs found", []interface{}{}, fiber.Map{
				"current_page": page,
				"limit":        limit,
				"total_items":  0,
				"total_pages":  0,
				"has_next":     false,
				"has_prev":     false,
			})
		}
		return response.Error(c, fiber.StatusInternalServerError, "Failed to open log file", err.Error())
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	// Increase buffer for large log lines (e.g. stack traces)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Error reading log file", err.Error())
	}

	totalItems := len(lines)
	totalPages := 0
	if limit > 0 {
		totalPages = (totalItems + limit - 1) / limit
	}

	// Calculate start and end indices for reversed list (newest first)
	// Page 1: last 'limit' items
	startIndex := totalItems - (page * limit)
	endIndex := totalItems - ((page - 1) * limit)

	if endIndex > totalItems {
		endIndex = totalItems
	}
	if endIndex < 0 {
		endIndex = 0
	}
	if startIndex < 0 {
		startIndex = 0
	}

	var logs []map[string]interface{}

	// Read backwards from endIndex-1 down to startIndex
	for i := endIndex - 1; i >= startIndex; i-- {
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(lines[i]), &entry); err == nil {
			logs = append(logs, entry)
		}
	}

	if logs == nil {
		logs = []map[string]interface{}{}
	}

	meta := fiber.Map{
		"current_page": page,
		"limit":        limit,
		"total_items":  totalItems,
		"total_pages":  totalPages,
		"has_next":     page < totalPages,
		"has_prev":     page > 1,
	}

	return response.SuccessWithMeta(c, "Logs retrieved successfully", logs, meta)
}

func (h *HealthHandler) ClearLogs(c *fiber.Ctx) error {
	_, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	logPath := "logs/logs.json"

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return response.Success(c, "Logs cleared successfully (file did not exist)", nil)
	}

	if err := os.Truncate(logPath, 0); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to clear logs", err.Error())
	}

	return response.Success(c, "Logs cleared successfully", nil)
}
