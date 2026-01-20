package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/middleware"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/service"
	"github.com/user/video-downloader-backend/pkg/response"
	"github.com/user/video-downloader-backend/pkg/utils"
)

type DownloadHandler struct {
	svc     service.DownloadService
	userSvc service.UserService
}

func NewDownloadHandler(svc service.DownloadService, userSvc service.UserService) *DownloadHandler {
	return &DownloadHandler{svc: svc, userSvc: userSvc}
}

func (h *DownloadHandler) DownloadVideo(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req model.DownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		return response.Error(c, fiber.StatusBadRequest, response.ValidationErrors{Errors: errs}.Error(), nil)
	}

	if req.URL == "" {
		return response.Error(c, fiber.StatusBadRequest, "URL is required", nil)
	}

	log.Info().
		Str("url", req.URL).
		Str("type", req.Type).
		Msg("Received download request")

	var userID *uuid.UUID
	if req.UserID != nil {
		id, err := uuid.Parse(*req.UserID)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
		}
		user, err := h.userSvc.FindByID(ctx, id)
		if err != nil {
			log.Error().Err(err).Str("user_id", id.String()).Msg("Failed to find user")
			return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
		}
		userID = &user.ID
	}

	ip := c.IP()

	start := time.Now()

	result, err := h.svc.ProcessDownload(ctx, req, userID, ip)
	if err != nil {
		log.Error().Err(err).Str("url", req.URL).Msg("Failed to process download request")
		return response.Error(c, fiber.StatusInternalServerError, "Failed to process download", err.Error())
	}

	log.Info().
		Str("url", req.URL).
		Str("task_id", result.ID.String()).
		Dur("processing_time", time.Since(start)).
		Msg("Download request processed successfully")

	event := &model.DownloadEvent{
		Type:      "download.queued",
		TaskID:    result.ID,
		UserID:    result.UserID,
		Status:    "queued",
		CreatedAt: time.Now(),
	}

	log.Info().
		Str("type", event.Type).
		Str("task_id", event.TaskID.String()).
		Str("status", event.Status).
		Msg("Broadcasting initial queued download event")

	go func(e *model.DownloadEvent) {
		defaultDownloadEventHub.Broadcast(e)
	}(event)

	return response.Success(c, "Download processed successfully", result)
}

func (h *DownloadHandler) GetHistory(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)
	// Must be authenticated
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	history, err := h.svc.GetUserHistory(ctx, userID, page, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch history", err.Error())
	}

	return response.Success(c, "User download history", history)
}

func (h *DownloadHandler) FindByID(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid ID", err.Error())
	}

	task, err := h.svc.FindByID(ctx, id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch download", err.Error())
	}

	return response.Success(c, "Download fetched successfully", task)
}

func (h *DownloadHandler) GetDownloads(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	params := model.QueryParamsRequest{
		Page:    c.QueryInt("page", 1),
		Limit:   c.QueryInt("limit", 10),
		Search:  c.Query("search"),
		SortBy:  c.Query("sort_by"),
		OrderBy: c.Query("order_by"),
		Status:  c.Query("status"),
		UserID:  c.Query("user_id"),
	}

	result, err := h.svc.FindAll(ctx, params)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch downloads", err.Error())
	}
	return response.Success(c, "Downloads fetched successfully", result)
}

func (h *DownloadHandler) UpdateDownload(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid ID", err.Error())
	}

	var task model.DownloadTask
	if err := c.BodyParser(&task); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.svc.Update(ctx, id, &task); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to update download", err.Error())
	}
	return response.Success(c, "Download updated successfully", nil)
}

func (h *DownloadHandler) DeleteDownload(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid ID", err.Error())
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to delete download", err.Error())
	}
	return response.Success(c, "Download deleted successfully", nil)
}

