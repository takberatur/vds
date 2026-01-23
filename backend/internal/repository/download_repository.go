package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/video-downloader-backend/internal/infrastructure/contextpool"
	"github.com/user/video-downloader-backend/internal/model"
)

type DownloadRepository interface {
	BaseRepository
	Create(ctx context.Context, task *model.DownloadTask) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.DownloadTask, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.DownloadTask, error)
	FindAll(ctx context.Context, params model.QueryParamsRequest) ([]*model.DownloadTask, model.Pagination, error)
	Update(ctx context.Context, task *model.DownloadTask) error
	Delete(ctx context.Context, id uuid.UUID) error
	BulkDelete(ctx context.Context, ids []uuid.UUID) error
	AddFile(ctx context.Context, file *model.DownloadFile) error
	FindOldAndCompleted(ctx context.Context, cutoff time.Time, limit int) ([]*model.DownloadTask, error)
}

type downloadRepository struct {
	*baseRepository
}

func NewDownloadRepository(db *pgxpool.Pool) DownloadRepository {
	return &downloadRepository{
		baseRepository: NewBaseRepository(db).(*baseRepository),
	}
}

func (r *downloadRepository) Create(ctx context.Context, task *model.DownloadTask) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		INSERT INTO downloads (user_id, app_id, original_url, platform_id, platform_type, status, file_path, format, thumbnail_url, title, file_size, duration, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`
	now := time.Now()
	err := r.db.QueryRow(subCtx, query,
		task.UserID,
		task.AppID,
		task.OriginalURL,
		task.PlatformID,
		task.PlatformType,
		task.Status,
		task.FilePath,
		task.Format,
		task.ThumbnailURL,
		task.Title,
		task.FileSize,
		task.Duration,
		now,
	).Scan(&task.ID)

	task.CreatedAt = now
	return err
}

func (r *downloadRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.DownloadTask, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		SELECT * FROM downloads
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	var tasks []*model.DownloadTask
	if err := pgxscan.Select(subCtx, r.db, &tasks, query, userID, limit, offset); err != nil {
		return nil, err
	}

	if len(tasks) > 0 {
		taskIDs := make([]uuid.UUID, len(tasks))
		for i, t := range tasks {
			taskIDs[i] = t.ID
		}

		filesQuery := `
			SELECT id, download_id, url, format_id, resolution, extension, file_size, created_at
			FROM download_files
			WHERE download_id = ANY($1)
		`
		var allFiles []model.DownloadFile
		if err := pgxscan.Select(subCtx, r.db, &allFiles, filesQuery, taskIDs); err != nil {
			return nil, err
		}

		filesMap := make(map[uuid.UUID][]model.DownloadFile)
		for _, f := range allFiles {
			filesMap[f.DownloadID] = append(filesMap[f.DownloadID], f)
		}

		for _, t := range tasks {
			if files, ok := filesMap[t.ID]; ok {
				t.DownloadFiles = files
				t.Formats = r.mapFilesToFormats(files)
			}
		}
	}

	return tasks, nil
}

func (r *downloadRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.DownloadTask, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
        SELECT 
            d.id, d.user_id, d.app_id, d.platform_id, d.original_url, d.platform_type, d.file_path, d.thumbnail_url, 
            d.title, d.duration, d.file_size, d.format, d.status, d.error_message, d.ip_address, d.created_at,
            u.email as user_email,
            p.name as platform_name, p.slug as platform_slug, p.thumbnail_url as platform_thumbnail_url, 
            p.type as platform_type, p.is_active as platform_is_active, p.is_premium as platform_is_premium
        FROM downloads d 
        LEFT JOIN users u ON d.user_id = u.id
        LEFT JOIN platforms p ON d.platform_id = p.id
        WHERE d.id = $1
    `

	var task model.DownloadTask
	var userEmail *string
	var platformName, platformSlug, platformThumbnailURL, platformType *string
	var platformIsActive, platformIsPremium *bool

	err := r.db.QueryRow(subCtx, query, id).Scan(
		&task.ID, &task.UserID, &task.AppID, &task.PlatformID, &task.OriginalURL, &task.PlatformType,
		&task.FilePath, &task.ThumbnailURL, &task.Title, &task.Duration, &task.FileSize, &task.Format,
		&task.Status, &task.ErrorMessage, &task.IPAddress, &task.CreatedAt,
		&userEmail,
		&platformName, &platformSlug, &platformThumbnailURL, &platformType, &platformIsActive, &platformIsPremium,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("download task not found")
		}
		return nil, err
	}

	if userEmail != nil && task.UserID != nil {
		task.User = &model.User{
			ID:    *task.UserID,
			Email: *userEmail,
		}
	}

	if platformName != nil {
		task.Platform = &model.Platform{
			ID:           task.PlatformID,
			Name:         *platformName,
			Slug:         *platformSlug,
			ThumbnailURL: *platformThumbnailURL,
			Type:         *platformType,
			IsActive:     *platformIsActive,
			IsPremium:    *platformIsPremium,
		}
	}

	filesQuery := `
        SELECT id, download_id, url, format_id, resolution, extension, file_size, created_at
        FROM download_files
        WHERE download_id = $1
        ORDER BY created_at ASC
    `

	filesRows, err := r.db.Query(subCtx, filesQuery, id)
	if err != nil {
		return nil, err
	}
	defer filesRows.Close()

	var downloadFiles []model.DownloadFile
	for filesRows.Next() {
		var file model.DownloadFile
		err = filesRows.Scan(
			&file.ID, &file.DownloadID, &file.URL, &file.FormatID, &file.Resolution,
			&file.Extension, &file.FileSize, &file.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		downloadFiles = append(downloadFiles, file)
	}

	if err = filesRows.Err(); err != nil {
		return nil, err
	}

	task.DownloadFiles = downloadFiles
	task.Formats = r.mapFilesToFormats(downloadFiles)

	return &task, nil
}

