-- Enable pgcrypto for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users & Roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE, -- 'admin', 'customer'
    permissions JSONB DEFAULT '{}', -- List of permissions e.g. {"can_download_premium": true}
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- Nullable for OAuth only users
    full_name VARCHAR(100),
    avatar_url TEXT,
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE -- Soft delete
);

-- OAuth Providers (Google, Facebook)
CREATE TABLE oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- 'google', 'facebook'
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expiry_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider, provider_user_id)
);

-- Registered Applications (For Android Apps)
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    package_name VARCHAR(255) NOT NULL UNIQUE, -- e.g. com.example.videodownloader
    api_key VARCHAR(64) UNIQUE NOT NULL, -- Generated secure key for API access
    secret_key VARCHAR(128) NOT NULL, -- For signing requests if needed
    version VARCHAR(20),
    platform VARCHAR(20) DEFAULT 'android',
    enable_monetization BOOLEAN DEFAULT FALSE,
    enable_admob BOOLEAN DEFAULT FALSE,
    enable_unity_ad BOOLEAN DEFAULT FALSE,
    enable_start_app BOOLEAN DEFAULT FALSE,
    enable_in_app_purchase BOOLEAN DEFAULT FALSE,
    admob_ad_unit_id TEXT,
    unity_ad_unit_id TEXT,
    start_app_ad_unit_id TEXT,
    admob_banner_ad_unit_id TEXT,
    admob_interstitial_ad_unit_id TEXT,
    admob_native_ad_unit_id TEXT,
    admob_rewarded_ad_unit_id TEXT,
    unity_banner_ad_unit_id TEXT,
    unity_interstitial_ad_unit_id TEXT,
    unity_native_ad_unit_id TEXT,
    unity_rewarded_ad_unit_id TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Supported Downloader Platforms (Youtube, TikTok, etc.)
CREATE TABLE platforms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE, -- 'youtube', 'tiktok', 'instagram'
    slug VARCHAR(50) NOT NULL UNIQUE,
		type VARCHAR(20) NOT NULL DEFAULT 'youtube' CHECK (type IN ('youtube', 'tiktok', 'instagram', 'facebook', 'twitter', 'vimeo', 'dailymotion', 'rumble', 'any-video-downloader', 'snackvideo', 'linkedin', 'baidu', 'pinterest', 'snapchat', 'twitch', 'youtube-to-mp3', 'facebook-to-mp3', 'tiktok-to-mp3', 'linkedin-to-mp3', 'snackvideo-to-mp3', 'twitch-to-mp3', 'baidu-to-mp3', 'pinterest-to-mp3', 'snapchat-to-mp3', 'instagram-to-mp3', 'twitter-to-mp3', 'vimeo-to-mp3', 'dailymotion-to-mp3', 'rumble-to-mp3')),
    thumbnail_url TEXT NOT NULL,
		category VARCHAR(50) NOT NULL DEFAULT 'video' CHECK (category IN ('video', 'audio', 'image', 'document', 'other')),
    url_pattern VARCHAR(255), -- Regex to match URL
    is_active BOOLEAN DEFAULT TRUE,
    is_premium BOOLEAN DEFAULT FALSE, -- If true, only premium users can download
    config JSONB DEFAULT '{}', -- Specific config like cookies, headers, etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Downloads History
