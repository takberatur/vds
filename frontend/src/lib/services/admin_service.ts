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
			const data: any = response.data || {};
			const linesRaw = data.lines;
			const lines = Array.isArray(linesRaw)
				? linesRaw
				: typeof linesRaw === 'string'
					? linesRaw.split('\n')
					: [];

			return {
				lines,
				path: typeof data.path === 'string' ? data.path : '',
				valid: Boolean(data.valid)
			};
		} catch (error) {
			console.error('Error fetching cookies:', error);
			return {
				lines: [],
				path: '',
				valid: false,
			}
		}
	}

	async updateCookies(content: string): Promise<CookieItem | null> {
		try {
			const response = await this.api.authRequest('PUT', '/protected-admin/cookies', { content });
			if (!response.success) {
				throw new Error(response.message || 'Failed to update cookies');
			}
			return await this.getCookies();
		} catch (error) {
			console.error('Error updating cookies:', error);
			return null;
		}
	}

	async FindUserAll(query: QueryParams): Promise<PaginatedResult<User>> {
		try {
			const params = new URLSearchParams();
			Object.entries(query).forEach(([key, value]) => {
				if (value !== undefined && value !== null && value !== '') {
					params.append(key, String(value));
				}
			});
			const queryString = params.toString();
			const endpoint = `/protected-admin/users/search${queryString ? `?${queryString}` : ''}`;

			const response = await this.api.authRequest('GET', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get users');
			}
			return {
				data: response.data as User[] || [],
				pagination: response.pagination as ApiPagination,
			}
		} catch (error) {
			console.error('Error fetching users:', error);
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

	async FindUserByID(id: string): Promise<User | Error> {
		try {
			const endpoint = `/protected-admin/users/find/${id}`;
			const response = await this.api.authRequest('GET', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get user');
			}
			return response.data as User;
		} catch (error) {
			console.error('Error fetching user:', error);
			return error as Error;
		}
	}

	async BulkDeleteUser(ids: string[]): Promise<void | Error> {
		try {
			const endpoint = `/protected-admin/users/bulk`;
			const response = await this.api.authRequest('DELETE', endpoint, { ids });
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete users');
			}
			return response.data as void;
		} catch (error) {
			console.error('Error deleting users:', error);
			return error as Error;
		}
	}

	async DeleteUser(id: string): Promise<void | Error> {
		try {
			const endpoint = `/protected-admin/users/${id}`;
			const response = await this.api.authRequest('DELETE', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete user');
			}
			return response.data as void;
		} catch (error) {
			console.error('Error deleting user:', error);
			return error as Error;
		}
	}
}
