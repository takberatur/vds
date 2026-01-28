package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/user/video-downloader-backend/internal/config"
	"github.com/user/video-downloader-backend/internal/infrastructure"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
	"github.com/user/video-downloader-backend/pkg/logger"
	"github.com/user/video-downloader-backend/pkg/utils"
)

func main() {
	logger.Init()
	log.Info().Msg("Starting worker...")

	cfg := config.LoadConfig()

	db, err := infrastructure.NewPostgresClient(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Pool.Close()

	downloadRepo := repository.NewDownloadRepository(db.Pool)
	downloader := infrastructure.NewFallbackDownloader()

	storageClient, err := infrastructure.NewStorageClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioUseSSL,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create storage client")
	}

	// Ensure bucket exists
	if err := storageClient.CreateBucket(context.Background(), cfg.MinioBucket); err != nil {
		log.Fatal().Err(err).Msg("failed to create minio bucket")
	}

	// Start Cleanup Cron Job
	go startCleanupCron(context.Background(), downloadRepo, storageClient, cfg.MinioBucket)

	server := infrastructure.NewTaskServer(cfg.RedisAddr, cfg.RedisPassword)

	redisClient, err := infrastructure.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis for events")
	}
	defer redisClient.Close()

	// Initialize Centrifugo Client
	centrifugoClient := infrastructure.NewCentrifugoClient(cfg.CentrifugoURL, cfg.CentrifugoAPIKey)

	mux := asynq.NewServeMux()

	mux.HandleFunc(infrastructure.TypeVideoDownload, func(ctx context.Context, t *asynq.Task) error {
		var task model.DownloadTask
		if err := json.Unmarshal(t.Payload(), &task); err != nil {
			return err
		}

		if err := handleVideoDownloadTask(ctx, downloadRepo, redisClient, centrifugoClient, downloader, storageClient, cfg.MinioBucket, cfg.EncryptionKey, &task); err != nil {
			return err
		}

		return nil
	})

	if err := server.Run(mux); err != nil {
		log.Fatal().Err(err).Msg("asynq server stopped with error")
	}
}

func startCleanupCron(ctx context.Context, downloadRepo repository.DownloadRepository, storageClient infrastructure.StorageClient, bucketName string) {
	// Run every 10 minutes
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	log.Info().Msg("Cleanup cron job initialized (interval: 10m, retention: 30m)")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Info().Msg("Starting cleanup cron job execution")
			// Cleanup tasks older than 30 minutes
			cutoff := time.Now().Add(-30 * time.Minute)

			// Fetch in batches
			for {
				tasks, err := downloadRepo.FindOldAndCompleted(ctx, cutoff, 100)
				if err != nil {
					log.Error().Err(err).Msg("Failed to find old tasks for cleanup")
					break
				}
				if len(tasks) == 0 {
					break
				}

				var idsToDelete []uuid.UUID
				for _, task := range tasks {
					// Delete files from MinIO
					// Folder structure: platform_type/task_id/
					prefix := fmt.Sprintf("%s/%s/", task.PlatformType, task.ID.String())
					if err := storageClient.DeleteFolder(ctx, bucketName, prefix); err != nil {
						log.Error().Err(err).Str("task_id", task.ID.String()).Msg("Failed to delete folder from MinIO")
						// We continue to delete from DB to avoid orphan records,
						// or we could skip adding to idsToDelete.
						// Let's delete from DB to keep it clean as requested.
					}
					idsToDelete = append(idsToDelete, task.ID)
				}

				if len(idsToDelete) > 0 {
					if err := downloadRepo.BulkDelete(ctx, idsToDelete); err != nil {
						log.Error().Err(err).Msg("Failed to bulk delete tasks from DB")
					} else {
						log.Info().Int("count", len(idsToDelete)).Msg("Deleted old tasks and files")
					}
				}
			}
		}
	}
}

func publishDownloadEvent(ctx context.Context, redisClient infrastructure.RedisClient, centrifugoClient infrastructure.CentrifugoClient, event *model.DownloadEvent) error {
	// 1. Publish to Redis (Old method - kept for backward compatibility during migration)
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err := redisClient.Publish(ctx, infrastructure.DownloadEventChannel, data).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to publish event to Redis")
		// Don't return error here, try Centrifugo too
	}

	// 2. Publish to Centrifugo (New method - per-task channel)
	// Channel format: "download:progress:<task_id>"
	channel := fmt.Sprintf("download:progress:%s", event.TaskID.String())
	if err := centrifugoClient.Publish(ctx, channel, event); err != nil {
		log.Error().Err(err).Str("channel", channel).Msg("Failed to publish event to Centrifugo")
		// We log error but don't fail the task, as this is notification only
	}

	return nil
}

func handleVideoDownloadTask(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, centrifugoClient infrastructure.CentrifugoClient, downloader infrastructure.DownloaderClient, storageClient infrastructure.StorageClient, bucketName string, encryptionKey string, task *model.DownloadTask) error {
	task.Status = "processing"
	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task to processing")
	}

	if err := publishStartEvent(ctx, redisClient, centrifugoClient, task); err != nil {
		log.Error().Err(err).Msg("failed to publish start event")
	}

	if err := processDownloadTask(ctx, downloadRepo, redisClient, centrifugoClient, downloader, storageClient, bucketName, encryptionKey, task); err != nil {
		failErr := markTaskFailed(ctx, downloadRepo, redisClient, centrifugoClient, task, err)
		if failErr != nil {
			log.Error().Err(failErr).Str("task_id", task.ID.String()).Msg("failed to mark task as failed")
		}
		return err
	}

	return nil
}

