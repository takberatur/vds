import type {
	UpdateProfileSchema,
	UpdatePasswordSchema,
	ResetPasswordSchema,
	ContactSchema,
	RegisterAppSchema,
	UpdateApplicationSchema,
	UpdatePlatformSchema,
	DownloadVideoSchema,
} from '@/utils/schema';

declare global {
	interface Window {
		gc: NodeJS.GCFunction | undefined;
	}

	// ==========================================
	// Database Models (Matches 000001_init_schema.up.sql)
	// ==========================================

	type Role = {
		id: string;
		name: string;
		permissions: Record<string, boolean>; // JSONB
		created_at: string;
	};

	type User = {
		id: string;
		email: string;
		password_hash?: string | null;
		full_name: string;
		avatar_url: string;
		role_id: string | null;
		is_active: boolean;
		last_login_at: string | null;
		created_at: string;
		updated_at: string;
		deleted_at?: string | null;

		// Relations
		role?: Role;
		oauth_providers?: OAuthProvider[];
		downloads?: Download[];
		subscriptions?: Subscription[];
		transactions?: Transaction[];
	};

	type OAuthProvider = {
		id: string;
		user_id: string;
		provider: string; // 'google', 'facebook'
		provider_user_id: string;
		access_token?: string | null;
		refresh_token?: string | null;
		expiry_at?: string | null;
		created_at: string;

		// Relations
		user?: User;
	};

	type Application = {
		id: string;
		name: string;
		package_name: string;
		api_key: string;
		secret_key: string;
		version?: string | null;
		platform: string; // default 'android'
		enable_monetization: boolean;
		enable_admob: boolean;
		enable_unity_ad: boolean;
		enable_start_app: boolean;
		enable_in_app_purchase: boolean;
		admob_ad_unit_id?: string | null;
		unity_ad_unit_id?: string | null;
		start_app_ad_unit_id?: string | null;
		admob_banner_ad_unit_id?: string | null;
		admob_interstitial_ad_unit_id?: string | null;
		admob_native_ad_unit_id?: string | null;
		admob_rewarded_ad_unit_id?: string | null;
		unity_banner_ad_unit_id?: string | null;
		unity_interstitial_ad_unit_id?: string | null;
		unity_native_ad_unit_id?: string | null;
		unity_rewarded_ad_unit_id?: string | null;
		is_active: boolean;
		created_at: string;
		updated_at: string;

		// Relations
		in_app_products?: InAppProduct[];
		subscriptions?: Subscription[];
		downloads?: Download[];
	};

	type InAppProduct = {
		id: string;
		app_id: string;
		product_id?: string | null;
		product_type?: string | null; // 'in_app' | 'subscription'
		sku_code?: string | null;
		title?: string | null;
		description?: string | null;
		price?: number | null;
		currency?: string | null;
		billing_period: string;
		trial_period_days: number;
		is_active: boolean;
		is_featured: boolean;
		sort_order: number;
		features: Record<string, any>; // JSONB
		created_at: string;
		updated_at: string;

		// Relations
		application?: Application;
		subscriptions?: Subscription[];
	};

	type Platform = {
		id: string;
		name: string;
		slug: string;
		type: 'youtube' | 'tiktok' | 'instagram' | 'facebook' | 'twitter' | 'vimeo' | 'dailymotion' | 'rumble' | 'any-video-downloader' | 'youtube-to-mp3' | 'snackvideo';
		thumbnail_url: string;
		url_pattern?: string | null;
		is_active: boolean;
		is_premium: boolean;
		config: Record<string, any>; // JSONB
		created_at: string;

		// Relations
		downloads?: Download[];
	};

	type Download = {
		id: string;
		user_id?: string | null;
		app_id?: string | null;
		platform_id?: string | null;
		original_url: string;
		file_path?: string | null;
		thumbnail_url?: string | null;
		title?: string | null;
		duration?: number | null;
		file_size?: number | null;
		format?: string | null;
		status: 'pending' | 'processing' | 'completed' | 'failed';
		error_message?: string | null;
		formats?: DownloadFormat[] | null;
		ip_address?: string | null;
		created_at: string;
		// Relations
		user?: User;
		application?: Application;
		platform?: Platform;
	};

	type DownloadFormat = {
		url: string;
		filesize?: number | null;
		format_id?: string | null;
		acodec?: string | null;
		vcodec?: string | null;
		ext?: string | null;
		height?: number | null;
		width?: number | null;
		tbr?: number | null;
	}

	type Subscription = {
		id: string;
		user_id?: string | null;
		app_id?: string | null;
		original_transaction_id: string;
		product_id: string;
		purchase_token: string;
		platform: string;
		start_time: string;
		end_time: string;
		status: 'active' | 'expired' | 'canceled';
		auto_renew: boolean;
		created_at: string;
		updated_at: string;

		// Relations
		user?: User;
		transactions?: Transaction[];
	};

	type Transaction = {
		id: string;
		user_id?: string | null;
		app_id?: string | null;
		subscription_id?: string | null;
		amount: number;
		currency: string;
		provider: string; // 'google_play'
		status: 'success' | 'refunded';
		provider_response?: Record<string, any> | null; // JSONB
		created_at: string;

		// Relations
		user?: User;
		subscription?: Subscription;
	};

	type AnalyticsDaily = {
		id: string;
		date: string;
		total_downloads: number;
		total_users: number;
		active_users: number;
		total_revenue: number;
		updated_at: string;
	};

	// ==========================================
	// Settings (Preserved as requested)
	// ==========================================

	type Setting = {
		id: number;
		key: string;
		value: string;
		description: string;
		group_name: string;
		created_at: string;
		updated_at: string;
	};

	type SettingsValue = {
		WEBSITE: SettingWeb;
		EMAIL: SettingEmail;
		SYSTEM: SettingSystem;
		MONETIZE: SettingMonetize;
	};
	type SettingWeb = {
		site_name?: string;
		site_tagline?: string;
		site_description?: string;
		site_keywords?: string;
		site_logo?: string;
		site_favicon?: string;
		site_email?: string;
		site_phone?: string;
		site_url?: string;
		site_created_at?: string;
	};
	type SettingEmail = {
		smtp_enabled?: boolean;
		smtp_service?: string;
		smtp_host?: string;
		smtp_port?: number;
		smtp_user?: string;
		smtp_password?: string;
		from_email?: string;
		from_name?: string;
	};

	type SettingSystem = {
		enable_documentation?: boolean;
		maintenance_mode?: boolean;
		maintenance_message?: string;
		source_logo_favicon: 'local' | 'remote';
		histats_tracking_code?: string;
		google_analytics_code?: string;
	};
	type SettingMonetize = {
		enable_monetize?: boolean;
		type_monetize?: 'adsense' | 'revenuecat' | 'adsterra';
		auto_ad_code?: string;
		popup_ad_code?: string;
		socialbar_ad_code?: string;
		banner_rectangle_ad_code?: string;
		banner_horizontal_ad_code?: string;
		banner_vertical_ad_code?: string;
		native_ad_code?: string;
		direct_link_ad_code?: string;
	};


	// ==========================================
	// Service Interfaces
	// ==========================================
	interface AuthService {
		loginEmail(email: string, password: string): Promise<{ access_token: string, user: User } | Error>;
		loginGoogle(token: string): Promise<{ access_token: string, user: User } | Error>;
		forgotPassword(email: string): Promise<string | Error>;
		resetPassword(request: ResetPasswordSchema): Promise<string | Error>;
	}
	interface SettingService {
		getPublicSettings(): Promise<SettingsValue | Error>;
		getAllSettings(): Promise<Setting[] | Error>;
		updateBulkSetting(settings: { key: string; value: string; description?: string; group_name: string }[]): Promise<void | Error>
		updateFavicon(favicon: File): Promise<string | Error>;
		updateLogo(logo: File): Promise<string | Error>;
	}
	interface UserService {
		getCurrentUser(): Promise<User | Error>;
		updateProfile(request: UpdateProfileSchema): Promise<void | Error>;
		updatePassword(request: UpdatePasswordSchema): Promise<void | Error>;
		updateAvatar(file: File): Promise<string | Error>;
		clientUpdateProfile(request: {
			full_name: string;
			email: string;
		}): Promise<void | Error>
		clientUpdateAvatar(file: File): Promise<string | Error>
		clientUpdatePassword(request: UpdatePasswordSchema): Promise<void | Error>
	}
	interface PlatformService {
		GetPlatforms(query: QueryParams): Promise<PaginatedResult<Platform>>;
		GetPlatformByID(id: string): Promise<Platform | Error>
		GetPlatformByType(type_: string): Promise<Platform | Error>
		UpdatePlatform(data: UpdatePlatformSchema): Promise<void | Error>;
		DeletePlatform(id: string): Promise<void | Error>;
		BulkDeletePlatforms(ids: string[]): Promise<void | Error>;
		UploadThumbnail(platformID: string, file: File): Promise<string | Error>
		GetAll(): Promise<Platform[] | Error>;
		PublicGetPlatformByID(id: string): Promise<Platform | Error>;
		PublicGetPlatformBySlug(slug: string): Promise<Error | Platform>
		PublicGetPlatformByType(type_: string): Promise<Platform | Error>
	}
	interface AdminService {
		getDashboardData(query: QueryParams): Promise<PaginatedResult<DashboardResponse>>
	}
	interface ApplicationService {
		GetApplications(query: QueryParams): Promise<PaginatedResult<Application>>
		create(data: RegisterAppSchema): Promise<string | Error>;
		findByID(id: string): Promise<Application | Error>;
		update(id: string, data: UpdateApplicationSchema): Promise<string | Error>;
		delete(id: string): Promise<string | Error>;
		bulkDelete(ids: string[]): Promise<string | Error>;
	}
	interface DownloadService {
		GetDownloads(query: QueryParams): Promise<PaginatedResult<Download>>
		FindByID(id: number): Promise<Download | Error>
		Delete(id: string): Promise<string | Error>
		BulkDelete(ids: string[]): Promise<string | Error>
	}
	interface ServerStatusService {
		GetServerHealth(): Promise<ServerHealthResponse | null>;
		GetServerLogs(page?: number, limit?: number): Promise<PaginatedResult<ServerLogsResponse> | null>;
		ClearServerLogs(): Promise<void | Error>
	}
	interface WebService {
		Contact(data: ContactSchema): Promise<void | Error>;
		DownloadVideo(data: DownloadVideoSchema): Promise<ApiResponse<Download>>
	}

	// ==========================================
	// Analytics Interfaces
	// ==========================================
	interface DashboardStats {
		total_users: number;
		total_apps: number;
		total_platforms: number;
		total_downloads: number;
		total_subscriptions: number;
		total_transactions: number;
	}
	interface DashboardData {
		stats: DashboardStats;
		analytics: AnalyticsDaily[];
		recent_downloads: Download[];
	}
	interface DashboardResponse {
		data: DashboardData;
		pagination: ApiPagination;
	}
	// ==========================================
	// Server Status Interfaces
	// ==========================================
	interface ServerHealthResponse {
		database: string;
		redis: string;
		time: string;
	}
	interface ServerLogsResponse {
		level: string;
		message: string;
		time: string;
		count?: number;
		duration?: string;
		sql?: string;
		port?: number;
		args?: string[];
		command?: string;
		pipeline_size?: number;
		ip?: string;
		latency?: string;
		method?: string;
		path?: string;
		status?: number;
		user_agent?: string;
		error?: string;
	}
}

export { };