func (r *downloadRepository) mapFilesToFormats(files []model.DownloadFile) []model.DownloadFormat {
	if len(files) == 0 {
		return nil
	}
	formats := make([]model.DownloadFormat, 0, len(files))
	for _, file := range files {
		var height *int
		if file.Resolution != nil {
			var h int
			if _, err := fmt.Sscanf(*file.Resolution, "%dp", &h); err == nil {
				height = &h
			}
		}

		formatID := ""
		if file.FormatID != nil {
			formatID = *file.FormatID
		}
		ext := ""
		if file.Extension != nil {
			ext = *file.Extension
		}

		formats = append(formats, model.DownloadFormat{
			URL:      file.URL,
			Filesize: file.FileSize,
			FormatID: formatID,
			Ext:      ext,
			Height:   height,
		})
	}
	return formats
}

func (r *downloadRepository) FindAll(ctx context.Context, params model.QueryParamsRequest) ([]*model.DownloadTask, model.Pagination, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		SELECT 
			d.id, d.user_id, d.app_id, d.platform_id, d.original_url, d.file_path, d.thumbnail_url, 
			d.title, d.duration, d.file_size, d.format, d.status, d.error_message, d.ip_address, d.created_at,
			u.email as user_email,
			p.name as platform_name, p.slug as platform_slug, p.thumbnail_url as platform_thumbnail_url, p.type as platform_type, p.is_active as platform_is_active, p.is_premium as platform_is_premium,
			f.id as file_id, f.download_id, f.url, f.format_id, f.resolution, f.extension, f.file_size, f.created_at
		FROM downloads d
		LEFT JOIN users u ON d.user_id = u.id
		LEFT JOIN platforms p ON d.platform_id = p.id
		LEFT JOIN download_files f ON d.id = f.download_id
	`

	// Apply filters
	whereClauses := []string{}
	args := []interface{}{}
	argId := 1

	if params.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(d.title ILIKE $%d OR d.original_url ILIKE $%d)", argId, argId))
		args = append(args, "%"+params.Search+"%")
		argId++
	}

	if params.UserID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.user_id = $%d", argId))
		args = append(args, params.UserID)
		argId++
	}

	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.status = $%d", argId))
		args = append(args, params.Status)
		argId++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM downloads d"
	if len(whereClauses) > 0 {
		countQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var totalItems int64
	err := r.db.QueryRow(subCtx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, model.Pagination{}, err
	}

	orderBy := "d.created_at"
	if params.SortBy != "" {
		switch params.SortBy {
		case "title":
			orderBy = "d.title"
		case "file_size":
			orderBy = "d.file_size"
		case "status":
			orderBy = "d.status"
		}
	}

	orderDir := "DESC"
	if strings.ToLower(params.OrderBy) == "asc" {
		orderDir = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

	// Apply pagination
	page := params.Page
	if page < 1 {
		page = 1
	}
	limit := params.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argId, argId+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(subCtx, query, args...)
	if err != nil {
		return nil, model.Pagination{}, err
	}
	defer rows.Close()

	var tasks []*model.DownloadTask
	for rows.Next() {
		var task model.DownloadTask
		var userEmail *string
		var platformName, platformSlug, platformThumbnailURL, platformType *string
		var platformIsActive, platformIsPremium *bool
		var fileID uuid.UUID
		var downloadID uuid.UUID
		var url, formatID, resolution, extension *string
		var fileSize *int64
		var createdAt time.Time

		err := rows.Scan(
			&task.ID, &task.UserID, &task.AppID, &task.PlatformID, &task.OriginalURL, &task.PlatformType, &task.FilePath, &task.ThumbnailURL,
			&task.Title, &task.Duration, &task.FileSize, &task.Format, &task.Status, &task.ErrorMessage, &task.IPAddress, &task.CreatedAt,
			&userEmail,
			&platformName, &platformSlug, &platformThumbnailURL, &platformType, &platformIsActive, &platformIsPremium,
			&fileID, &downloadID, &url, &formatID, &resolution, &extension, &fileSize, &createdAt,
		)
		if err != nil {
			return nil, model.Pagination{}, err
		}

		if task.UserID != nil && userEmail != nil {
			task.User = &model.User{
				ID:    *task.UserID,
				Email: *userEmail,
			}
		}

		if platformName != nil {
			task.Platform = &model.Platform{
				ID:           task.PlatformID,
				Name:         *platformName,
				Slug:         *platformSlug,
				ThumbnailURL: *platformThumbnailURL,
				Type:         *platformType,
				IsActive:     *platformIsActive,
				IsPremium:    *platformIsPremium,
			}
		}

		if fileID != uuid.Nil {
			task.DownloadFiles = append(task.DownloadFiles, model.DownloadFile{
				ID:         fileID,
				DownloadID: downloadID,
				URL:        *url,
				FormatID:   formatID,
				Resolution: resolution,
				Extension:  extension,
				FileSize:   fileSize,
				CreatedAt:  createdAt,
			})
		}

		tasks = append(tasks, &task)
	}

	pagination := model.Pagination{
		CurrentPage: page,
		Limit:       limit,
		TotalItems:  totalItems,
		TotalPages:  int((totalItems + int64(limit) - 1) / int64(limit)),
		HasNext:     int64(offset+limit) < totalItems,
		HasPrev:     page > 1,
	}

	return tasks, pagination, nil
}

func (r *downloadRepository) Update(ctx context.Context, task *model.DownloadTask) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		UPDATE downloads 
		SET status = $1, file_path = $2, format = $3, thumbnail_url = $4, 
			title = $5, file_size = $6, duration = $7, error_message = $8
		WHERE id = $9
	`
	_, err := r.db.Exec(subCtx, query,
		task.Status, task.FilePath, task.Format, task.ThumbnailURL,
		task.Title, task.FileSize, task.Duration, task.ErrorMessage, task.ID)
	return err
}

