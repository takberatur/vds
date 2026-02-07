import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';

export class SubscriptionServiceImpl extends BaseService implements SubscriptionService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}

	async FindAll(query: QueryParams): Promise<PaginatedResult<Subscription>> {
		try {
			const params = new URLSearchParams();
			Object.entries(query).forEach(([key, value]) => {
				if (value !== undefined && value !== null && value !== '') {
					params.append(key, String(value));
				}
			});
			const queryString = params.toString();
			const endpoint = `/protected-admin/subscriptions${queryString ? `?${queryString}` : ''}`;


			const response = await this.api.authRequest('GET', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get subscriptions');
			}
			return {
				data: response.data as Subscription[] || [],
				pagination: response.pagination as ApiPagination,
			}
		} catch (error) {
			console.error('Error getting subscriptions:', error);
			console.error('Error fetching subscriptions:', error);
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
	async FindByID(id: string): Promise<Subscription | Error> {
		try {
			const endpoint = `/protected-admin/subscriptions/${id}`;
			const response = await this.api.authRequest('GET', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to get subscription');
			}
			return response.data as Subscription;
		} catch (error) {
			console.error('Error fetching subscription:', error);
			return error as Error;
		}
	}
	async BulkDelete(ids: string[]): Promise<void | Error> {
		try {
			const endpoint = `/protected-admin/subscriptions/bulk`;
			const response = await this.api.authRequest('DELETE', endpoint, { ids });
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete subscriptions');
			}
		} catch (error) {
			console.error('Error deleting subscriptions:', error);
			console.error('Error bulk deleting subscriptions:', error);
			return error as Error;
		}
	}
	async Delete(id: string): Promise<void | Error> {
		try {
			const endpoint = `/protected-admin/subscriptions/${id}`;
			const response = await this.api.authRequest('DELETE', endpoint);
			if (!response.success) {
				throw new Error(response.message || 'Failed to delete subscription');
			}
		} catch (error) {
			console.error('Error deleting subscription:', error);
			console.error('Error deleting subscription:', error);
			return error as Error;
		}
	}
}

