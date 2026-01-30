import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';

export class AdminServiceImpl extends BaseService implements AdminService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}

	async getDashboardData(query: QueryParams): Promise<PaginatedResult<DashboardResponse>> {
		try {
			const params = new URLSearchParams();
			Object.entries(query).forEach(([key, value]) => {
				if (value !== undefined && value !== null && value !== '') {
					params.append(key, String(value));
				}
			});
			const queryString = params.toString();
			const endpoint = `/protected-admin/dashboard${queryString ? `?${queryString}` : ''}`;

			const response = await this.api.authRequest('GET', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get dashboard data');
			}
			return {
				data: response.data as DashboardResponse[] || [],
				pagination: response.pagination as ApiPagination,
			}
		} catch (error) {
			console.error('Error fetching dashboard data:', error);
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

	async getCookies(): Promise<CookieItem> {
		try {
			const response = await this.api.authRequest<CookieItem>('GET', '/protected-admin/cookies');
			if (!response.success) {
				throw new Error(response.message || 'Failed to get cookies');
			}
			return response.data || {
				lines: '',
				path: '',
				valid: false,
			};
		} catch (error) {
			console.error('Error fetching cookies:', error);
			return {
				lines: '',
				path: '',
				valid: false,
			}
		}
	}

	async updateCookies(cookies: string[]): Promise<boolean> {
		try {
			const response = await this.api.authRequest('PUT', '/protected-admin/cookies', { cookies });
			if (!response.success) {
				throw new Error(response.message || 'Failed to update cookies');
			}
			return response.success;
		} catch (error) {
			console.error('Error updating cookies:', error);
			return false;
		}
	}
}
