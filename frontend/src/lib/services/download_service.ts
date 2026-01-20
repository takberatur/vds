import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';

export class DownloadServiceImpl extends BaseService implements DownloadService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}

	async GetDownloads(query: QueryParams): Promise<PaginatedResult<Download>> {
		try {
			const params = new URLSearchParams();
			Object.entries(query).forEach(([key, value]) => {
				if (value !== undefined && value !== null && value !== '') {
					params.append(key, String(value));
				}
			});
			const queryString = params.toString();
			const endpoint = `/protected-admin/downloads${queryString ? `?${queryString}` : ''}`;


			const response = await this.api.authRequest('GET', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get downloads');
			}
			return {
				data: response.data as Download[] || [],
				pagination: response.pagination as ApiPagination,
			}
		} catch (error) {
			console.error('Error getting downloads:', error);
			console.error('Error fetching downloads:', error);
			return {
				data: [],
				pagination: {
					current_page: 0,
					limit: 0,
					total_items: 0,
					total_pages: 0,
					has_prev: false,
					has_next: false,
				},
			};
		}
	}

	async FindByID(id: number): Promise<Download | Error> {
		try {
			const endpoint = `/protected-admin/downloads/${id}`;
			const response = await this.api.authRequest('GET', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get download');
			}
			return response.data as Download || null;
		} catch (error) {
			console.error('Error getting download:', error);
			console.error('Error fetching download:', error);
			return error instanceof Error ? error : new Error('Failed to get download');
		}
	}

	async Delete(id: string): Promise<string | Error> {
		try {
			const response = await this.api.authRequest('DELETE', `/protected-admin/downloads/${id}`);
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete download');
			}
			return response.message || 'Download deleted successfully';
		} catch (error) {
			console.error('Error deleting download:', error);
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}

	async BulkDelete(ids: string[]): Promise<string | Error> {
		try {
			const response = await this.api.authRequest('DELETE', `/protected-admin/downloads/bulk`, { ids });
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete downloads');
			}
			return response.message || 'Downloads deleted successfully';
		} catch (error) {
			console.error('Error deleting downloads:', error);
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
}
