import type { RequestEvent } from '@sveltejs/kit';
import { ApiClientHandler, AuthHelper, QueryHelper, LanguageHelper } from '@/helpers';
import {
	AuthServiceImpl,
	SettingServiceImpl,
	UserServiceImpl,
	PlatformServiceImpl,
	AdminServiceImpl,
	ApplicationServiceImpl,
	DownloadServiceImpl,
	ServerStatusServiceImpl,
	WebServiceImpl
} from '@/services';

export class Dependencies {
	public readonly apiClient: ApiClient;
	public readonly queryHelper: QueryHelper;
	public readonly languageHelper: LanguageHelper;
	public readonly authHelper: AuthHelper;

	public readonly authService: AuthServiceImpl;
	public readonly settingService: SettingServiceImpl;
	public readonly userService: UserServiceImpl;
	public readonly adminService: AdminServiceImpl;
	public readonly platformService: PlatformServiceImpl;
	public readonly applicationService: ApplicationServiceImpl;
	public readonly downloadService: DownloadServiceImpl;
	public readonly serverStatusService: ServerStatusServiceImpl;
	public readonly webService: WebServiceImpl;

	constructor(event: RequestEvent) {
		this.apiClient = new ApiClientHandler(event);
		this.queryHelper = new QueryHelper(event);
		this.languageHelper = new LanguageHelper(event);
		this.authHelper = new AuthHelper(event);

		this.authService = new AuthServiceImpl(event, this.apiClient);
		this.settingService = new SettingServiceImpl(event, this.apiClient);
		this.userService = new UserServiceImpl(event, this.apiClient);
		this.adminService = new AdminServiceImpl(event, this.apiClient);
		this.platformService = new PlatformServiceImpl(event, this.apiClient);
		this.applicationService = new ApplicationServiceImpl(event, this.apiClient);
		this.downloadService = new DownloadServiceImpl(event, this.apiClient);
		this.serverStatusService = new ServerStatusServiceImpl(event, this.apiClient);
		this.webService = new WebServiceImpl(event, this.apiClient);
	}
}