CREATE TABLE downloads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    app_id UUID REFERENCES applications(id) ON DELETE SET NULL,
    platform_id UUID NOT NULL REFERENCES platforms(id) ON DELETE SET NULL,
		platform_type TEXT NOT NULL,
    original_url TEXT NOT NULL,
    file_path TEXT, -- Local path or S3 URL
    thumbnail_url TEXT,
    title TEXT,
    duration INT, -- in seconds
    file_size BIGINT, -- in bytes
		encrypted_data BYTEA, -- Encoded data for download files encode string to avoid special characters
    format VARCHAR(20), -- mp4, mp3
    status VARCHAR(20) NOT NULL, -- 'pending', 'processing', 'completed', 'failed'
    error_message TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE download_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    download_id UUID NOT NULL REFERENCES downloads(id) ON DELETE CASCADE,
    url TEXT NOT NULL, -- MinIO / S3 URL
    format_id VARCHAR(50), -- yt-dlp format id
    resolution VARCHAR(50), -- e.g. 1920x1080
    extension VARCHAR(10), -- mp4, webm, mp3
    file_size BIGINT,
		encrypted_data BYTEA,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- In-App Purchases & Subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		app_id UUID REFERENCES applications(id) ON DELETE CASCADE,
    original_transaction_id VARCHAR(255) UNIQUE NOT NULL, -- From Play Store
    product_id VARCHAR(100) NOT NULL, -- e.g. 'premium_monthly'
    purchase_token TEXT NOT NULL,
    platform VARCHAR(20) DEFAULT 'android',
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'active', 'expired', 'canceled'
    auto_renew BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		app_id UUID REFERENCES applications(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    provider VARCHAR(50) DEFAULT 'google_play',
    status VARCHAR(20) NOT NULL, -- 'success', 'refunded'
    provider_response JSONB, -- Full receipt data
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Settings (Dynamic Config)
CREATE TABLE settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(255) UNIQUE NOT NULL,
    value TEXT,
    description TEXT,
    group_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Analytics (Simple Aggregation)
CREATE TABLE analytics_daily (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL UNIQUE,
    total_downloads INT DEFAULT 0,
    total_users INT DEFAULT 0,
    active_users INT DEFAULT 0,
    total_revenue DECIMAL(10, 2) DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE in_app_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    product_id TEXT,
    product_type TEXT,
    sku_code VARCHAR(100),
    title VARCHAR(255),
    description TEXT,
    price DECIMAL(10,2),
    currency VARCHAR(10),
    billing_period VARCHAR(50), -- P1M, P1Y, etc
    trial_period_days INT DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_featured BOOLEAN NOT NULL DEFAULT false,
    sort_order INT DEFAULT 0,
    features JSONB DEFAULT '{}', -- {"vpn": true, "ads_removal": true, "storage": "5GB"}
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT unique_product UNIQUE (app_id, product_id)
);

-- Seed Initial Data
-- Insert roles with fixed UUIDs or let them be generated.
-- We will rely on name lookups in the application, but for initial seed, we can just insert them.
INSERT INTO roles (name, permissions) VALUES 
('admin', '{"all": true}'),
('customer', '{"download_basic": true}')
ON CONFLICT (name) DO NOTHING;

-- Seed Default Settings
INSERT INTO settings (key, value, description, group_name) VALUES
-- WEBSITE Group
('site_name', 'Simontok', 'Website Name', 'WEBSITE'),
('site_tagline', 'Download Full HD Videos Without Watermark for Free', 'Website Tagline', 'WEBSITE'),
('site_description', 'Simontok helps you download videos from any site without watermarks for free in MP4 or MP3 online. Fast, HD quality, and just enter the link. Try it now!', 'Website Meta Description', 'WEBSITE'),
('site_keywords', 'Simontok, video downloader, free downloader', 'Website Meta Keywords', 'WEBSITE'),
('site_logo', '', 'URL to Website Logo', 'WEBSITE'),
('site_favicon', '', 'URL to Website Favicon', 'WEBSITE'),
('site_email', 'admin@example.com', 'Contact Email', 'WEBSITE'),
('site_phone', '', 'Contact Phone', 'WEBSITE'),
('site_url', 'http://localhost:3000', 'Website Public URL', 'WEBSITE'),

-- EMAIL Group
('smtp_enabled', 'false', 'Enable SMTP', 'EMAIL'),
('smtp_service', 'gmail', 'SMTP Service Provider', 'EMAIL'),
('smtp_host', 'smtp.gmail.com', 'SMTP Host', 'EMAIL'),
('smtp_port', '587', 'SMTP Port', 'EMAIL'),
('smtp_user', '', 'SMTP Username', 'EMAIL'),
('smtp_password', '', 'SMTP Password', 'EMAIL'),
('from_email', 'noreply@example.com', 'From Email Address', 'EMAIL'),
('from_name', 'Simontok', 'From Name', 'EMAIL'),

-- SYSTEM Group
('enable_documentation', 'false', 'Enable Documentation', 'SYSTEM'),
('maintenance_mode', 'false', 'Enable Maintenance Mode', 'SYSTEM'),
('maintenance_message', 'We are currently performing maintenance. Please check back later.', 'Maintenance Message', 'SYSTEM'),
('source_logo_favicon', 'local', 'Source of Logo/Favicon (local/remote)', 'SYSTEM'),
('histats_tracking_code', '', 'Histats Tracking Code', 'SYSTEM'),
('google_analytics_code', '', 'Google Analytics Code', 'SYSTEM'),
('play_store_app_url', '', 'Google Play Store App URL', 'SYSTEM'),
('app_store_app_url', '', 'Apple App Store App URL', 'SYSTEM'),

-- MONETIZE Group
('enable_monetize', 'false', 'Enable Monetization Features', 'MONETIZE'),
('type_monetize', 'adsense', 'Monetization Type (adsense, revenuecat, adsterra)', 'MONETIZE'),
('publisher_id', '', 'Adsense Publisher ID', 'MONETIZE'),
('enable_popup_ad', 'false', 'Enable Popup Ad', 'MONETIZE'),
('enable_socialbar_ad', 'false', 'Enable Social Bar Ad', 'MONETIZE'),
('popup_ad_code', '', 'Popup Ad Code', 'MONETIZE'),
('socialbar_ad_code', '', 'Social Bar Ad Code', 'MONETIZE'),
('auto_ad_code', '', 'Auto Ad Code', 'MONETIZE'),
('banner_rectangle_ad_code', '', 'Banner Rectangle Ad Code', 'MONETIZE'),
('banner_horizontal_ad_code', '', 'Banner Horizontal Ad Code', 'MONETIZE'),
('banner_vertical_ad_code', '', 'Banner Vertical Ad Code', 'MONETIZE'),
('native_ad_code', '', 'Native Ad Code', 'MONETIZE'),
('direct_link_ad_code', '', 'Direct Link Ad Code', 'MONETIZE')
ON CONFLICT (key) DO NOTHING;
