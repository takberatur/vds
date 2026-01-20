import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';
import type { UpdateApplicationSchema, RegisterAppSchema } from '@/utils/schema';

export class ApplicationServiceImpl extends BaseService implements ApplicationService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}
	async GetApplications(query: QueryParams): Promise<PaginatedResult<Application>> {
		try {
			const params = new URLSearchParams();
			Object.entries(query).forEach(([key, value]) => {
				if (value !== undefined && value !== null && value !== '') {
					params.append(key, String(value));
				}
			});
			const queryString = params.toString();
			const endpoint = `/protected-admin/applications${queryString ? `?${queryString}` : ''}`;


			const response = await this.api.authRequest('GET', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get applications');
			}
			return {
				data: response.data as Application[] || [],
				pagination: response.pagination as ApiPagination,
			}
		} catch (error) {
			console.error('Error getting applications:', error);
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

	async create(data: RegisterAppSchema): Promise<string | Error> {
		try {
			const response = await this.api.authRequest('POST', '/protected-admin/applications', data);
			if (!response.success) {
				throw new Error(response.message || 'Failed to create application');
			}
			return response.message || 'Application created successfully';
		} catch (error) {
			console.error('Error creating application:', error);
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}

	async findByID(id: string): Promise<Application | Error> {
		try {
			const response = await this.api.authRequest<Application>('GET', `/protected-admin/applications/${id}`);
			if (!response.success) {
				throw new Error(response.message || 'Failed to find application');
			}
			if (!response.data) {
				throw new Error('No application data found');
			}
			return response.data
		} catch (error) {
			console.error('Error finding application:', error);
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}

	async update(id: string, data: UpdateApplicationSchema): Promise<string | Error> {
		try {
			const response = await this.api.authRequest('PUT', `/protected-admin/applications/${id}`, data);
			if (!response.success) {
				throw new Error(response.message || 'Failed to update application');
			}
			return response.message || 'Application updated successfully';
		} catch (error) {
			console.error('Error updating application:', error);
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}

	async delete(id: string): Promise<string | Error> {
		try {
			const response = await this.api.authRequest('DELETE', `/protected-admin/applications/${id}`);
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete application');
			}
			return response.message || 'Application deleted successfully';
		} catch (error) {
			console.error('Error deleting application:', error);
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}

	async bulkDelete(ids: string[]): Promise<string | Error> {
		try {
			const response = await this.api.authRequest('DELETE', `/protected-admin/applications/bulk`, { ids });
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete applications');
			}
			return response.message || 'Applications deleted successfully';
		} catch (error) {
			console.error('Error deleting applications:', error);
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}
}
