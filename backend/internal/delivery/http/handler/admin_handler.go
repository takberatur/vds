package handler

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/logger"
	"github.com/user/video-downloader-backend/pkg/response"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}
func (h *AdminHandler) GetDashboardData(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	params := model.QueryParamsRequest{}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dateFrom); err == nil {
			params.DateFrom = t
		}
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse(time.RFC3339, dateTo); err == nil {
			params.DateTo = t
		}
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	params.Limit = limit

	resp, err := h.adminService.GetDashboardData(ctx, params)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to get dashboard data", err.Error())
	}

	return response.SuccessWithMeta(c, "Dashboard data retrieved successfully", resp.Data, resp.Pagination)
}
func (h *AdminHandler) GetCookies(c *fiber.Ctx) error {
	cookiesPath := getCookiesFilePath()

	b, err := os.ReadFile(cookiesPath)
	if err != nil {
		if os.IsNotExist(err) {
			data := map[string]any{
				"path":  cookiesPath,
				"lines": defaultNetscapeHeaderLines(),
				"valid": false,
			}
			return response.Success(c, "Cookies file not found; returning template", data)
		}
		return response.Error(c, fiber.StatusInternalServerError, "Failed to read cookies file", err.Error())
	}

	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	raw := string(b)
	raw = strings.TrimRight(raw, "\n") + "\n"
	lines := splitLinesPreserve(raw)

	data := map[string]any{
		"path":  cookiesPath,
		"lines": lines,
		"valid": infrastructure.IsValidNetscapeCookiesFile(cookiesPath),
	}

	return response.Success(c, "Cookies file retrieved successfully", data)
}
func (h *AdminHandler) UpdateCookies(c *fiber.Ctx) error {
	start := time.Now()
	actor := adminActorIdentity(c)
	var req struct {
		Cookies []string `json:"cookies"`
		Content string   `json:"content"`
	}
	if err := c.BodyParser(&req); err != nil {
		logger.NotifyTelegram("[cookies] UpdateCookies invalid payload actor=%s ip=%s err=%s", actor, c.IP(), err.Error())
		return response.Error(c, fiber.StatusBadRequest, "Invalid request payload", err.Error())
	}

	content := strings.TrimSpace(req.Content)
	var lines []string
	if content != "" {
		lines = splitLinesPreserve(content + "\n")
	} else {
		for _, l := range req.Cookies {
			lines = append(lines, strings.TrimRight(l, "\r\n"))
		}
	}

	if len(lines) == 0 {
		logger.NotifyTelegram("[cookies] UpdateCookies empty content actor=%s ip=%s", actor, c.IP())
		return response.Error(c, fiber.StatusBadRequest, "Cookies content is required", nil)
	}

	normalizedLines := normalizeToNetscape(lines)
	contentOut := strings.Join(normalizedLines, "\n")
	contentOut = strings.ReplaceAll(contentOut, "\r\n", "\n")
	contentOut = strings.TrimLeft(contentOut, "\ufeff")
	if !strings.HasSuffix(contentOut, "\n") {
		contentOut += "\n"
	}

	cookiesPath := getCookiesFilePath()
	dir := filepath.Dir(cookiesPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.NotifyTelegram("[cookies] UpdateCookies mkdir failed actor=%s path=%s ip=%s err=%s", actor, cookiesPath, c.IP(), err.Error())
		return response.Error(c, fiber.StatusInternalServerError, "Failed to prepare cookies directory", err.Error())
	}

	tmpPath := cookiesPath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		logger.NotifyTelegram("[cookies] UpdateCookies open tmp failed actor=%s path=%s ip=%s err=%s", actor, tmpPath, c.IP(), err.Error())
		return response.Error(c, fiber.StatusInternalServerError, "Failed to open temp cookies file", err.Error())
	}
	_, werr := f.WriteString(contentOut)
	cerr := f.Close()
	if werr != nil {
		_ = os.Remove(tmpPath)
		logger.NotifyTelegram("[cookies] UpdateCookies write tmp failed actor=%s path=%s ip=%s err=%s", actor, tmpPath, c.IP(), werr.Error())
		return response.Error(c, fiber.StatusInternalServerError, "Failed to write cookies file", werr.Error())
	}
	if cerr != nil {
		_ = os.Remove(tmpPath)
		logger.NotifyTelegram("[cookies] UpdateCookies close tmp failed actor=%s path=%s ip=%s err=%s", actor, tmpPath, c.IP(), cerr.Error())
		return response.Error(c, fiber.StatusInternalServerError, "Failed to write cookies file", cerr.Error())
	}

	if err := os.Rename(tmpPath, cookiesPath); err != nil {
		writeDirect := func() error {
			tf, e := os.OpenFile(cookiesPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
			if e != nil {
				return e
			}
			_, we := tf.WriteString(contentOut)
			ce := tf.Close()
			if we != nil {
				return we
			}
			return ce
		}

		tryDirect := false
		if errors.Is(err, syscall.EBUSY) || errors.Is(err, syscall.EXDEV) || strings.Contains(strings.ToLower(err.Error()), "device or resource busy") {
			tryDirect = true
		}
		if !tryDirect {
			_ = os.Remove(tmpPath)
			logger.NotifyTelegram("[cookies] UpdateCookies rename failed actor=%s path=%s ip=%s err=%s", actor, cookiesPath, c.IP(), err.Error())
			return response.Error(c, fiber.StatusInternalServerError, "Failed to replace cookies file", err.Error())
		}

		_ = os.Remove(tmpPath)
		if derr := writeDirect(); derr != nil {
			logger.NotifyTelegram("[cookies] UpdateCookies direct write failed actor=%s path=%s ip=%s err=%s", actor, cookiesPath, c.IP(), derr.Error())
			return response.Error(c, fiber.StatusInternalServerError, "Failed to replace cookies file", derr.Error())
		}
	}

	valid := infrastructure.IsValidNetscapeCookiesFile(cookiesPath)
	if st, statErr := os.Stat(cookiesPath); statErr == nil {
		dur := time.Since(start)
		logger.NotifyTelegram(
			"[cookies] updated actor=%s valid=%t size=%dB path=%s ip=%s dur=%s",
			actor,
			valid,
			st.Size(),
			cookiesPath,
			c.IP(),
			dur.Truncate(time.Millisecond).String(),
		)
	} else {
		logger.NotifyTelegram("[cookies] updated stat_failed actor=%s valid=%t path=%s ip=%s err=%s", actor, valid, cookiesPath, c.IP(), fmt.Sprint(statErr))
	}

	data := map[string]any{
		"path":  cookiesPath,
		"valid": valid,
	}
	return response.Success(c, "Cookies file updated successfully", data)
}

