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
		{ID: uuid.New(), Name: "Youtube", Slug: "youtube", Type: "youtube", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fyoutube.png", Category: "video", URLPattern: stringPtr(`youtube\.com|youtu\.be`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Facebook", Slug: "facebook", Type: "facebook", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Ffacebook.png", Category: "video", URLPattern: stringPtr(`facebook\.com|fb\.watch`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Twitter", Slug: "twitter", Type: "twitter", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Ftwitter.png", Category: "video", URLPattern: stringPtr(`twitter\.com|x\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "TikTok", Slug: "tiktok", Type: "tiktok", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Ftiktok.png", Category: "video", URLPattern: stringPtr(`tiktok\.com|vm\.tiktok\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Instagram", Slug: "instagram", Type: "instagram", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Finstagram.png", Category: "video", URLPattern: stringPtr(`instagram\.com|instagr\.am`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Vimeo", Slug: "vimeo", Type: "vimeo", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fvimeo.png", Category: "video", URLPattern: stringPtr(`vimeo\.com|player\.vimeo\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "DailyMotion", Slug: "dailymotion", Type: "dailymotion", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fdailymotion.png", Category: "video", URLPattern: stringPtr(`dailymotion\.com|dai\.ly`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Rumble", Slug: "rumble", Type: "rumble", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Frumble.png", Category: "video", URLPattern: stringPtr(`rumble\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Any video downloader", Slug: "any-video-downloader", Type: "any-video-downloader", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fany-video.png", Category: "video", URLPattern: stringPtr(`.*`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Snack", Slug: "snackvideo", Type: "snackvideo", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fsnackvideo.png", Category: "video", URLPattern: stringPtr(`snackvideo\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Linkedin", Slug: "linkedin", Type: "linkedin", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Flinkedin.png", Category: "video", URLPattern: stringPtr(`linkedin\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Baidu", Slug: "baidu", Type: "baidu", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fbaidu.png", Category: "video", URLPattern: stringPtr(`baidu\.com`), IsActive: false, IsPremium: false},
		{ID: uuid.New(), Name: "Pinterest", Slug: "pinterest", Type: "pinterest", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fpinterest.png", Category: "video", URLPattern: stringPtr(`pinterest\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Snapchat", Slug: "snapchat", Type: "snapchat", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fsnapchat.png", Category: "video", URLPattern: stringPtr(`snapchat\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Twitch", Slug: "twitch", Type: "twitch", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Ftwitch.png", Category: "video", URLPattern: stringPtr(`twitch\.tv`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Youtube To MP3", Slug: "youtube-to-mp3", Type: "youtube-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fyoutube-mp3.png", Category: "audio", URLPattern: stringPtr(`youtube\.com|youtu\.be`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Facebook To MP3", Slug: "facebook-to-mp3", Type: "facebook-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Ffacebook-mp3.png", Category: "audio", URLPattern: stringPtr(`facebook\.com|fb\.watch`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Twitter To MP3", Slug: "twitter-to-mp3", Type: "twitter-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Ftwitter-mp3.png", Category: "audio", URLPattern: stringPtr(`twitter\.com|x\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "TikTok To MP3", Slug: "tiktok-to-mp3", Type: "tiktok-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Ftiktok-mp3.png", Category: "audio", URLPattern: stringPtr(`tiktok\.com|vm\.tiktok\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Linkedin To MP3", Slug: "linkedin-to-mp3", Type: "linkedin-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Flinkedin-mp3.png", Category: "audio", URLPattern: stringPtr(`linkedin\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Snackvideo To MP3", Slug: "snackvideo-to-mp3", Type: "snackvideo-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fsnackvideo-mp3.png", Category: "audio", URLPattern: stringPtr(`snackvideo\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Twitch To MP3", Slug: "twitch-to-mp3", Type: "twitch-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Ftwitch-mp3.png", Category: "audio", URLPattern: stringPtr(`twitch\.tv`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Baidu To MP3", Slug: "baidu-to-mp3", Type: "baidu-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fbaidu-mp3.png", Category: "audio", URLPattern: stringPtr(`baidu\.com`), IsActive: false, IsPremium: false},
		{ID: uuid.New(), Name: "Pinterest To MP3", Slug: "pinterest-to-mp3", Type: "pinterest-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fpinterest-mp3.png", Category: "audio", URLPattern: stringPtr(`pinterest\.com`), IsActive: true, IsPremium: true},
		{ID: uuid.New(), Name: "Snapchat To MP3", Slug: "snapchat-to-mp3", Type: "snapchat-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fsnapchat-mp3.png", Category: "audio", URLPattern: stringPtr(`snapchat\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Instagram To MP3", Slug: "instagram-to-mp3", Type: "instagram-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Finstagram-mp3.png", Category: "audio", URLPattern: stringPtr(`instagram\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Vimeo To MP3", Slug: "vimeo-to-mp3", Type: "vimeo-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fvimeo-mp3.png", Category: "audio", URLPattern: stringPtr(`vimeo\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Dailymotion To MP3", Slug: "dailymotion-to-mp3", Type: "dailymotion-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Fdailymotion-mp3.png", Category: "audio", URLPattern: stringPtr(`dailymotion\.com`), IsActive: true, IsPremium: false},
		{ID: uuid.New(), Name: "Rumble To MP3", Slug: "rumble-to-mp3", Type: "rumble-to-mp3", ThumbnailURL: "https://console-storage.infrastructures.help/api/v1/buckets/video-downloader/objects/download?preview=true&prefix=platforms%2Frumble-mp3.png", Category: "audio", URLPattern: stringPtr(`rumble\.com`), IsActive: true, IsPremium: false},
	}

	for _, platform := range platforms {
		query := `
			INSERT INTO platforms (id, name, slug, type, thumbnail_url, category, url_pattern, is_active, is_premium)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (slug) DO UPDATE 
			SET name = EXCLUDED.name, slug = EXCLUDED.slug, type = EXCLUDED.type, thumbnail_url = EXCLUDED.thumbnail_url, category = EXCLUDED.category, url_pattern = EXCLUDED.url_pattern, is_active = EXCLUDED.is_active, is_premium = EXCLUDED.is_premium
		`
		_, err := dbPool.Exec(ctx, query, platform.ID, platform.Name, platform.Slug, platform.Type, platform.ThumbnailURL, platform.Category, platform.URLPattern, platform.IsActive, platform.IsPremium)
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
