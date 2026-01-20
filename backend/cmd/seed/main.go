package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/video-downloader-backend/internal/config"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/pkg/utils"
)

func main() {
	cfg := config.LoadConfig()

	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	ctx := context.Background()

	roles := []model.Role{
		{ID: uuid.New(), Name: "admin", Permissions: map[string]bool{"all": true}},
		{ID: uuid.New(), Name: "customer", Permissions: map[string]bool{"read": true, "download": true}},
	}

	for _, role := range roles {
		query := `
			INSERT INTO roles (id, name, permissions, created_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (name) DO UPDATE 
			SET permissions = EXCLUDED.permissions
		`
		_, err := dbPool.Exec(ctx, query, role.ID, role.Name, role.Permissions, time.Now())
		if err != nil {
			log.Printf("Failed to seed role %s: %v", role.Name, err)
		} else {
			log.Printf("Seeded role: %s", role.Name)
		}
	}

	adminEmail := "admin@admin.com"
	adminPassword := "password"
	adminName := "Admin"

	var exists bool
	err = dbPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", adminEmail).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check existing admin: %v", err)
	}

	if !exists {
		hashedPassword, err := utils.HashPassword(adminPassword)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		var roleID uuid.UUID
		err = dbPool.QueryRow(ctx, "SELECT id FROM roles WHERE name = $1", "admin").Scan(&roleID)
		if err != nil {
			log.Fatalf("Failed to get admin role: %v", err)
		}

		query := `
			INSERT INTO users (email, password_hash, full_name, role_id, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err = dbPool.Exec(ctx, query, adminEmail, hashedPassword, adminName, roleID, true, time.Now(), time.Now())
		if err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}
		log.Printf("Admin user created successfully. Email: %s, Password: %s", adminEmail, adminPassword)
	} else {
		log.Println("Admin user already exists. Skipping.")
	}

	platforms := []model.Platform{
		{ID: uuid.New(), Name: "Youtube", Slug: "youtube", Type: "youtube", ThumbnailURL: "https://pngimg.com/d/youtube_button_PNG42.png", URLPattern: stringPtr(`youtube\.com|youtu\.be`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Facebook", Slug: "facebook", Type: "facebook", ThumbnailURL: "https://upload.wikimedia.org/wikipedia/commons/6/6c/Facebook_Logo_2023.png", URLPattern: stringPtr(`facebook\.com|fb\.watch`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Twitter", Slug: "twitter", Type: "twitter", ThumbnailURL: "https://images.freeimages.com/image/large-previews/f35/x-twitter-logo-on-black-circle-5694247.png", URLPattern: stringPtr(`twitter\.com|x\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "TikTok", Slug: "tiktok", Type: "tiktok", ThumbnailURL: "https://i.pinimg.com/originals/0b/db/be/0bdbbef30f3d9833eb35f3befadd4b27.png", URLPattern: stringPtr(`tiktok\.com|vm\.tiktok\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Instagram", Slug: "instagram", Type: "instagram", ThumbnailURL: "https://upload.wikimedia.org/wikipedia/commons/a/a5/Instagram_icon.png", URLPattern: stringPtr(`instagram\.com|instagr\.am`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Vimeo", Slug: "vimeo", Type: "vimeo", ThumbnailURL: "https://static.vecteezy.com/system/resources/previews/023/986/982/non_2x/vimeo-logo-vimeo-logo-transparent-vimeo-icon-transparent-free-free-png.png", URLPattern: stringPtr(`vimeo\.com|player\.vimeo\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "DailyMotion", Slug: "dailymotion", Type: "dailymotion", ThumbnailURL: "https://upload.wikimedia.org/wikipedia/commons/thumb/2/27/Logo_dailymotion.png/1200px-Logo_dailymotion.png", URLPattern: stringPtr(`dailymotion\.com|dai\.ly`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Rumble", Slug: "rumble", Type: "rumble", ThumbnailURL: "https://companieslogo.com/img/orig/RUM-79ca46cb.png?t=1720244493", URLPattern: stringPtr(`rumble\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Any video downloader", Slug: "any-video-downloader", Type: "any-video-downloader", ThumbnailURL: "https://dl.memuplay.com/new_market/img/free.download.allvideodownloader.privatebrowser.icon.2024-02-22-09-37-13.png", URLPattern: stringPtr(`.*`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Youtube To MP3", Slug: "youtube-to-mp3", Type: "youtube-to-mp3", ThumbnailURL: "https://upload.wikimedia.org/wikipedia/commons/d/d8/YouTubeMusic_Logo.png", URLPattern: stringPtr(`youtube\.com|youtu\.be`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Snack", Slug: "snackvideo", Type: "snackvideo", ThumbnailURL: "https://cdn-www.bluestacks.com/bs-images/4xLfCP15gq8Pq2VYbgv98heZoZceGP_LgCXN0abjPpWtmOCbyhUhh5tH0S5pw1TQssY.png", URLPattern: stringPtr(`snackvideo\.com`), IsActive: true, IsPremium: false},
	}

	for _, platform := range platforms {
		query := `
			INSERT INTO platforms (id, name, slug, type, thumbnail_url, url_pattern, is_active, is_premium)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (slug) DO UPDATE 
			SET name = EXCLUDED.name, slug = EXCLUDED.slug, type = EXCLUDED.type, thumbnail_url = EXCLUDED.thumbnail_url, url_pattern = EXCLUDED.url_pattern, is_active = EXCLUDED.is_active, is_premium = EXCLUDED.is_premium
		`
		_, err := dbPool.Exec(ctx, query, platform.ID, platform.Name, platform.Slug, platform.Type, platform.ThumbnailURL, platform.URLPattern, platform.IsActive, platform.IsPremium)
		if err != nil {
			log.Printf("Failed to seed platform %s: %v", platform.Name, err)
		} else {
			log.Printf("Seeded platform: %s", platform.Name)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