func processDownloadTask(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, centrifugoClient infrastructure.CentrifugoClient, downloader infrastructure.DownloaderClient, storageClient infrastructure.StorageClient, bucketName string, encryptionKey string, task *model.DownloadTask) error {
	if err := publishProgressEvent(ctx, redisClient, centrifugoClient, task, 10); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Int("progress", 10).Msg("failed to publish progress event (start)")
	}

	// 1. Get Video Info
	info, err := downloader.GetVideoInfo(ctx, task.OriginalURL)
	if err != nil {
		return err
	}

	// 2. Update Task Metadata
	if info != nil {
		if info.Title != "" {
			t := info.Title
			task.Title = &t
		}
		if info.Thumbnail != "" {
			t := info.Thumbnail
			task.ThumbnailURL = &t
		}
		if info.Duration != nil && *info.Duration > 0 {
			d := int(*info.Duration)
			task.Duration = &d
		}
	}

	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task metadata")
	}

	if err := publishProgressEvent(ctx, redisClient, centrifugoClient, task, 30); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Int("progress", 30).Msg("failed to publish progress event (metadata)")
	}

	isDailymotion := strings.Contains(strings.ToLower(task.OriginalURL), "dailymotion.com") || strings.Contains(strings.ToLower(task.OriginalURL), "dai.ly")

	// Special handling for platforms with direct URLs (YouTube, Facebook, Twitter/X)
	// We skip the download-upload loop and just save the direct URLs
	isTwitter := strings.ToLower(task.PlatformType) == "twitter" ||
		strings.ToLower(task.PlatformType) == "x" ||
		strings.Contains(strings.ToLower(task.OriginalURL), "twitter.com") ||
		strings.Contains(strings.ToLower(task.OriginalURL), "x.com") ||
		strings.Contains(strings.ToLower(task.OriginalURL), "twimg.com")

	isInstagram := strings.ToLower(task.PlatformType) == "instagram" ||
		strings.Contains(strings.ToLower(task.OriginalURL), "instagram.com")

	isTiktok := strings.ToLower(task.PlatformType) == "tiktok" ||
		strings.Contains(strings.ToLower(task.OriginalURL), "tiktok.com")

	if strings.ToLower(task.PlatformType) == "youtube" ||
		strings.ToLower(task.PlatformType) == "facebook" ||
		isTwitter ||
		isInstagram ||
		isTiktok {
		log.Info().Str("platform", task.PlatformType).Msg("Processing as direct download (no-upload)")
		return processDirectLinkTask(ctx, downloadRepo, redisClient, centrifugoClient, downloader, task, info, encryptionKey)
	}

	if isDailymotion {
		log.Info().Str("url", task.OriginalURL).Msg("Processing Dailymotion with Chromedp + ffmpeg (upload)")

		outboundProxy := sanitizeProxyURL(os.Getenv("OUTBOUND_PROXY_URL"))
		tempFile, err := os.CreateTemp("", "dailymotion-*.mp4")
		if err != nil {
			return err
		}
		tempPath := tempFile.Name()
		tempFile.Close()
		defer os.Remove(tempPath)

		downloaded := false
		ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"

		if videoID := extractDailymotionID(task.OriginalURL); videoID != "" {
			if m3u8URL, title, thumb, dur, err := fetchDailymotionMasterPlaylist(ctx, videoID, task.OriginalURL, outboundProxy); err == nil && m3u8URL != "" {
				if err := downloadHLSWithFFmpeg(ctx, m3u8URL, ua, task.OriginalURL, "", tempPath); err == nil {
					if title != "" {
						t := title
						task.Title = &t
					}
					if thumb != "" {
						t := thumb
						task.ThumbnailURL = &t
					}
					if dur > 0 {
						d := int(dur)
						task.Duration = &d
					}
					downloaded = true
				}
			}
		}

		if !downloaded {
			chromedpStrategy := infrastructure.NewChromedpStrategy()
			m3u8URL, cookieFilePath, chromedpUA, err := chromedpStrategy.GetMasterPlaylist(ctx, task.OriginalURL)
			if cookieFilePath != "" {
				defer os.Remove(cookieFilePath)
			}
			if err != nil {
				task.Status = "failed"
				msg := err.Error()
				task.ErrorMessage = &msg
				_ = downloadRepo.Update(ctx, task)
				_ = markTaskFailed(ctx, downloadRepo, redisClient, centrifugoClient, task, err)
				return err
			}
			if chromedpUA != "" {
				ua = chromedpUA
			}

			imp := strings.TrimSpace(os.Getenv("YTDLP_IMPERSONATE"))
			if imp == "" {
				imp = "chrome"
			}

			run := func(args []string) (string, error) {
				cmd := exec.CommandContext(ctx, "yt-dlp", args...)
				var stderr bytes.Buffer
				cmd.Stderr = &stderr
				err := cmd.Run()
				return stderr.String(), err
			}

			pageArgs := []string{
				"--no-warnings",
				"--no-playlist",
				"--force-overwrites",
				"--no-part",
				"--merge-output-format", "mp4",
				"-o", tempPath,
				"--referer", task.OriginalURL,
				"--add-header", fmt.Sprintf("User-Agent: %s", ua),
				"--add-header", "Origin: https://www.dailymotion.com",
			}
			pageArgsWithImp := append(append([]string{}, pageArgs...), "--impersonate", imp)
			if outboundProxy != "" {
				pageArgs = append(pageArgs, "--proxy", outboundProxy)
				pageArgsWithImp = append(pageArgsWithImp, "--proxy", outboundProxy)
			}
			if cookieFilePath != "" {
				pageArgs = append(pageArgs, "--cookies", cookieFilePath)
				pageArgsWithImp = append(pageArgsWithImp, "--cookies", cookieFilePath)
			}
			pageArgs = append(pageArgs, task.OriginalURL)
			pageArgsWithImp = append(pageArgsWithImp, task.OriginalURL)

			if stderr, err := run(pageArgsWithImp); err == nil {
				if fi, err := os.Stat(tempPath); err == nil && fi.Size() > 0 {
					downloaded = true
				}
			} else if strings.Contains(stderr, "Impersonate target") {
				if stderr2, err2 := run(pageArgs); err2 == nil {
					if fi, err := os.Stat(tempPath); err == nil && fi.Size() > 0 {
						downloaded = true
					}
				} else {
					log.Warn().Err(err2).Str("stderr", stderr2).Msg("Dailymotion yt-dlp page download failed, falling back to ffmpeg HLS")
				}
			} else {
				log.Warn().Err(err).Str("stderr", stderr).Msg("Dailymotion yt-dlp page download failed, falling back to ffmpeg HLS")
			}

			if !downloaded {
				cookieHeader := netscapeCookiesToHeader(cookieFilePath, []string{"dailymotion.com"})
				if err := downloadHLSWithFFmpeg(ctx, m3u8URL, ua, task.OriginalURL, cookieHeader, tempPath); err != nil {
					task.Status = "failed"
					msg := err.Error()
					task.ErrorMessage = &msg
					_ = downloadRepo.Update(ctx, task)
					_ = markTaskFailed(ctx, downloadRepo, redisClient, centrifugoClient, task, err)
					return err
				}
				downloaded = true
			}
		}

		fiLocal, err := os.Stat(tempPath)
		if err != nil || fiLocal.Size() == 0 {
			e := fmt.Errorf("downloaded file is empty")
			task.Status = "failed"
			msg := e.Error()
			task.ErrorMessage = &msg
			_ = downloadRepo.Update(ctx, task)
			_ = markTaskFailed(ctx, downloadRepo, redisClient, centrifugoClient, task, e)
			return e
		}

		f, err := os.Open(tempPath)
		if err != nil {
			return err
		}
		fi, _ := f.Stat()
		objectName := fmt.Sprintf("%s/%s/%s.%s", "dailymotion", task.ID.String(), "best", "mp4")
		minioURL, err := storageClient.UploadFile(ctx, bucketName, objectName, f, fi.Size(), "video/mp4")
		f.Close()
		if err != nil {
			task.Status = "failed"
			msg := err.Error()
			task.ErrorMessage = &msg
			_ = downloadRepo.Update(ctx, task)
			_ = markTaskFailed(ctx, downloadRepo, redisClient, centrifugoClient, task, err)
			return err
		}

		ext := "mp4"
		size := fi.Size()
		fID := "best"
		res := "best"
		downloadFile := &model.DownloadFile{
			DownloadID:    task.ID,
			URL:           minioURL,
			FormatID:      &fID,
			Resolution:    &res,
			Extension:     &ext,
			FileSize:      &size,
			EncryptedData: nil,
		}
		_ = downloadRepo.AddFile(ctx, downloadFile)

		task.FilePath = &minioURL
		task.FileSize = &size
		task.Format = &fID
		task.Status = "completed"
		if err := downloadRepo.Update(ctx, task); err != nil {
			log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task to completed")
		}
		task.DownloadFiles = []model.DownloadFile{*downloadFile}
		if err := publishCompletionEvent(ctx, redisClient, centrifugoClient, task); err != nil {
			log.Error().Err(err).Msg("failed to publish complete event")
		}
		return nil
	}

	// 3. Select Formats to Download
	// For now, let's pick the best video format and maybe some common ones
	// Or just download the best one first as requested, and then handle others if needed.
	// The user wants "pilihan semua format jika ada".
	// Let's filter formats to get unique resolutions (e.g. 1080p, 720p, 480p, 360p)
	selectedFormats := pickFormatsToDownload(info.Formats)

	// If no formats found but we have a download URL, use it
	if len(selectedFormats) == 0 && info.DownloadURL != "" {
		selectedFormats = append(selectedFormats, infrastructure.FormatInfo{
			URL: info.DownloadURL,
			Ext: "mp4",
		})
	}

	downloadedAny := false
	for i, fmtInfo := range selectedFormats {
		progress := 30 + int(float64(i)/float64(len(selectedFormats))*50)
		if err := publishProgressEvent(ctx, redisClient, centrifugoClient, task, progress); err != nil {
			log.Error().Err(err).Str("task_id", task.ID.String()).Int("progress", progress).Msg("failed to publish progress event (downloading)")
		}

		// 4. Download each format
		// Use resolution in temp filename to avoid invalid chars from selector (e.g. "/")
		tempPattern := fmt.Sprintf("vid-%dp-*.%s", fmtInfo.Height, fmtInfo.Ext)
		if fmtInfo.Height == nil || *fmtInfo.Height == 0 {
			tempPattern = fmt.Sprintf("vid-best-*.%s", fmtInfo.Ext)
		}

		tempFile, err := os.CreateTemp("", tempPattern)
		if err != nil {
			log.Error().Err(err).Msg("failed to create temp file")
			continue
		}
		tempPath := tempFile.Name()
		tempFile.Close()
		defer os.Remove(tempPath)

		err = downloader.DownloadToPath(ctx, task.OriginalURL, fmtInfo.FormatID, tempPath, nil)
		if err != nil {
			log.Error().Err(err).Str("format", fmtInfo.FormatID).Msg("failed to download format")
			continue
		}

		// 5. Upload to MinIO
		f, err := os.Open(tempPath)
		if err != nil {
			log.Error().Err(err).Msg("failed to open temp file for upload")
			continue
		}

		fi, _ := f.Stat()
		resolution := ""
		if fmtInfo.Height != nil && *fmtInfo.Height > 0 {
			resolution = fmt.Sprintf("%dp", *fmtInfo.Height)
		}

		// Use resolution for filename to avoid special chars from format selector
		safeName := resolution
		if safeName == "" {
			safeName = "best"
		}
		objectName := fmt.Sprintf("%s/%s/%s.%s", task.PlatformType, task.ID.String(), safeName, fmtInfo.Ext)

		contentType := "video/mp4"
		if fmtInfo.Ext == "webm" {
			contentType = "video/webm"
		}

		minioURL, err := storageClient.UploadFile(ctx, bucketName, objectName, f, fi.Size(), contentType)
		f.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to upload to minio")
			continue
		}
		downloadedAny = true

		// 6. Save to download_files
		ext := fmtInfo.Ext
		size := fi.Size()
		// For DB, we can store the resolution as format_id if actual ID is complex selector
		fID := safeName
		res := resolution

		downloadFile := &model.DownloadFile{
			DownloadID:    task.ID,
			URL:           minioURL,
			FormatID:      &fID,
			Resolution:    &res,
			Extension:     &ext,
			FileSize:      &size,
			EncryptedData: nil,
		}
		if err := downloadRepo.AddFile(ctx, downloadFile); err != nil {
			log.Error().Err(err).Msg("failed to add download file record")
		}

		// Set primary file path to the first (usually best) format
		if i == 0 {
			task.FilePath = &minioURL
			task.FileSize = &size
			task.Format = &fID
		}
	}

	// 7. Update Task Status
	if downloadedAny {
		task.Status = "completed"
		if err := downloadRepo.Update(ctx, task); err != nil {
			log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task to completed")
		}

		if err := publishCompletionEvent(ctx, redisClient, centrifugoClient, task); err != nil {
			log.Error().Err(err).Msg("failed to publish complete event")
		}
	} else {
		task.Status = "failed"
		msg := "all formats failed to download"
		task.ErrorMessage = &msg
		_ = markTaskFailed(ctx, downloadRepo, redisClient, centrifugoClient, task, fmt.Errorf("%s", msg))
		return fmt.Errorf("%s", msg)
	}

	return nil
}