func (h *DownloadHandler) BulkDeleteDownloads(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	var req struct {
		IDs []uuid.UUID `json:"ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if len(req.IDs) == 0 {
		return response.Error(c, fiber.StatusBadRequest, "No IDs provided", nil)
	}

	if err := h.svc.BulkDelete(ctx, req.IDs); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to bulk delete downloads", err.Error())
	}
	return response.Success(c, "Downloads deleted successfully", nil)
}

func (h *DownloadHandler) ProxyDownload(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)

	taskIDStr := c.Query("task_id")
	formatID := c.Query("format_id")
	filename := c.Query("filename")

	// New flow: download based on task_id using yt-dlp streaming
	if taskIDStr != "" {
		id, err := uuid.Parse(taskIDStr)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid task ID", err.Error())
		}

		task, err := h.svc.FindByID(ctx, id)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch download task", err.Error())
		}
		if task == nil {
			return response.Error(c, fiber.StatusNotFound, "Download task not found", nil)
		}

		pageURL := task.OriginalURL
		if pageURL == "" {
			return response.Error(c, fiber.StatusBadRequest, "Original URL is missing for this task", nil)
		}

		if filename == "" {
			if task.Title != nil && *task.Title != "" {
				filename = *task.Title
			} else {
				filename = "download"
			}
		}

		if !strings.HasSuffix(strings.ToLower(filename), ".mp4") {
			filename += ".mp4"
		}

		filename = strings.ReplaceAll(filename, `"`, `\"`)
		encodedFilename := url.QueryEscape(filename)

		// Don't set headers yet, wait until we have a successful download
		// c.Set("Content-Type", "video/mp4")
		// c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))

		isTikTok := strings.Contains(strings.ToLower(pageURL), "tiktok.com")
		isRumble := strings.Contains(strings.ToLower(pageURL), "rumble.com")
		isVimeo := strings.Contains(strings.ToLower(pageURL), "vimeo.com")

		args := []string{
			"--js-runtimes", "node",
			"--no-playlist",
			"--merge-output-format", "mp4",
			"--no-check-certificate",
			// Always set a browser User-Agent.
			// For Vimeo, this is crucial because the Docker container's yt-dlp fails to impersonate automatically.
			"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
			// Set socket timeout to prevent hanging on connection issues
			"--socket-timeout", "30",
		}

		var targetURL string

		// Improved Vimeo Strategy: Use FilePath from DB if available (contains hash), otherwise construct Player URL
		if isVimeo {
			// Priority 1: Use FilePath from DB if it's a Player URL (might contain access tokens/hashes like ?h=...)
			if task.FilePath != nil && strings.Contains(*task.FilePath, "player.vimeo.com") {
				targetURL = *task.FilePath
				// Unescape HTML entities (e.g., &amp; -> &) which might have been saved from raw HTML
				targetURL = html.UnescapeString(targetURL)
				log.Info().Str("target_url", targetURL).Msg("Using Vimeo Player URL from DB Task")
			} else {
				// Priority 2: Construct Player URL from ID
				// Regex to find the numeric ID
				re := regexp.MustCompile(`vimeo\.com/(?:channels/(?:\w+/)?|groups/[^/]+/videos/|video/|)(\d+)`)
				matches := re.FindStringSubmatch(pageURL)
				if len(matches) > 1 {
					vimeoID := matches[1]
					targetURL = fmt.Sprintf("https://player.vimeo.com/video/%s", vimeoID)
					log.Info().Str("vimeo_id", vimeoID).Msg("Constructed Vimeo Player URL for download")
				} else {
					// Fallback to pageURL if we can't parse ID
					targetURL = pageURL
				}
			}
		} else {
			targetURL = pageURL
			if task.FilePath != nil {
				log.Debug().Str("file_path", *task.FilePath).Msg("Task FilePath from DB")
			}

			// Use FilePath as targetURL if it's a valid direct URL and different from pageURL
			// This helps for Rumble and other sites where we pre-resolved the direct link
			if task.FilePath != nil && *task.FilePath != "" &&
				strings.HasPrefix(*task.FilePath, "http") &&
				*task.FilePath != pageURL {

				// For TikTok, we only use FilePath if it DOES NOT contain public-proxy
				// Because the issue "This site can't be reached" often happens when we try to download from our own proxy recursively
				if isTikTok && strings.Contains(*task.FilePath, "/api/v1/public-proxy/") {
					log.Warn().Str("task_id", task.ID.String()).Msg("Ignoring recursive proxy URL for TikTok, using pageURL")
					targetURL = pageURL
				} else {
					targetURL = *task.FilePath
				}
			}
		}

		// Prevent recursion if FilePath somehow contains the proxy URL itself
		if task.FilePath != nil && strings.Contains(*task.FilePath, "/api/v1/public-proxy/") {
			log.Error().Str("task_id", task.ID.String()).Str("file_path", *task.FilePath).Msg("Recursive proxy URL detected in task FilePath, ignoring")
			// Reset targetURL to OriginalURL to avoid loop, though it might fail if OriginalURL is protected
			targetURL = pageURL
		}

		// Check if we should use lux strategy
		if task.FilePath != nil && strings.HasPrefix(*task.FilePath, "lux://") {
			targetURL = strings.TrimPrefix(*task.FilePath, "lux://")
			tmpDir := os.TempDir()
			filenameBase := task.ID.String()

			log.Info().Str("target_url", targetURL).Msg("Executing Lux download strategy")

			// lux -o tmpDir -O filenameBase targetURL
			cmd := exec.CommandContext(ctx, "lux", "-o", tmpDir, "-O", filenameBase, targetURL)
			if err := cmd.Run(); err != nil {
				log.Error().Err(err).Msg("Lux download failed")
				return response.Error(c, fiber.StatusInternalServerError, "Lux download failed", err.Error())
			}

			// Find the file. lux might add extension.
			matches, err := filepath.Glob(filepath.Join(tmpDir, filenameBase+".*"))
			if err != nil || len(matches) == 0 {
				return response.Error(c, fiber.StatusInternalServerError, "Downloaded file not found", nil)
			}
			filePath := matches[0]
			log.Info().Str("file_path", filePath).Msg("Lux download successful, serving file")

			return c.SendFile(filePath)
		}

		if isTikTok {
			// For TikTok, ensure we use a clean User-Agent and referer to avoid anti-bot issues
			args = append(args, "--referer", "https://www.tiktok.com/")
			// Remove the generic User-Agent added above, let yt-dlp handle it or use a specific one if needed
			// But since we appended it to args slice, we can't easily remove it.
			// Instead, we can override it by adding another user-agent flag if yt-dlp supports last-wins,
			// or we should structure the args construction better.
			// For now, let's just add extra headers.
		}

		if isRumble {
			args = append(args, "--referer", "https://rumble.com/")
			args = append(args, "--add-header", "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
			args = append(args, "--add-header", "Accept-Language: en-US,en;q=0.9")
		}

		if isVimeo {
			args = append(args, "--referer", "https://vimeo.com/")
			// Add headers mimicking a real browser navigation to the player
			args = append(args, "--add-header", "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			args = append(args, "--add-header", "Accept-Language: en-us,en;q=0.5")
			args = append(args, "--add-header", "Sec-Fetch-Mode: navigate")
		}

		effectiveFormatID := strings.TrimSpace(formatID)
		if effectiveFormatID == "" || strings.EqualFold(effectiveFormatID, "download") {
			effectiveFormatID = ""
		}

		if effectiveFormatID != "" {
			args = append(args, "-f", effectiveFormatID)
		} else {
			if isTikTok {
				args = append(args, "-f", "best[ext=mp4][vcodec=h264]/best[ext=mp4]")
			} else if isVimeo {
				// Vimeo often requires merging video+audio streams
				args = append(args, "-f", "bestvideo+bestaudio/best")
			} else {
				args = append(args, "-f", "best[ext=mp4][vcodec=h264]/best[ext=mp4]/best")
			}
		}

		// Optimization: Try to get direct stream URL for Vimeo using Custom Strategy FIRST
		// This avoids yt-dlp's slow download-then-serve mechanism and ffmpeg merging
		if isVimeo {
			log.Info().Str("target_url", targetURL).Msg("Attempting Vimeo direct stream optimization")
			vimeoStrategy := infrastructure.NewVimeoStrategy()
			// Use targetURL which is the Player URL
			if directURL, err := vimeoStrategy.GetDirectURL(ctx, targetURL); err == nil && directURL != "" {
				log.Info().Str("direct_url", directURL).Msg("Vimeo direct stream found, proxying to client")

				req, err := http.NewRequestWithContext(ctx, "GET", directURL, nil)
				if err == nil {
					req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

					// Use the strategy's client which has proper timeout configuration
					resp, err := vimeoStrategy.Client.Do(req)
					if err == nil && resp.StatusCode == http.StatusOK {
						defer resp.Body.Close()

						// Set headers for the client response
						c.Set("Content-Type", "video/mp4")
						if resp.ContentLength > 0 {
							c.Set("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
						}
						c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))

						// Stream the body directly
						return c.Status(http.StatusOK).SendStream(resp.Body)
					}
				}
			} else {
				log.Warn().Err(err).Msg("Vimeo direct stream optimization failed, falling back to yt-dlp")
			}
		}

		tempFile := filepath.Join(os.TempDir(), "download-"+uuid.New().String()+".mp4")

		// Capture base args before appending output and targetURL for fallback
		baseArgs := make([]string, len(args))
		copy(baseArgs, args)

		args = append(args, "-o", tempFile, targetURL)

		cmd := exec.CommandContext(ctx, "yt-dlp", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			log.Error().
				Err(err).
				Str("url", targetURL).
				Str("stderr", stderr.String()).
				Msg("yt-dlp download failed on targetURL")

			// Fallback 1: Direct HTTP stream (if targetURL is http)
			// ... (logic below)

			// Fallback 2: If targetURL was NOT pageURL, try downloading from pageURL (OriginalURL)
			// This handles cases where the direct link (targetURL) is expired (403 Forbidden)
			// For Vimeo, this is critical as the player URL often fails with 403 or XML errors
			if targetURL != pageURL {
				log.Info().Str("url", pageURL).Msg("Falling back to yt-dlp on OriginalURL (pageURL)")

				// Re-run yt-dlp with pageURL
				// Ensure we include referer for Vimeo fallback
				fallbackArgs := append(baseArgs, "-o", tempFile, pageURL)

				// Explicitly ensure referer is present for Vimeo fallback if not already in baseArgs
				// (baseArgs should have it, but let's be safe for the fallback command)
				if isVimeo {
					// Check if referer is already in baseArgs (simple check)
					hasReferer := false
					for _, arg := range baseArgs {
						if strings.Contains(arg, "vimeo.com") && strings.Contains(arg, "--referer") {
							hasReferer = true
							break
						}
					}
					if !hasReferer {
						fallbackArgs = append(fallbackArgs, "--referer", "https://vimeo.com/")
					}
				}

				cmd2 := exec.CommandContext(ctx, "yt-dlp", fallbackArgs...)
				var stderr2 bytes.Buffer
				cmd2.Stderr = &stderr2

				if err2 := cmd2.Run(); err2 == nil {
					// Success! Continue to serve file
					log.Info().Msg("yt-dlp fallback on OriginalURL succeeded")
					goto ServeFile
				} else {
					log.Error().Err(err2).Str("stderr", stderr2.String()).Msg("yt-dlp fallback on OriginalURL failed")
				}
			}

			// Fallback 3: Lux strategy for Vimeo
			// If yt-dlp fails for Vimeo, try using lux as a last resort
			if isVimeo {
				log.Info().Str("url", pageURL).Msg("Falling back to Lux for Vimeo")

				// Lux usage: lux -o <output_dir> -O <output_filename> <url>
				// Note: lux doesn't support specifying exact filename with extension in -O, it appends extension automatically.
				// But we can rename it later or find it.
				luxFilenameBase := "lux-" + uuid.New().String()
				luxArgs := []string{"-o", os.TempDir(), "-O", luxFilenameBase, pageURL}

				luxCmd := exec.CommandContext(ctx, "lux", luxArgs...)
				if luxErr := luxCmd.Run(); luxErr == nil {
					log.Info().Msg("Lux fallback succeeded")

					// Find the file. lux adds extension.
					matches, err := filepath.Glob(filepath.Join(os.TempDir(), luxFilenameBase+".*"))
					if err == nil && len(matches) > 0 {
						tempFile = matches[0] // Update tempFile to point to the lux downloaded file
						goto ServeFile
					} else {
						log.Error().Msg("Lux succeeded but file not found")
					}
				} else {
					log.Error().Err(luxErr).Msg("Lux fallback failed")
				}
			}

			// Fallback 4: Custom Vimeo Strategy (Golang implementation)
			// This strategy parses the player page for the config object and extracts direct MP4 links
			if isVimeo {
				log.Info().Str("url", targetURL).Msg("Falling back to Custom Vimeo Strategy")
				vimeoStrategy := infrastructure.NewVimeoStrategy()
				vimeoFilename := "vimeo-custom-" + uuid.New().String() + ".mp4"
				vimeoFilePath := filepath.Join(os.TempDir(), vimeoFilename)

				// Use targetURL which is likely the player URL
				if err := vimeoStrategy.Download(ctx, targetURL, vimeoFilePath); err == nil {
					log.Info().Msg("Custom Vimeo Strategy succeeded")
					tempFile = vimeoFilePath
					goto ServeFile
				} else {
					log.Error().Err(err).Msg("Custom Vimeo Strategy failed")
				}
			}

			// Fallback: If yt-dlp failed and we have a direct http URL (from Chromedp/Lux),
			// try to stream it directly using http.Client
			// We also try fallback if the targetURL looks like a direct file (ends in .mp4) even if it equals pageURL
			shouldFallback := (targetURL != pageURL && strings.HasPrefix(targetURL, "http")) ||
				(strings.HasSuffix(strings.ToLower(targetURL), ".mp4") && strings.HasPrefix(targetURL, "http"))

			// Exclude known non-video pages from direct stream fallback to prevent serving HTML as video
			if strings.Contains(targetURL, "player.vimeo.com") {
				log.Warn().Str("target_url", targetURL).Msg("Skipping direct HTTP fallback for Vimeo player URL")
				shouldFallback = false
			}

			if shouldFallback {
				log.Info().Str("target_url", targetURL).Msg("Falling back to direct HTTP stream")
				req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
				if err != nil {
					return response.Error(c, fiber.StatusInternalServerError, "Failed to create fallback request", err.Error())
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36")
				// Pass existing headers? Maybe Referer
				if isRumble {
					req.Header.Set("Referer", "https://rumble.com/")
				}

				client := &http.Client{Timeout: 0}
				resp, err := client.Do(req)
				if err != nil {
					return response.Error(c, fiber.StatusBadGateway, "Failed to fetch remote file (fallback)", err.Error())
				}

				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
					resp.Body.Close()
					// Try one last time with pageURL if we haven't already
					// Actually, we already tried pageURL above if targetURL != pageURL.
					return response.Error(c, resp.StatusCode, "Remote server returned error (fallback)", nil)
				}

				// Success with direct stream
				c.Set("Content-Type", resp.Header.Get("Content-Type"))
				c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))
				if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
					c.Set("Content-Length", contentLength)
				}

				// Wrap body to ensure it closes on EOF
				reader := &autoCloseReader{ReadCloser: resp.Body}
				return c.SendStream(reader)
			}

			return response.Error(c, fiber.StatusBadGateway, "Failed to download video", err.Error())
		}

	ServeFile:
		file, err := os.Open(tempFile)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Failed to open downloaded video", err.Error())
		}
		// Jangan hapus file di defer, biarkan sistem operasi yang membersihkan atau gunakan cron job
		// defer func() {
		// 	_ = file.Close()
		// 	_ = os.Remove(tempFile)
		// }()

		if fi, err := file.Stat(); err == nil {
			c.Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
		}

		c.Set("Content-Type", "video/mp4")
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))

		return c.SendStream(file)
	}

	// Fallback: simple HTTP proxy based on direct file URL (non-task based)
	urlStr := c.Query("url")

	if urlStr == "" {
		return response.Error(c, fiber.StatusBadRequest, "URL is required", nil)
	}

	urlStr = strings.Trim(urlStr, "`")
	urlStr = strings.ReplaceAll(urlStr, "`", "")

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to create request", err.Error())
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")

	client := &http.Client{
		Timeout: 0,
	}

	resp, err := client.Do(req)
	if err != nil {
		return response.Error(c, fiber.StatusBadGateway, "Failed to fetch remote file", err.Error())
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return response.Error(c, resp.StatusCode, "Remote server returned error", nil)
	}

	c.Status(resp.StatusCode)
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Set("Content-Type", contentType)
	if ar := resp.Header.Get("Accept-Ranges"); ar != "" {
		c.Set("Accept-Ranges", ar)
	}
	if cr := resp.Header.Get("Content-Range"); cr != "" {
		c.Set("Content-Range", cr)
	}

	if filename == "" {
		filename = "download.mp4"
	}

	filename = strings.ReplaceAll(filename, `"`, `\"`)
	encodedFilename := url.QueryEscape(filename)

	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		c.Set("Content-Length", contentLength)
	}

	return c.SendStream(resp.Body)
}

