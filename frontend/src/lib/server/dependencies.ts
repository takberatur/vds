import type { RequestEvent } from '@sveltejs/kit';
import { ApiClientHandler, AuthHelper, QueryHelper, LanguageHelper, PostHelper } from '@/helpers';
import {
	AuthServiceImpl,
	SettingServiceImpl,
	UserServiceImpl,
	PlatformServiceImpl,
	AdminServiceImpl,
	ApplicationServiceImpl,
	DownloadServiceImpl,
	ServerStatusServiceImpl,
	WebServiceImpl,
	SubscriptionServiceImpl,
} from '@/services';

export class Dependencies {
	public readonly apiClient: ApiClient;
	public readonly queryHelper: QueryHelper;
	public readonly languageHelper: LanguageHelper;
	public readonly authHelper: AuthHelper;
	public readonly postHelper: PostHelper;

	public readonly authService: AuthServiceImpl;
	public readonly settingService: SettingServiceImpl;
	public readonly userService: UserServiceImpl;
	public readonly adminService: AdminServiceImpl;
	public readonly platformService: PlatformServiceImpl;
	public readonly applicationService: ApplicationServiceImpl;
	public readonly downloadService: DownloadServiceImpl;
	public readonly serverStatusService: ServerStatusServiceImpl;
	public readonly webService: WebServiceImpl;
	public readonly subscriptionService: SubscriptionServiceImpl;

	constructor(event: RequestEvent) {
		this.apiClient = new ApiClientHandler(event);
		this.queryHelper = new QueryHelper(event);
		this.languageHelper = new LanguageHelper(event);
		this.authHelper = new AuthHelper(event);
		this.postHelper = new PostHelper(event);

		this.authService = new AuthServiceImpl(event, this.apiClient);
		this.settingService = new SettingServiceImpl(event, this.apiClient);
		this.userService = new UserServiceImpl(event, this.apiClient);
		this.adminService = new AdminServiceImpl(event, this.apiClient);
		this.platformService = new PlatformServiceImpl(event, this.apiClient);
		this.applicationService = new ApplicationServiceImpl(event, this.apiClient);
		this.downloadService = new DownloadServiceImpl(event, this.apiClient);
		this.serverStatusService = new ServerStatusServiceImpl(event, this.apiClient);
		this.webService = new WebServiceImpl(event, this.apiClient);
		this.subscriptionService = new SubscriptionServiceImpl(event, this.apiClient);
	}
}