func (r *downloadRepository) Delete(ctx context.Context, id uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `DELETE FROM downloads WHERE id = $1`
	_, err := r.db.Exec(subCtx, query, id)
	return err
}

func (r *downloadRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `DELETE FROM downloads WHERE id = ANY($1)`
	_, err := r.db.Exec(subCtx, query, ids)
	return err
}

func (r *downloadRepository) AddFile(ctx context.Context, file *model.DownloadFile) error {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 15*time.Second)
	defer cancel()

	query := `
		INSERT INTO download_files (download_id, url, format_id, resolution, extension, file_size, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	now := time.Now()
	err := r.db.QueryRow(subCtx, query,
		file.DownloadID,
		file.URL,
		file.FormatID,
		file.Resolution,
		file.Extension,
		file.FileSize,
		now,
	).Scan(&file.ID)

	file.CreatedAt = now
	return err
}

func (r *downloadRepository) FindOldAndCompleted(ctx context.Context, cutoff time.Time, limit int) ([]*model.DownloadTask, error) {
	subCtx, cancel := contextpool.WithTimeoutIfNone(ctx, 30*time.Second)
	defer cancel()

	query := `
		SELECT id, platform_type, status, created_at
		FROM downloads
		WHERE created_at < $1
		LIMIT $2
	`
	// We only need ID and platform_type for deletion
	// But let's reuse struct.
	var tasks []*model.DownloadTask
	rows, err := r.db.Query(subCtx, query, cutoff, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task model.DownloadTask
		err := rows.Scan(&task.ID, &task.PlatformType, &task.Status, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}