type DownloadEventHub struct {
	mu          sync.RWMutex
	clients     map[*websocket.Conn]struct{}
	userClients map[uuid.UUID]map[*websocket.Conn]struct{}
}

func NewDownloadEventHub() *DownloadEventHub {
	return &DownloadEventHub{
		clients:     make(map[*websocket.Conn]struct{}),
		userClients: make(map[uuid.UUID]map[*websocket.Conn]struct{}),
	}
}

func (h *DownloadHandler) DownloadEvents(c *websocket.Conn) {
	defaultDownloadEventHub.Add(c)
	defer defaultDownloadEventHub.Remove(c)

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *DownloadHandler) DownloadEventsByUser(c *websocket.Conn) {
	userIDParam := c.Params("user_id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		_ = c.Close()
		return
	}

	defaultDownloadEventHub.AddForUser(userID, c)
	defer defaultDownloadEventHub.RemoveForUser(userID, c)

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *DownloadEventHub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
	log.Info().
		Int("total_clients", len(h.clients)).
		Msg("Websocket client connected to anonymous download events")
}

func (h *DownloadEventHub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	total := len(h.clients)
	h.mu.Unlock()
	log.Info().
		Int("total_clients", total).
		Msg("Websocket client disconnected from anonymous download events")
}

func (h *DownloadEventHub) AddForUser(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	conns, ok := h.userClients[userID]
	if !ok {
		conns = make(map[*websocket.Conn]struct{})
		h.userClients[userID] = conns
	}
	conns[conn] = struct{}{}
	h.mu.Unlock()
	log.Info().
		Str("user_id", userID.String()).
		Int("user_clients", len(conns)).
		Msg("Websocket client connected to user-specific download events")
}