func adminActorIdentity(c *fiber.Ctx) string {
	userID := "-"
	if v := c.Locals("user_id"); v != nil {
		switch t := v.(type) {
		case uuid.UUID:
			userID = t.String()
		case string:
			userID = t
		default:
			userID = fmt.Sprint(v)
		}
	}

	email := "-"
	if v, ok := c.Locals("email").(string); ok && strings.TrimSpace(v) != "" {
		email = strings.TrimSpace(v)
	}

	role := "-"
	if v, ok := c.Locals("role_name").(string); ok && strings.TrimSpace(v) != "" {
		role = strings.TrimSpace(v)
	}

	apiKey := strings.TrimSpace(c.Get("X-API-Key"))
	apiKeyMasked := "-"
	if apiKey != "" {
		apiKeyMasked = maskSecret(apiKey)
	}

	return fmt.Sprintf("user_id=%s email=%s role=%s api_key=%s", userID, email, role, apiKeyMasked)
}

func maskSecret(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	if len(s) <= 10 {
		return s[:min(3, len(s))] + "..."
	}
	return s[:6] + "..." + s[len(s)-4:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getCookiesFilePath() string {
	if v := strings.TrimSpace(os.Getenv("COOKIES_FILE_PATH")); v != "" {
		return v
	}
	if _, err := os.Stat("/app/cookies.txt"); err == nil {
		return "/app/cookies.txt"
	}
	return "cookies.txt"
}

func defaultNetscapeHeaderLines() []string {
	return []string{
		"# Netscape HTTP Cookie File",
		"# This is a generated file! Do not edit.",
		"",
	}
}

func splitLinesPreserve(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimPrefix(s, "\ufeff")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

func normalizeToNetscape(lines []string) []string {
	out := make([]string, 0, len(lines)+3)
	for _, l := range lines {
		l = strings.TrimRight(l, "\r\n")
		out = append(out, l)
	}

	i := 0
	for i < len(out) && strings.TrimSpace(out[i]) == "" {
		i++
	}
	if i >= len(out) || !strings.HasPrefix(strings.TrimSpace(out[i]), "# Netscape HTTP Cookie File") {
		out = append(defaultNetscapeHeaderLines(), out...)
	} else {
		if strings.TrimPrefix(strings.TrimSpace(out[i]), "# Netscape HTTP Cookie File") != "" {
			out[i] = "# Netscape HTTP Cookie File"
		}
	}

	out2 := make([]string, 0, len(out))
	seenNonEmpty := false
	for _, l := range out {
		t := strings.TrimSpace(l)
		if t == "" && !seenNonEmpty {
			continue
		}
		if t != "" {
			seenNonEmpty = true
		}
		out2 = append(out2, l)
	}

	return out2
}