func netscapeCookiesToHeader(path string, domainSuffixes []string) string {
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(b), "\n")
	m := make(map[string]string, 64)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		domain := strings.TrimSpace(fields[0])
		domain = strings.TrimPrefix(domain, ".")
		ok := false
		for _, suf := range domainSuffixes {
			if strings.HasSuffix(domain, suf) {
				ok = true
				break
			}
		}
		if !ok {
			continue
		}
		name := strings.TrimSpace(fields[5])
		value := strings.TrimSpace(fields[6])
		if name == "" || value == "" {
			continue
		}
		m[name] = value
	}
	if len(m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, "; ")
}

func sanitizeProxyURL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, "`")
	s = strings.Trim(s, "\"")
	s = strings.Trim(s, "'")
	return strings.TrimSpace(s)
}

func extractDailymotionID(u string) string {
	u = strings.TrimSpace(u)
	u = strings.Trim(u, "`")
	u = strings.Trim(u, "\"")
	u = strings.Trim(u, "'")
	reList := []*regexp.Regexp{
		regexp.MustCompile(`dailymotion\.com/video/([a-zA-Z0-9]+)`),
		regexp.MustCompile(`dai\.ly/([a-zA-Z0-9]+)`),
	}
	for _, re := range reList {
		if m := re.FindStringSubmatch(u); len(m) > 1 {
			return m[1]
		}
	}
	return ""
}