func (h *DownloadEventHub) RemoveForUser(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	if conns, ok := h.userClients[userID]; ok {
		delete(conns, conn)
		remaining := len(conns)
		if len(conns) == 0 {
			delete(h.userClients, userID)
		}
		log.Info().
			Str("user_id", userID.String()).
			Int("user_clients_remaining", remaining).
			Msg("Websocket client disconnected from user-specific download events")
	}
	h.mu.Unlock()
}

type autoCloseReader struct {
	io.ReadCloser
}

func (a *autoCloseReader) Read(p []byte) (int, error) {
	n, err := a.ReadCloser.Read(p)
	if err == io.EOF {
		_ = a.ReadCloser.Close()
	}
	return n, err
}

func (h *DownloadEventHub) Broadcast(event *model.DownloadEvent) {
	if event == nil {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	totalAnon := len(h.clients)
	anonConns := make([]*websocket.Conn, 0, totalAnon)
	for conn := range h.clients {
		anonConns = append(anonConns, conn)
	}

	totalUser := 0
	var userConns []*websocket.Conn
	if event.UserID != nil {
		if conns, ok := h.userClients[*event.UserID]; ok {
			totalUser = len(conns)
			userConns = make([]*websocket.Conn, 0, len(conns))
			for conn := range conns {
				userConns = append(userConns, conn)
			}
		}
	}
	h.mu.RUnlock()

	log.Info().
		Str("type", event.Type).
		Str("task_id", event.TaskID.String()).
		Str("status", event.Status).
		Int("anon_clients", totalAnon).
		Int("user_clients", totalUser).
		Msg("Broadcasting download event to websocket clients")

	for _, conn := range anonConns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			go h.Remove(conn)
		}
	}

	if event.UserID != nil {
		for _, conn := range userConns {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				go h.RemoveForUser(*event.UserID, conn)
			}
		}
	}
}

var defaultDownloadEventHub = NewDownloadEventHub()

func BroadcastDownloadEvent(event *model.DownloadEvent) {
	defaultDownloadEventHub.Broadcast(event)
}
