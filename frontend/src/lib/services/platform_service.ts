import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';
import type { UpdatePlatformSchema } from '@/utils/schema';

export class PlatformServiceImpl extends BaseService implements PlatformService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}
	async GetPlatforms(query: QueryParams): Promise<PaginatedResult<Platform>> {
		try {
			const params = new URLSearchParams();
			Object.entries(query).forEach(([key, value]) => {
				if (value !== undefined && value !== null && value !== '') {
					params.append(key, String(value));
				}
			});
			const queryString = params.toString();
			const endpoint = `/protected-admin/platforms${queryString ? `?${queryString}` : ''}`;

			const response = await this.api.authRequest('GET', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get platforms');
			}
			return {
				data: response.data as Platform[] || [],
				pagination: response.pagination as ApiPagination,
			}
		} catch (error) {
			console.error('Error fetching platforms:', error);
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
	async GetPlatformByID(id: string): Promise<Platform | Error> {
		try {
			const response = await this.api.authRequest<Platform>('GET', `/protected-admin/platforms/${id}`);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get platform');
			}
			if (!response.data) {
				throw new Error('No platform data found');
			}
			return response.data
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async GetPlatformByType(type_: string): Promise<Platform | Error> {
		try {
			const response = await this.api.authRequest<Platform>('GET', `/protected-admin/platforms/type/${type_}`);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get platform');
			}
			if (!response.data) {
				throw new Error('No platform data found');
			}
			return response.data
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async UpdatePlatform(data: UpdatePlatformSchema): Promise<void | Error> {
		try {
			const response = await this.api.authRequest('PUT', `/protected-admin/platforms/${data.id}`, data);
			if (!response.success) {
				throw new Error(response.message || 'Failed to update platform');
			}
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async DeletePlatform(id: string): Promise<void | Error> {
		try {
			const response = await this.api.authRequest('DELETE', '/protected-admin/platforms/' + id);
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete platform');
			}
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async BulkDeletePlatforms(ids: string[]): Promise<void | Error> {
		try {
			const response = await this.api.authRequest('DELETE', '/protected-admin/platforms/bulk', { ids });
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete platforms');
			}
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async UploadThumbnail(platformID: string, file: File): Promise<string | Error> {
		try {
			const formData = new FormData();
			formData.append('thumbnail', file);

			const response = await this.api.multipartAuthRequest<{ thumbnail_url: string }>('POST', `/protected-admin/platforms/thumbnail/${platformID}`, formData);
			if (!response.success) {
				throw new Error(response.message || 'Failed to upload thumbnail');
			}
			return response.data?.thumbnail_url || '';
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async GetAll(): Promise<Platform[] | Error> {
		try {
			const response = await this.api.publicRequest<Platform[]>('GET', '/web-client/platforms');
			if (!response.success) {
				throw new Error(response.message || 'Failed to get platforms');
			}
			return response.data || [];
		} catch (error) {
			console.error('Error fetching platforms:', error);
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async PublicGetPlatformByID(id: string): Promise<Platform | Error> {
		try {
			const response = await this.api.publicRequest<Platform>('GET', `/web-client/platforms/${id}`);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get platform');
			}
			if (!response.data) {
				throw new Error('No platform data found');
			}
			return response.data
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async PublicGetPlatformBySlug(slug: string): Promise<Platform | Error> {
		try {
			const response = await this.api.publicRequest<Platform>('GET', `/web-client/platforms/slug/${slug}`);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get platform');
			}
			if (!response.data) {
				throw new Error('No platform data found');
			}
			return response.data
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async PublicGetPlatformByType(type_: string): Promise<Platform | Error> {
		try {
			const response = await this.api.publicRequest<Platform>('GET', `/web-client/platforms/type/${type_}`);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get platform');
			}
			if (!response.data) {
				throw new Error('No platform data found');
			}
			return response.data
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
	async PublicGetPlatformsByCategory(category: string): Promise<Platform[] | Error> {
		try {
			const response = await this.api.publicRequest<Platform[]>('GET', `/web-client/platforms/category/${category}`);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get platforms');
			}
			return response.data || [];
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
}