func fetchDailymotionMasterPlaylist(ctx context.Context, videoID string, referer string, outboundProxy string) (string, string, string, float64, error) {
	reqURL := fmt.Sprintf("https://www.dailymotion.com/player/metadata/video/%s", videoID)

	var transport http.RoundTripper = http.DefaultTransport
	if outboundProxy != "" {
		if pu, err := url.Parse(outboundProxy); err == nil {
			transport = &http.Transport{Proxy: http.ProxyURL(pu)}
		}
	}
	client := &http.Client{Timeout: 25 * time.Second, Transport: transport}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", "", "", 0, err
	}

	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	req.Header.Set("Origin", "https://www.dailymotion.com")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", "", "", 0, fmt.Errorf("metadata status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", "", "", 0, err
	}

	title, _ := raw["title"].(string)
	thumb, _ := raw["poster_url"].(string)

	var duration float64
	switch v := raw["duration"].(type) {
	case float64:
		duration = v
	case int:
		duration = float64(v)
	}

	qualities, _ := raw["qualities"].(map[string]interface{})
	if len(qualities) == 0 {
		return "", title, thumb, duration, fmt.Errorf("metadata missing qualities")
	}

	type stream struct {
		url  string
		kind string
	}

	parseList := func(v interface{}) []stream {
		arr, ok := v.([]interface{})
		if !ok {
			return nil
		}
		out := make([]stream, 0, len(arr))
		for _, it := range arr {
			m, ok := it.(map[string]interface{})
			if !ok {
				continue
			}
			u, _ := m["url"].(string)
			t, _ := m["type"].(string)
			if u == "" {
				continue
			}
			out = append(out, stream{url: u, kind: t})
		}
		return out
	}

	bestKey := ""
	bestNum := -1
	for k := range qualities {
		if n, err := strconv.Atoi(k); err == nil {
			if n > bestNum {
				bestNum = n
				bestKey = k
			}
		}
	}

	tryKeys := []string{}
	if bestKey != "" {
		tryKeys = append(tryKeys, bestKey)
	}
	if _, ok := qualities["auto"]; ok {
		tryKeys = append(tryKeys, "auto")
	}
	for k := range qualities {
		if k == bestKey || k == "auto" {
			continue
		}
		tryKeys = append(tryKeys, k)
	}

	pickFrom := func(list []stream) string {
		for _, s := range list {
			if strings.Contains(s.kind, "application") && strings.Contains(strings.ToLower(s.kind), "mpeg") && strings.Contains(s.url, ".m3u8") {
				return s.url
			}
		}
		for _, s := range list {
			if strings.Contains(s.url, ".m3u8") {
				return s.url
			}
		}
		for _, s := range list {
			if strings.Contains(strings.ToLower(s.kind), "video/mp4") || strings.Contains(s.url, ".mp4") {
				return s.url
			}
		}
		if len(list) > 0 {
			return list[0].url
		}
		return ""
	}

	for _, key := range tryKeys {
		if v, ok := qualities[key]; ok {
			u := pickFrom(parseList(v))
			if u != "" {
				return u, title, thumb, duration, nil
			}
		}
	}

	return "", title, thumb, duration, fmt.Errorf("no stream url found in qualities")
}

func downloadHLSWithFFmpeg(ctx context.Context, m3u8URL string, userAgent string, referer string, cookieHeader string, outPath string) error {
	outboundProxy := sanitizeProxyURL(os.Getenv("OUTBOUND_PROXY_URL"))

	manifestPath := ""
	{
		tmp, err := os.CreateTemp("", "manifest-*.m3u8")
		if err == nil {
			manifestPath = tmp.Name()
			tmp.Close()
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
			cmd := exec.CommandContext(ctx, py, "-c", pyCode, m3u8URL, userAgent, referer, cookieHeader, manifestPath, outboundProxy)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("curl_cffi manifest fetch failed: %w, stderr: %s", err, stderr.String())
			}

			content, err := os.ReadFile(manifestPath)
			if err != nil {
				return err
			}
			if !bytes.HasPrefix(content, []byte("#EXTM3U")) {
				snippet := string(content)
				if len(snippet) > 300 {
					snippet = snippet[:300]
				}
				return fmt.Errorf("manifest is not m3u8: %s", snippet)
			}
		}
	}

	var headerLines []string
	headerLines = append(headerLines, fmt.Sprintf("User-Agent: %s", userAgent))
	headerLines = append(headerLines, fmt.Sprintf("Referer: %s", referer))
	headerLines = append(headerLines, "Origin: https://www.dailymotion.com")
	headerLines = append(headerLines, "Accept: */*")
	headerLines = append(headerLines, "Accept-Language: en-US,en;q=0.9")
	headerLines = append(headerLines, "Accept-Encoding: identity")
	if cookieHeader != "" {
		headerLines = append(headerLines, fmt.Sprintf("Cookie: %s", cookieHeader))
	}

	ffmpegArgs := []string{
		"-y",
		"-loglevel", "error",
		"-user_agent", userAgent,
		"-headers", strings.Join(headerLines, "\r\n") + "\r\n",
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "2",
	}

	if outboundProxy != "" {
		ffmpegArgs = append(ffmpegArgs, "-http_proxy", outboundProxy)
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
		outPath,
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w, stderr: %s", err, stderr.String())
	}
	fi, err := os.Stat(outPath)
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		return fmt.Errorf("downloaded file is empty")
	}
	return nil
}

