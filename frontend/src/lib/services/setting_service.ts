
import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';

export class SettingServiceImpl extends BaseService implements SettingService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}

	async getPublicSettings(): Promise<SettingsValue | Error> {
		try {
			const response = await this.api.publicRequest<SettingsValue>('GET', '/public-admin/settings/public');
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to fetch public settings');
			}
			if (!response.data) {
				throw new Error(response.error?.message || response.message || 'Failed to fetch public settings');
			}
			return response.data;
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to fetch public settings');
		}
	}


	async getAllSettings(): Promise<Setting[] | Error> {
		try {
			const response = await this.api.authRequest<Setting[]>('GET', '/protected-admin/settings');
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to fetch settings');
			}
			if (!response.data) {
				throw new Error(response.error?.message || response.message || 'Failed to fetch settings');
			}
			return response.data;
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to fetch settings');
		}
	}


	async updateBulkSetting(settings: { key: string; value: string; description?: string; group_name: string }[]): Promise<void | Error> {
		try {
			const response = await this.api.authRequest<void>('PUT', `/protected-admin/settings/bulk`, settings);
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to update setting');
			}
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to update setting');
		}
	}

	async updateFavicon(favicon: File): Promise<string | Error> {
		try {
			const formData = new FormData();
			formData.append('file', favicon);
			formData.append('key', 'site_favicon');

			const response = await this.api.multipartAuthRequest<{ url: string }>('POST', `/protected-admin/settings/upload`, formData);
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to update favicon');
			}
			return response.data?.url || '';
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to update favicon');
		}
	}

	async updateLogo(logo: File): Promise<string | Error> {
		try {
			const formData = new FormData();
			formData.append('file', logo);
			formData.append('key', 'site_logo');

			const response = await this.api.multipartAuthRequest<{ url: string }>('POST', `/protected-admin/settings/upload`, formData);
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to update logo');
			}
			return response.data?.url || '';
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to update logo');
		}
	}
}
