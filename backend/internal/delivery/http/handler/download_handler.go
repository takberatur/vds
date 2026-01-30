package handler

import (
	"bytes"
	"encoding/json"
	"errors"
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
	"github.com/user/video-downloader-backend/internal/config"
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

func (h *DownloadHandler) DownloadVideoToMp3(c *fiber.Ctx) error {
	return response.Success(c, "Download processed successfully", nil)
}

func (h *DownloadHandler) GetHistory(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)
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
	return response.SuccessWithMeta(c, "Downloads fetched successfully",
		result.Data,
		result.Pagination,
	)
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
	var tempFile string

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

		// Optimization: Check if we have already downloaded files for this task in storage
		// If so, redirect to the storage URL directly instead of re-downloading
		if task.Status == "completed" && len(task.DownloadFiles) > 0 {
			var targetFile *model.DownloadFile

			// If format_id is specified, try to find matching file
			// The worker saves format_id as resolution (e.g. "1080p") or "best"
			if formatID != "" {
				for i := range task.DownloadFiles {
					f := &task.DownloadFiles[i]
					// Check exact match or if formatID matches resolution
					if (f.FormatID != nil && *f.FormatID == formatID) ||
						(f.Resolution != nil && *f.Resolution == formatID) {
						targetFile = f
						break
					}
				}
			}

			// If no specific format found or requested, try to find the "best" or default file
			if targetFile == nil && len(task.DownloadFiles) > 0 {
				// If formatID was requested but not found, we might want to fall back to the best available
				// or continue to re-download.
				// But usually if status is completed, we should have the files.
				// Let's pick the first one which is usually the best quality
				targetFile = &task.DownloadFiles[0]
			}

			if targetFile != nil {
				// Check if file is encrypted
				if targetFile.EncryptedData != nil {
					log.Info().Str("task_id", task.ID.String()).Msg("Found encrypted file in DB, decrypting and streaming")

					cfg := config.LoadConfig()
					decrypted, err := utils.DecryptData(*targetFile.EncryptedData, cfg.EncryptionKey)
					if err != nil {
						log.Error().Err(err).Msg("Failed to decrypt video")
						return response.Error(c, fiber.StatusInternalServerError, "Failed to decrypt video", err.Error())
					}

					// Stream decrypted data
					c.Set("Content-Type", "video/mp4")
					c.Set("Content-Length", fmt.Sprintf("%d", len(decrypted)))

					// Use filename from query or task title
					finalFilename := filename
					if finalFilename == "" {
						if task.Title != nil {
							finalFilename = *task.Title
						} else {
							finalFilename = "download"
						}
					}
					if !strings.HasSuffix(strings.ToLower(finalFilename), ".mp4") {
						finalFilename += ".mp4"
					}
					// Sanitize filename
					finalFilename = strings.ReplaceAll(finalFilename, `"`, `\"`)
					encodedFilename := url.QueryEscape(finalFilename)

					c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, finalFilename, encodedFilename))

					return c.Status(http.StatusOK).SendStream(bytes.NewReader(decrypted))
				}

				if targetFile.URL != "" {
					finalURL := targetFile.URL

					// Clean finalURL from dirty characters (backticks, spaces) that might be in DB from old tasks
					finalURL = strings.TrimSpace(finalURL)
					finalURL = strings.Trim(finalURL, "`")
					finalURL = strings.ReplaceAll(finalURL, "`", "")
					finalURL = strings.Trim(finalURL, "'")
					finalURL = strings.Trim(finalURL, "\"")
					// Trim spaces again in case they were inside quotes/backticks
					finalURL = strings.TrimSpace(finalURL)

					// Prevent recursive redirects if finalURL is the proxy URL itself
					// This fixes ERR_UNSAFE_REDIRECT when the DB contains a URL pointing to this proxy
					if strings.Contains(finalURL, "/api/v1/public-proxy/") {
						log.Warn().
							Str("task_id", task.ID.String()).
							Str("url", finalURL).
							Msg("Recursive proxy URL detected in completed task, skipping redirect optimization to avoid loop")
						// Fall through to normal processing
					} else {
						log.Info().
							Str("task_id", task.ID.String()).
							Str("url", finalURL).
							Str("original_url", targetFile.URL).
							Str("format_id", formatID).
							Msg("Found existing file in storage, redirecting")

						return c.Redirect(finalURL)
					}
				}
			}
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
		isDailymotion := strings.Contains(strings.ToLower(pageURL), "dailymotion.com") || strings.Contains(strings.ToLower(pageURL), "dai.ly")

		args := []string{
			"--js-runtimes", "node",
			"--no-playlist",
			"--merge-output-format", "mp4",
			"--no-check-certificate",
			// Set socket timeout to prevent hanging on connection issues
			"--socket-timeout", "30",
		}

		// Set User-Agent for most platforms, but skip for Dailymotion if it causes 403
		// For Vimeo, this is crucial because the Docker container's yt-dlp fails to impersonate automatically.
		if !isDailymotion {
			args = append(args, "--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36")
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
				} else if isDailymotion && strings.Contains(strings.ToLower(*task.FilePath), ".m3u8") {
					log.Warn().Str("task_id", task.ID.String()).Msg("Ignoring direct m3u8 FilePath for Dailymotion, using pageURL")
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

		// Clean targetURL from dirty characters (backticks, spaces) that might be in DB
		targetURL = strings.TrimSpace(targetURL)
		targetURL = strings.Trim(targetURL, "`")
		targetURL = strings.ReplaceAll(targetURL, "`", "")
		targetURL = strings.Trim(targetURL, "'")
		targetURL = strings.Trim(targetURL, "\"")

		// Check if we should use lux strategy
		// Skip Lux for Dailymotion as it tends to download HTML files instead of video
		if task.FilePath != nil && strings.HasPrefix(*task.FilePath, "lux://") && !isDailymotion {
			targetURL = strings.TrimPrefix(*task.FilePath, "lux://")
			tmpDir := os.TempDir()
			filenameBase := task.ID.String()

			log.Info().Str("target_url", targetURL).Msg("Executing Lux download strategy")

			// lux -o tmpDir -O filenameBase targetURL
			cmd := exec.CommandContext(ctx, "lux", "-o", tmpDir, "-O", filenameBase, targetURL)
			if err := cmd.Run(); err != nil {
				log.Warn().Err(err).Msg("Lux download failed, falling back to yt-dlp")
				// Fall through to yt-dlp
			} else {
				// Find the file. lux might add extension.
				matches, err := filepath.Glob(filepath.Join(tmpDir, filenameBase+".*"))
				if err != nil || len(matches) == 0 {
					log.Warn().Msg("Lux downloaded file not found, falling back to yt-dlp")
					// Fall through
				} else {
					// Prioritize .mp4 or actual video files, avoid .m3u8 or .xml
					foundVideo := false
					var selectedFile string
					for _, m := range matches {
						ext := strings.ToLower(filepath.Ext(m))
						if ext == ".mp4" || ext == ".mkv" || ext == ".webm" {
							selectedFile = m
							foundVideo = true
							break
						}
					}

					shouldServe := true
					if !foundVideo {
						// If no video extension found, check if the first match is NOT a text file
						firstMatch := matches[0]
						ext := strings.ToLower(filepath.Ext(firstMatch))

						if ext == ".m3u8" {
							log.Warn().Str("file", firstMatch).Msg("Lux downloaded m3u8 playlist, falling back to yt-dlp for proper processing")
							shouldServe = false
						} else if ext == ".xml" || ext == ".txt" || ext == ".html" {
							log.Warn().Str("file", firstMatch).Msg("Lux downloaded a non-video file, falling back to yt-dlp")
							shouldServe = false
						} else {
							selectedFile = firstMatch
						}
					}

					if shouldServe {
						tempFile = selectedFile
						log.Info().Str("file_path", tempFile).Msg("Lux download successful, serving file")

						// Serve the file directly to avoid goto scope issues
						if _, err := os.Stat(tempFile); err != nil {
							log.Warn().Err(err).Msg("Lux downloaded file vanished, falling back to yt-dlp")
						} else {
							c.Set("Content-Type", "video/mp4")
							c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))
							return c.SendFile(tempFile)
						}
					}
				}
			}
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

		if isDailymotion {
			// For Dailymotion, prefer using the actual video page URL as referer if available
			if pageURL != "" {
				args = append(args, "--referer", pageURL)
			} else {
				args = append(args, "--referer", "https://www.dailymotion.com/")
			}
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
			} else if isDailymotion {
				args = append(args, "-f", "best[ext=mp4]/best")
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

		{
			// Capture base args before appending output and targetURL for fallback
			baseArgs := make([]string, len(args))
			copy(baseArgs, args)

			if isDailymotion {
				outboundProxy := strings.TrimSpace(os.Getenv("OUTBOUND_PROXY_URL"))
				outboundProxy = strings.TrimSpace(strings.Trim(strings.Trim(strings.Trim(outboundProxy, "`"), "\""), "'"))
				imp := strings.TrimSpace(os.Getenv("YTDLP_IMPERSONATE"))
				if imp == "" {
					imp = "chrome"
				}

				tempFile, err := os.CreateTemp("", "dailymotion-*.mp4")
				if err != nil {
					log.Error().Err(err).Msg("Failed to create temp file")
					return response.Error(c, fiber.StatusInternalServerError, "Failed to create temp file", err.Error())
				}
				tempPath := tempFile.Name()
				tempFile.Close()

				defer func() {
					if err := os.Remove(tempPath); err != nil {
						log.Warn().Err(err).Str("path", tempPath).Msg("Failed to remove temp file")
					}
				}()

				quickArgs := []string{
					"--no-warnings",
					"--no-playlist",
					"--force-overwrites",
					"--no-part",
					"--merge-output-format", "mp4",
					"-o", tempPath,
					"--referer", targetURL,
					"--impersonate", imp,
				}
				if outboundProxy != "" {
					quickArgs = append(quickArgs, "--proxy", outboundProxy)
				}
				if infrastructure.IsValidNetscapeCookiesFile("/app/cookies.txt") && !strings.EqualFold(strings.TrimSpace(os.Getenv("DISABLE_COOKIES_FILE")), "true") {
					quickArgs = append(quickArgs, "--cookies", "/app/cookies.txt")
				}
				quickArgs = append(quickArgs, targetURL)

				cmd := exec.CommandContext(ctx, "yt-dlp", quickArgs...)
				var quickStderr bytes.Buffer
				cmd.Stderr = &quickStderr
				log.Info().Msg("Trying yt-dlp Dailymotion download (impersonate)")
				if err := cmd.Run(); err == nil {
					if fi, err := os.Stat(tempPath); err == nil && fi.Size() > 0 {
						c.Set("Content-Type", "video/mp4")
						c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))
						return c.SendFile(tempPath)
					}
				} else {
					log.Warn().Err(err).Str("stderr", quickStderr.String()).Msg("yt-dlp Dailymotion (impersonate) failed, falling back to Chromedp interception")
				}

				log.Info().Str("url", targetURL).Msg("Streaming Dailymotion video via Chromedp interception")
				chromedpStrategy := infrastructure.NewChromedpStrategy()
				m3u8URL, cookieFilePath, ua, err := chromedpStrategy.GetMasterPlaylist(ctx, targetURL)
				if err != nil {
					log.Error().Err(err).Msg("Failed to resolve stream via Chromedp")
					return response.Error(c, fiber.StatusInternalServerError, "Failed to resolve video stream: "+err.Error(), err.Error())
				}

				// Clean up cookie file after request
				if cookieFilePath != "" {
					defer func() {
						os.Remove(cookieFilePath)
					}()
				}

				// Create temp file for download
				// We download to a file instead of streaming to avoid ERR_EMPTY_RESPONSE issues with unstable pipes
				tempFile2, err := os.CreateTemp("", "dailymotion-*.mp4")
				if err != nil {
					log.Error().Err(err).Msg("Failed to create temp file")
					return response.Error(c, fiber.StatusInternalServerError, "Failed to create temp file", err.Error())
				}
				tempPath2 := tempFile2.Name()
				tempFile2.Close() // Close so yt-dlp can write to it

				// Clean up temp file after request completes
				// We use a closure to capture the current tempPath
				cleanupPaths := []string{tempPath2}
				defer func() {
					for _, p := range cleanupPaths {
						if err := os.Remove(p); err != nil {
							log.Warn().Err(err).Str("path", p).Msg("Failed to remove temp file")
						}
					}
				}()

				log.Info().Str("path", tempPath2).Msg("Downloading Dailymotion video to temp file using yt-dlp")

				readCookieHeader := func(path string) string {
					if path == "" {
						return ""
					}
					b, err := os.ReadFile(path)
					if err != nil {
						return ""
					}
					lines := strings.Split(string(b), "\n")
					seen := make(map[string]struct{}, 64)
					var parts []string
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line == "" || strings.HasPrefix(line, "#") {
							continue
						}
						fields := strings.Split(line, "\t")
						if len(fields) < 7 {
							continue
						}
						name := strings.TrimSpace(fields[5])
						value := strings.TrimSpace(fields[6])
						if name == "" || value == "" {
							continue
						}
						if _, ok := seen[name]; ok {
							continue
						}
						seen[name] = struct{}{}
						parts = append(parts, fmt.Sprintf("%s=%s", name, value))
					}
					return strings.Join(parts, "; ")
				}

				cookieHeader := readCookieHeader(cookieFilePath)
				outboundProxy2 := strings.TrimSpace(os.Getenv("OUTBOUND_PROXY_URL"))
				outboundProxy2 = strings.TrimSpace(strings.Trim(strings.Trim(strings.Trim(outboundProxy2, "`"), "\""), "'"))
				manifestFile, _ := os.CreateTemp("", "manifest-*.m3u8")
				manifestPath := ""
				if manifestFile != nil {
					manifestPath = manifestFile.Name()
					manifestFile.Close()
					defer os.Remove(manifestPath)

					py := "/opt/venv/bin/python"
					pyCode := `
import sys
from curl_cffi import requests

url = sys.argv[1]
ua = sys.argv[2]
referer = sys.argv[3]
cookie = sys.argv[4]
out_path = sys.argv[5]
proxy = sys.argv[6]

headers = {
  "User-Agent": ua,
  "Accept": "application/vnd.apple.mpegurl,application/x-mpegURL,*/*",
  "Accept-Language": "en-US,en;q=0.9",
  "Referer": referer,
  "Origin": "https://www.dailymotion.com",
  "Accept-Encoding": "identity",
}
if cookie:
  headers["Cookie"] = cookie

proxies = None
if proxy:
  proxies = {"http": proxy, "https": proxy}

r = requests.get(url, headers=headers, impersonate="chrome124", timeout=60, allow_redirects=True, proxies=proxies)
try:
  r.raise_for_status()
  with open(out_path, "wb") as f:
    f.write(r.content)
finally:
  r.close()
`
					manifestCmd := exec.CommandContext(ctx, py, "-c", pyCode, m3u8URL, ua, targetURL, cookieHeader, manifestPath, outboundProxy2)
					var manifestStderr bytes.Buffer
					manifestCmd.Stderr = &manifestStderr
					if err := manifestCmd.Run(); err != nil {
						return response.Error(c, fiber.StatusInternalServerError, "Download failed", "manifest fetch failed: "+manifestStderr.String())
					}

					b, err := os.ReadFile(manifestPath)
					if err != nil || !bytes.HasPrefix(b, []byte("#EXTM3U")) {
						s := string(b)
						if len(s) > 300 {
							s = s[:300]
						}
						return response.Error(c, fiber.StatusInternalServerError, "Download failed", "manifest invalid: "+s)
					}
				}

				ffmpegArgs := []string{
					"-y",
					"-loglevel", "error",
					"-user_agent", ua,
				}

				var headerLines []string
				headerLines = append(headerLines, fmt.Sprintf("User-Agent: %s", ua))
				headerLines = append(headerLines, fmt.Sprintf("Referer: %s", targetURL))
				headerLines = append(headerLines, "Origin: https://www.dailymotion.com")
				if cookieHeader != "" {
					headerLines = append(headerLines, fmt.Sprintf("Cookie: %s", cookieHeader))
				}
				ffmpegArgs = append(ffmpegArgs, "-headers", strings.Join(headerLines, "\r\n")+"\r\n")
				ffmpegArgs = append(ffmpegArgs, "-reconnect", "1", "-reconnect_streamed", "1", "-reconnect_delay_max", "2")
				if outboundProxy2 != "" {
					ffmpegArgs = append(ffmpegArgs, "-http_proxy", outboundProxy2)
				}

				input := m3u8URL
				if manifestPath != "" {
					ffmpegArgs = append(ffmpegArgs, "-protocol_whitelist", "file,crypto,tcp,tls,https,http")
					input = manifestPath
				}

				ffmpegArgs = append(ffmpegArgs,
					"-i", input,
					"-c:v", "copy",
					"-c:a", "aac",
					"-b:a", "128k",
					"-movflags", "+faststart",
					"-f", "mp4",
					tempPath2,
				)

				ffmpegCmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)
				var ffmpegStderr bytes.Buffer
				ffmpegCmd.Stderr = &ffmpegStderr

				log.Info().Msg("Starting ffmpeg HLS download to file")
				ffmpegErr := ffmpegCmd.Run()

				if ffmpegErr != nil {
					log.Warn().Err(ffmpegErr).Str("stderr", ffmpegStderr.String()).Msg("ffmpeg download failed, falling back to yt-dlp")

					ytDlpArgs := []string{
						"-o", tempPath,
						"--no-warnings",
						"--force-overwrites",
						"--no-part",
						"--force-generic-extractor",
					}

					ytDlpArgs = append(ytDlpArgs, "--add-header", fmt.Sprintf("User-Agent: %s", ua))
					ytDlpArgs = append(ytDlpArgs, "--add-header", fmt.Sprintf("Referer: %s", targetURL))
					ytDlpArgs = append(ytDlpArgs, "--referer", targetURL)
					ytDlpArgs = append(ytDlpArgs, "--add-header", "Origin: https://www.dailymotion.com")

					if cookieFilePath != "" {
						ytDlpArgs = append(ytDlpArgs, "--cookies", cookieFilePath)
					}

					ytDlpArgs = append(ytDlpArgs, m3u8URL)

					cmd := exec.CommandContext(ctx, "yt-dlp", ytDlpArgs...)
					var stderr bytes.Buffer
					cmd.Stderr = &stderr

					log.Info().Msg("Starting yt-dlp download to file")

					if err := cmd.Run(); err != nil {
						log.Error().Err(err).Str("stderr", stderr.String()).Msg("yt-dlp download failed")
						log.Warn().Msg("Dailymotion yt-dlp (m3u8) failed, trying yt-dlp on page URL")

						pageArgs := []string{
							"-o", tempPath,
							"--no-warnings",
							"--force-overwrites",
							"--no-part",
							"--merge-output-format", "mp4",
							"--no-playlist",
						}
						pageArgs = append(pageArgs, "--referer", targetURL)
						pageArgs = append(pageArgs, "--add-header", fmt.Sprintf("User-Agent: %s", ua))
						pageArgs = append(pageArgs, "--add-header", "Origin: https://www.dailymotion.com")
						if cookieFilePath != "" {
							pageArgs = append(pageArgs, "--cookies", cookieFilePath)
						}
						pageArgs = append(pageArgs, targetURL)

						pageCmd := exec.CommandContext(ctx, "yt-dlp", pageArgs...)
						var pageStderr bytes.Buffer
						pageCmd.Stderr = &pageStderr
						if err2 := pageCmd.Run(); err2 != nil {
							return response.Error(c, fiber.StatusInternalServerError, "Download failed", stderr.String()+"\n"+pageStderr.String())
						}
					}
				}

				probeCodec := func(streamSelector string) string {
					cmd := exec.CommandContext(ctx, "ffprobe",
						"-v", "error",
						"-select_streams", streamSelector,
						"-show_entries", "stream=codec_name",
						"-of", "default=nw=1:nk=1",
						tempPath,
					)
					out, err := cmd.Output()
					if err != nil {
						return ""
					}
					return strings.TrimSpace(string(out))
				}

				vcodec := probeCodec("v:0")
				acodec := probeCodec("a:0")
				if vcodec != "" || acodec != "" {
					log.Info().Str("vcodec", vcodec).Str("acodec", acodec).Msg("Dailymotion downloaded codecs")
				}

				if vcodec != "" && vcodec != "h264" {
					transFile, err := os.CreateTemp("", "dailymotion-h264-*.mp4")
					if err == nil {
						transPath := transFile.Name()
						transFile.Close()
						cleanupPaths = append(cleanupPaths, transPath)

						transcodeArgs := []string{
							"-y",
							"-loglevel", "error",
							"-i", tempPath2,
							"-c:v", "libx264",
							"-preset", "veryfast",
							"-crf", "23",
							"-c:a", "aac",
							"-b:a", "128k",
							"-movflags", "+faststart",
							transPath,
						}

						transcodeCmd := exec.CommandContext(ctx, "ffmpeg", transcodeArgs...)
						var transcodeStderr bytes.Buffer
						transcodeCmd.Stderr = &transcodeStderr
						log.Info().Msg("Transcoding Dailymotion to h264/aac for compatibility")
						if err := transcodeCmd.Run(); err == nil {
							tempPath2 = transPath
						} else {
							log.Warn().Err(err).Str("stderr", transcodeStderr.String()).Msg("Dailymotion transcode failed, serving original file")
						}
					}
				}

				// Verify file exists and has size
				fileInfo, err := os.Stat(tempPath2)
				if err != nil {
					log.Error().Err(err).Msg("Failed to stat downloaded file")
					return response.Error(c, fiber.StatusInternalServerError, "Download verification failed", err.Error())
				}

				if fileInfo.Size() == 0 {
					log.Error().Msg("Downloaded file is empty")
					return response.Error(c, fiber.StatusInternalServerError, "Downloaded file is empty", "Zero bytes downloaded")
				}

				log.Info().Int64("size", fileInfo.Size()).Msg("Download completed successfully, serving file")

				// Set headers for response
				c.Set("Content-Type", "video/mp4")
				c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))

				// Send the file
				return c.SendFile(tempPath2)
			}

			tempFile = filepath.Join(os.TempDir(), "download-"+uuid.New().String()+".mp4")

			args = append(args, "-o", tempFile, targetURL)

			cmd := exec.CommandContext(ctx, "yt-dlp", args...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			if err := cmd.Run(); err == nil {
				goto ServeFile
			}

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

				// Set User-Agent (default to Desktop)
				ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"

				// Try to get cookies from Redis for this task
				if taskCookies, err := h.svc.GetTaskCookies(ctx, id); err == nil && len(taskCookies) > 0 {
					var cookieStrings []string
					for k, v := range taskCookies {
						cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", k, v))
					}
					req.Header.Set("Cookie", strings.Join(cookieStrings, "; "))

					// If we have cookies and it's TikTok, switch to Mobile UA
					// This matches the strategy used in ChromedpStrategy
					if isTikTok {
						ua = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36"
						log.Info().Str("task_id", id.String()).Msg("Using Mobile UA and Cookies for TikTok proxy download")
					}
				}

				req.Header.Set("User-Agent", ua)

				// Pass existing headers? Maybe Referer
				if isRumble {
					req.Header.Set("Referer", "https://rumble.com/")
				}
				if isTikTok {
					req.Header.Set("Referer", "https://www.tiktok.com/")
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
				contentType := resp.Header.Get("Content-Type")
				if contentType == "" || strings.HasPrefix(contentType, "text/") || strings.HasPrefix(contentType, "application/octet-stream") {
					// For Dailymotion or others, if we get text/html or text/plain, it's likely an error page or m3u8
					if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "text/plain") {
						// Check if it's really a video but mislabeled?
						// Unlikely for text/html.
						// But for application/octet-stream it might be a video.
						// Let's be strict about text/html
						if strings.Contains(contentType, "text/html") {
							resp.Body.Close()
							return response.Error(c, fiber.StatusBadGateway, "Remote server returned HTML instead of video", nil)
						}
					}
					contentType = "video/mp4"
				}
				c.Set("Content-Type", contentType)
				c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))
				if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
					c.Set("Content-Length", contentLength)
				}

				// Wrap body to ensure it closes on EOF
				reader := &autoCloseReader{ReadCloser: resp.Body}
				return c.SendStream(reader)
			}

			return response.Error(c, fiber.StatusBadGateway, "Failed to download video", err)
		}

	ServeFile:
		// Verify file exists
		if _, err := os.Stat(tempFile); err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "Downloaded file not found", err.Error())
		}

		// Explicitly set Content-Type to video/mp4 to ensure browser handles it as video
		c.Set("Content-Type", "video/mp4")
		// Set Content-Disposition with UTF-8 support
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))

		defer os.Remove(tempFile)
		return c.SendFile(tempFile)
	}

	// Fallback: simple HTTP proxy based on direct file URL (non-task based)
	return h.proxyDirectURL(c)
}

// proxyDirectURL handles simple HTTP proxy based on direct file URL (non-task based)
func (h *DownloadHandler) proxyDirectURL(c *fiber.Ctx) error {
	ctx := middleware.HandlerContext(c)
	urlStr := c.Query("url")
	filename := c.Query("filename")

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
	req.Header.Set("Referer", "https://www.tiktok.com/") // Try to set Referer for TikTok/Others

	client := &http.Client{
		Timeout: 0,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow redirects
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			// Copy headers from original request
			for key, val := range via[0].Header {
				req.Header[key] = val
			}
			return nil
		},
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
	// If Content-Type is text/plain or text/html, it might be misconfigured upstream or a text error
	// Force it to video/mp4 if the user requested a file download and we are proxying
	if contentType == "" || strings.HasPrefix(contentType, "text/") {
		if strings.HasSuffix(filename, ".mp4") {
			contentType = "video/mp4"
		} else {
			contentType = "application/octet-stream"
		}
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

func (h *DownloadHandler) ProxyDownloadMp3(c *fiber.Ctx) error {
	return response.Success(c, "Download processed successfully", nil)
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