func processDirectLinkTask(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, centrifugoClient infrastructure.CentrifugoClient, downloader infrastructure.DownloaderClient, task *model.DownloadTask, info *infrastructure.VideoInfo, encryptionKey string) error {
	// 0. Ensure platform type is correct before we start
	// This helps with Twitter detection if it was missed earlier
	lowerURL := strings.ToLower(task.OriginalURL)
	if task.PlatformType == "" {
		if strings.Contains(lowerURL, "twitter.com") || strings.Contains(lowerURL, "x.com") || strings.Contains(lowerURL, "twimg.com") {
			task.PlatformType = "twitter"
		} else if strings.Contains(lowerURL, "youtube.com") || strings.Contains(lowerURL, "youtu.be") {
			task.PlatformType = "youtube"
		} else if strings.Contains(lowerURL, "facebook.com") || strings.Contains(lowerURL, "fb.watch") {
			task.PlatformType = "facebook"
		} else if strings.Contains(lowerURL, "instagram.com") {
			task.PlatformType = "instagram"
		} else if strings.Contains(lowerURL, "tiktok.com") {
			task.PlatformType = "tiktok"
		}
	}

	/*
		isTwitter := strings.EqualFold(task.PlatformType, "twitter") ||
			strings.EqualFold(task.PlatformType, "x") ||
			strings.Contains(lowerURL, "twitter.com") ||
			strings.Contains(lowerURL, "x.com") ||
			strings.Contains(lowerURL, "twimg.com")
	*/

	/*
		isInstagram := strings.EqualFold(task.PlatformType, "instagram") ||
			strings.Contains(lowerURL, "instagram.com")
	*/

	isTiktok := strings.EqualFold(task.PlatformType, "tiktok") ||
		strings.Contains(lowerURL, "tiktok.com")

	if isTiktok {
		return processTikTokEncryptedTask(ctx, downloadRepo, redisClient, centrifugoClient, downloader, task, info, encryptionKey)
	}

	// Helper function to clean URLs
	cleanURL := func(u string) string {
		// Always clean surrounding whitespace and backticks which are invalid in URLs
		// This fixes issues where URLs might be wrapped in backticks from logging or scraping artifacts
		u = strings.TrimSpace(u)
		u = strings.Trim(u, "`")
		u = strings.ReplaceAll(u, "`", "")
		u = strings.Trim(u, "'")
		u = strings.Trim(u, "\"")

		// Note: We do NOT remove query params (e.g. ?tag=21) as they are required for
		// Twitter, Instagram, and TikTok access.
		return u
	}

	// Sanitize task URLs immediately
	if task.ThumbnailURL != nil {
		cleaned := cleanURL(*task.ThumbnailURL)
		task.ThumbnailURL = &cleaned
	}
	if info != nil {
		if info.Thumbnail != "" {
			info.Thumbnail = cleanURL(info.Thumbnail)
		}
		if info.DownloadURL != "" {
			info.DownloadURL = cleanURL(info.DownloadURL)
		}
	}

	// 1. Gather all formats (or filtered)
	// We want to save ALL available formats that have a valid URL
	var formatsToSave []infrastructure.FormatInfo
	if len(info.Formats) > 0 {
		// Filter for YouTube/Facebook/Twitter/Instagram/TikTok:
		// Keep formats, but maybe add metadata to indicate video-only/audio-only if needed in future
		var validFormats []infrastructure.FormatInfo
		for _, f := range info.Formats {
			// Restore Filter: Only keep formats with both Video and Audio
			// This prevents "video only" or "audio only" links which are useless as direct downloads for average users

			// Relaxed check: empty string often means "present but unknown" in some contexts,
			// while "none" explicitly means missing.
			hasVideo := f.Vcodec != "none"
			hasAudio := f.Acodec != "none"

			// If specific codecs are empty, we give benefit of doubt for HTTP formats (often combined)
			// checking ext might help (mp4 usually has both unless specified)
			if f.Vcodec == "" && f.Acodec == "" && f.Ext == "mp4" {
				hasVideo = true
				hasAudio = true
			}

			// Special case for TikTok: yt-dlp might return formats with weird codec strings or "none"
			// but if it has a valid http URL and it's not m3u8, it's likely playable.
			if isTiktok && (strings.HasPrefix(f.URL, "http") && !strings.Contains(f.URL, ".m3u8")) {
				// Assume playable if it's a direct http link for tiktok
				hasVideo = true
				hasAudio = true
			}

			if hasVideo && hasAudio {
				// Clean URL before saving
				f.URL = cleanURL(f.URL)
				validFormats = append(validFormats, f)
			}
		}

		// Use filtered formats if any found, otherwise fallback to all
		if len(validFormats) > 0 {
			formatsToSave = validFormats
		} else {
			formatsToSave = info.Formats
		}
	} else if info.DownloadURL != "" {
		formatsToSave = append(formatsToSave, infrastructure.FormatInfo{
			URL: cleanURL(info.DownloadURL),
			Ext: "mp4", // Default guess
		})
	}

	// 2. Save to download_files table
	// We also need to find the "best" format to set as the main file for the task
	var bestFile *model.DownloadFile

	for i, f := range formatsToSave {
		if f.URL == "" {
			continue
		}

		// Construct DB model
		// Use safe defaults for nil pointers
		var resolution string
		if f.Height != nil && *f.Height > 0 {
			resolution = fmt.Sprintf("%dp", *f.Height)
		} else if f.Width != nil && *f.Width > 0 {
			resolution = fmt.Sprintf("%dw", *f.Width) // fallback
		}

		// Format ID from yt-dlp (e.g. "137", "22", "sb3")
		fmtID := f.FormatID
		if fmtID == "" {
			fmtID = "unknown"
		}

		ext := f.Ext
		if ext == "" {
			ext = "unknown"
		}

		var size int64
		if f.Filesize != nil {
			size = *f.Filesize
		}

		downloadFile := &model.DownloadFile{
			DownloadID:    task.ID,
			URL:           f.URL,
			FormatID:      &fmtID,
			Resolution:    &resolution,
			Extension:     &ext,
			FileSize:      &size,
			EncryptedData: nil,
		}

		// Save to DB
		if err := downloadRepo.AddFile(ctx, downloadFile); err != nil {
			log.Error().Err(err).Str("format_id", fmtID).Msg("failed to add direct download file record")
			// Continue to try other formats
		}

		// Append to task.DownloadFiles so publishCompletionEvent includes them
		task.DownloadFiles = append(task.DownloadFiles, *downloadFile)

		// Determine if this is the "best" file to represent the task
		// Heuristic: Highest resolution, or largest file size
		// Simple logic: if this is the first one, or better than current best
		if bestFile == nil {
			bestFile = downloadFile
		} else {
			// Compare resolution if available
			if f.Height != nil && bestFile.Resolution != nil {
				var currentH int
				fmt.Sscanf(*bestFile.Resolution, "%dp", &currentH)
				if *f.Height > currentH {
					bestFile = downloadFile
				}
			}
		}

		// Progress simulation?
		if i%5 == 0 {
			progress := 30 + int(float64(i)/float64(len(formatsToSave))*60)
			if progress > 99 {
				progress = 99
			}
			publishProgressEvent(ctx, redisClient, centrifugoClient, task, progress)
		}
	}

	// 3. Update Task Status & Main File Info
	if bestFile != nil {
		task.FilePath = &bestFile.URL
		task.FileSize = bestFile.FileSize
		task.Format = bestFile.FormatID
	} else if info.DownloadURL != "" {
		// Fallback if no formats loop worked but top level has URL
		u := info.DownloadURL
		task.FilePath = &u
	}

	// Validate that the FilePath is not the original platform URL
	if task.FilePath != nil {
		lowerPath := strings.ToLower(*task.FilePath)
		// Check if URL looks like a webpage rather than a media file
		// Note: Some CDNs might contain the domain, but usually not the full path structure of a post
		if strings.Contains(lowerPath, "instagram.com/reel/") ||
			strings.Contains(lowerPath, "instagram.com/p/") ||
			strings.Contains(lowerPath, "tiktok.com/@") ||
			strings.Contains(lowerPath, "tiktok.com/video/") ||
			(strings.Contains(lowerPath, "youtube.com/watch") && !strings.Contains(lowerPath, "googlevideo.com")) {

			errMsg := "Failed to extract direct video link (returned platform URL)"
			log.Error().Str("url", *task.FilePath).Msg(errMsg)

			// Mark as failed instead of completed
			return markTaskFailed(ctx, downloadRepo, redisClient, centrifugoClient, task, fmt.Errorf("%s", errMsg))
		}
	}

	task.Status = "completed"
	if err := downloadRepo.Update(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID.String()).Msg("failed to update task to completed")
		return err
	}

	// 4. Publish Completion
	if err := publishCompletionEvent(ctx, redisClient, centrifugoClient, task); err != nil {
		log.Error().Err(err).Msg("failed to publish complete event")
	}

	log.Info().Str("task_id", task.ID.String()).Int("files_count", len(task.DownloadFiles)).Msg("Direct download processing completed")
	return nil
}

func pickFormatsToDownload(formats []infrastructure.FormatInfo) []infrastructure.FormatInfo {
	if len(formats) == 0 {
		return nil
	}

	// Identify available heights from video streams
	// We want to capture the best quality for each resolution bucket
	heights := make(map[int]bool)
	for _, f := range formats {
		// Filter out audio-only (vcodec=none) and very low quality
		if f.Height != nil && *f.Height >= 360 && f.Vcodec != "none" {
			heights[*f.Height] = true
		}
	}

	var hList []int
	for h := range heights {
		hList = append(hList, h)
	}

	// Sort descending
	sort.Sort(sort.Reverse(sort.IntSlice(hList)))

	// Limit to top 4 resolutions
	if len(hList) > 4 {
		hList = hList[:4]
	}

	var result []infrastructure.FormatInfo
	for _, h := range hList {
		// Construct selector for "best video at height H + best audio"
		// This ensures we get combined video/audio file
		selector := fmt.Sprintf("bestvideo[height=%d]+bestaudio/best[height=%d]", h, h)

		result = append(result, infrastructure.FormatInfo{
			FormatID: selector,
			Height:   intPtr(h),
			Ext:      "mp4", // We'll let yt-dlp merge to mp4 (requires ffmpeg)
		})
	}

	return result
}

func intPtr(v int) *int {
	return &v
}

func publishStartEvent(ctx context.Context, redisClient infrastructure.RedisClient, centrifugoClient infrastructure.CentrifugoClient, task *model.DownloadTask) error {
	event := &model.DownloadEvent{
		Type:      "download.processing",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "processing",
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, centrifugoClient, event)
}

func processTikTokEncryptedTask(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, centrifugoClient infrastructure.CentrifugoClient, downloader infrastructure.DownloaderClient, task *model.DownloadTask, info *infrastructure.VideoInfo, encryptionKey string) error {
	log.Info().Str("task_id", task.ID.String()).Msg("Processing TikTok encrypted task")

	outboundProxy := sanitizeProxyURL(os.Getenv("OUTBOUND_PROXY_URL"))

	{
		tempFile, err := os.CreateTemp("", "tiktok-ytdlp-*.mp4")
		if err == nil {
			tempPath := tempFile.Name()
			tempFile.Close()
			defer os.Remove(tempPath)

			args := []string{
				"--no-warnings",
				"--no-playlist",
				"--force-overwrites",
				"--no-part",
				"--merge-output-format", "mp4",
				"-o", tempPath,
			}
			if outboundProxy != "" {
				args = append(args, "--proxy", outboundProxy)
			}
			if _, err := os.Stat("/app/cookies.txt"); err == nil {
				args = append(args, "--cookies", "/app/cookies.txt")
			}
			args = append(args, task.OriginalURL)

			cmd := exec.CommandContext(ctx, "yt-dlp", args...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err == nil {
				b, err := os.ReadFile(tempPath)
				if err == nil && len(b) > 0 {
					encryptedData, err := utils.EncryptData(b, encryptionKey)
					if err != nil {
						return fmt.Errorf("failed to encrypt data: %w", err)
					}

					fileSize := int64(len(b))
					fID := "encrypted"
					resolution := "original"
					ext := "mp4"
					dummyURL := fmt.Sprintf("encrypted://%s/video.%s", task.PlatformType, ext)

					dlFile := &model.DownloadFile{
						DownloadID:    task.ID,
						URL:           dummyURL,
						FormatID:      &fID,
						Resolution:    &resolution,
						Extension:     &ext,
						FileSize:      &fileSize,
						EncryptedData: &encryptedData,
					}

					if err := downloadRepo.AddFile(ctx, dlFile); err != nil {
						return err
					}

					task.EncryptedData = &encryptedData
					task.FilePath = &dummyURL
					task.FileSize = &fileSize
					task.Format = &ext
					task.Status = "completed"

					if err := downloadRepo.Update(ctx, task); err != nil {
						return err
					}

					task.DownloadFiles = []model.DownloadFile{*dlFile}
					if err := publishCompletionEvent(ctx, redisClient, centrifugoClient, task); err != nil {
						log.Error().Err(err).Msg("failed to publish complete event")
					}

					log.Info().Str("task_id", task.ID.String()).Int64("original_size", fileSize).Msg("TikTok encrypted task completed (yt-dlp)")
					return nil
				}
			} else {
				log.Warn().Err(err).Str("stderr", stderr.String()).Msg("TikTok yt-dlp download failed, falling back to direct HTTP")
			}
		}
	}

	var targetURL string
	var ext string = "mp4"

	if info.DownloadURL != "" {
		targetURL = info.DownloadURL
	}

	if len(info.Formats) > 0 {
		var bestFormat *infrastructure.FormatInfo
		for _, f := range info.Formats {
			// Check for valid video format
			// TikTok formats often have "h264" vcodec or just http URL
			// Avoid m3u8
			if f.URL != "" && !strings.Contains(f.URL, ".m3u8") {
				if bestFormat == nil {
					temp := f
					bestFormat = &temp
				} else {
					// Compare height if available
					if f.Height != nil && bestFormat.Height != nil && *f.Height > *bestFormat.Height {
						temp := f
						bestFormat = &temp
					} else if bestFormat.Height == nil && f.Height != nil {
						// Prefer one with height
						temp := f
						bestFormat = &temp
					}
				}
			}
		}
		if bestFormat != nil {
			targetURL = bestFormat.URL
			if bestFormat.Ext != "" {
				ext = bestFormat.Ext
			}
		}
	}

	if targetURL == "" {
		return fmt.Errorf("no suitable download URL found for TikTok task")
	}

	log.Info().Str("target_url", targetURL).Msg("TikTok selected download URL")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	userAgent := ""
	if info != nil {
		userAgent = info.UserAgent
	}
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Mobile Safari/537.36"
	}

	referer := task.OriginalURL
	if info != nil && info.WebpageURL != "" {
		referer = info.WebpageURL
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Dest", "video")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Range", "bytes=0-")
	req.Header.Set("Accept-Encoding", "identity")

	if referer != "" {
		req.Header.Set("Referer", "https://www.tiktok.com/")
		req.Header.Set("Origin", "https://www.tiktok.com")
	}

	if info != nil && len(info.Cookies) > 0 {
		var parts []string
		for k, v := range info.Cookies {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
		if len(parts) > 0 {
			req.Header.Set("Cookie", strings.Join(parts, "; "))
		}
	}

	var transport http.RoundTripper = http.DefaultTransport
	if proxyURL := outboundProxy; proxyURL != "" {
		if pu, err := url.Parse(proxyURL); err == nil {
			transport = &http.Transport{Proxy: http.ProxyURL(pu)}
		}
	}

	client := &http.Client{Timeout: 15 * time.Minute, Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download TikTok video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		if resp.StatusCode == http.StatusForbidden {
			cookieHeader := req.Header.Get("Cookie")
			log.Warn().Int("status", resp.StatusCode).Msg("TikTok direct HTTP got 403, trying curl_cffi impersonation")

			tempFile, err := os.CreateTemp("", "tiktok-curlcffi-*.mp4")
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}
			tempPath := tempFile.Name()
			tempFile.Close()
			defer os.Remove(tempPath)

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

headers = {"User-Agent": ua, "Accept": "*/*", "Accept-Language": "en-US,en;q=0.9", "Referer": referer, "Origin": "https://www.tiktok.com", "Range": "bytes=0-", "Accept-Encoding": "identity"}
if cookie:
    headers["Cookie"] = cookie

proxies = None
if proxy:
    proxies = {"http": proxy, "https": proxy}

r = requests.get(url, headers=headers, impersonate="chrome124", stream=True, timeout=300, allow_redirects=True, proxies=proxies)
try:
    r.raise_for_status()
    with open(out_path, "wb") as f:
        for chunk in r.iter_content(chunk_size=1024 * 1024):
            if chunk:
                f.write(chunk)
finally:
    r.close()
`
			outboundProxy := sanitizeProxyURL(os.Getenv("OUTBOUND_PROXY_URL"))
			curlReq := exec.CommandContext(ctx, py, "-c", pyCode, targetURL, userAgent, "https://www.tiktok.com/", cookieHeader, tempPath, outboundProxy)
			var curlStderr bytes.Buffer
			curlReq.Stderr = &curlStderr
			var curlStdout bytes.Buffer
			curlReq.Stdout = &curlStdout
			if err := curlReq.Run(); err != nil {
				log.Error().Err(err).Str("stderr", curlStderr.String()).Str("stdout", curlStdout.String()).Msg("curl_cffi download failed")
				return fmt.Errorf("failed to download TikTok video: status %d", resp.StatusCode)
			}

			data, err := os.ReadFile(tempPath)
			if err != nil {
				return fmt.Errorf("failed to read TikTok downloaded file: %w", err)
			}
			if len(data) == 0 {
				return fmt.Errorf("downloaded empty file")
			}

			encryptedData, err := utils.EncryptData(data, encryptionKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt data: %w", err)
			}

			fileSize := int64(len(data))
			fID := "encrypted"
			resolution := "original"
			dummyURL := fmt.Sprintf("encrypted://%s/video.%s", task.PlatformType, ext)

			dlFile := &model.DownloadFile{
				DownloadID:    task.ID,
				URL:           dummyURL,
				FormatID:      &fID,
				Resolution:    &resolution,
				Extension:     &ext,
				FileSize:      &fileSize,
				EncryptedData: &encryptedData,
			}

			if err := downloadRepo.AddFile(ctx, dlFile); err != nil {
				return err
			}

			task.EncryptedData = &encryptedData
			task.FilePath = &dummyURL
			task.FileSize = &fileSize
			task.Format = &ext
			task.Status = "completed"

			if err := downloadRepo.Update(ctx, task); err != nil {
				return err
			}

			task.DownloadFiles = []model.DownloadFile{*dlFile}
			if err := publishCompletionEvent(ctx, redisClient, centrifugoClient, task); err != nil {
				log.Error().Err(err).Msg("failed to publish complete event")
			}

			log.Info().Str("task_id", task.ID.String()).Int64("original_size", fileSize).Msg("TikTok encrypted task completed (curl_cffi)")
			return nil
		}
		return fmt.Errorf("failed to download TikTok video: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read TikTok response body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "text/html") && len(body) < 200*1024 {
		return fmt.Errorf("failed to download TikTok video: got html response status %d", resp.StatusCode)
	}

	data := body

	if len(data) == 0 {
		return fmt.Errorf("downloaded empty file")
	}

	encryptedData, err := utils.EncryptData(data, encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// 4. Save to DB
	// We store original file size for metadata, but the data is encrypted
	fileSize := int64(len(data))

	fID := "encrypted"
	resolution := "original"
	// Use a dummy URL that indicates it's encrypted
	dummyURL := fmt.Sprintf("encrypted://%s/video.%s", task.PlatformType, ext)

	dlFile := &model.DownloadFile{
		DownloadID:    task.ID,
		URL:           dummyURL,
		FormatID:      &fID,
		Resolution:    &resolution,
		Extension:     &ext,
		FileSize:      &fileSize,
		EncryptedData: &encryptedData,
	}

	if err := downloadRepo.AddFile(ctx, dlFile); err != nil {
		return err
	}

	task.EncryptedData = &encryptedData
	task.FilePath = &dummyURL
	task.FileSize = &fileSize
	task.Format = &ext
	task.Status = "completed"

	if err := downloadRepo.Update(ctx, task); err != nil {
		return err
	}

	// Publish completion
	// Need to populate DownloadFiles for event
	task.DownloadFiles = []model.DownloadFile{*dlFile}
	if err := publishCompletionEvent(ctx, redisClient, centrifugoClient, task); err != nil {
		log.Error().Err(err).Msg("failed to publish complete event")
	}

	log.Info().Str("task_id", task.ID.String()).Int64("original_size", fileSize).Msg("TikTok encrypted task completed")
	return nil
}

func publishProgressEvent(ctx context.Context, redisClient infrastructure.RedisClient, centrifugoClient infrastructure.CentrifugoClient, task *model.DownloadTask, progress int) error {
	event := &model.DownloadEvent{
		Type:      "download.processing",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "processing",
		Progress:  &progress,
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, centrifugoClient, event)
}

func publishCompletionEvent(ctx context.Context, redisClient infrastructure.RedisClient, centrifugoClient infrastructure.CentrifugoClient, task *model.DownloadTask) error {
	var payload *model.DownloadPayload

	// Prepare payload file path
	var payloadFilePath *string
	if task.FilePath != nil {
		payloadFilePath = task.FilePath
	}

	// Prepare formats for payload
	var payloadFormats []model.DownloadFormat
	for _, f := range task.DownloadFiles {
		format := model.DownloadFormat{
			URL:      f.URL,
			FormatID: getValueOrEmpty(f.FormatID),
			Ext:      getValueOrEmpty(f.Extension),
		}

		if f.FileSize != nil {
			format.Filesize = f.FileSize
		}

		// Try to parse height from resolution if possible, or leave it nil
		if f.Resolution != nil {
			var h int
			if _, err := fmt.Sscanf(*f.Resolution, "%dp", &h); err == nil {
				format.Height = &h
			}
		}

		payloadFormats = append(payloadFormats, format)
	}

	// Ensure platform type is set
	platformType := task.PlatformType
	if platformType == "" {
		// Fallback detection from OriginalURL
		lowerURL := strings.ToLower(task.OriginalURL)
		if strings.Contains(lowerURL, "twitter.com") || strings.Contains(lowerURL, "x.com") {
			platformType = "twitter"
		} else if strings.Contains(lowerURL, "youtube.com") || strings.Contains(lowerURL, "youtu.be") {
			platformType = "youtube"
		} else if strings.Contains(lowerURL, "facebook.com") || strings.Contains(lowerURL, "fb.watch") {
			platformType = "facebook"
		} else if strings.Contains(lowerURL, "instagram.com") {
			platformType = "instagram"
		} else if strings.Contains(lowerURL, "tiktok.com") {
			platformType = "tiktok"
		}
	}

	payload = &model.DownloadPayload{
		ID:           task.ID,
		Status:       task.Status,
		Progress:     100,
		Title:        getValueOrEmpty(task.Title),
		ThumbnailURL: getValueOrEmpty(task.ThumbnailURL),
		Type:         platformType,
		CreatedAt:    task.CreatedAt,
		FilePath:     payloadFilePath,
		Formats:      payloadFormats,
	}

	event := &model.DownloadEvent{
		Type:      "download.completed",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "completed",
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, centrifugoClient, event)
}

func getValueOrEmpty(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func markTaskFailed(ctx context.Context, downloadRepo repository.DownloadRepository, redisClient infrastructure.RedisClient, centrifugoClient infrastructure.CentrifugoClient, task *model.DownloadTask, err error) error {
	task.Status = "failed"
	errMsg := err.Error()
	task.ErrorMessage = &errMsg

	if updateErr := downloadRepo.Update(ctx, task); updateErr != nil {
		return updateErr
	}

	event := &model.DownloadEvent{
		Type:      "download.failed",
		TaskID:    task.ID,
		UserID:    task.UserID,
		Status:    "failed",
		Error:     errMsg,
		CreatedAt: time.Now(),
	}

	return publishDownloadEvent(ctx, redisClient, centrifugoClient, event)
}
