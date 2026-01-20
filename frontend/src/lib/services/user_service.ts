import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';
import type { UpdateProfileSchema, UpdatePasswordSchema } from '@/utils/schema';

export class UserServiceImpl extends BaseService implements UserService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}

	async getCurrentUser(): Promise<User | Error> {
		try {
			const response = await this.api.authRequest<User>('GET', '/web-client/protected-web/users/current');

			if (!response.success) {
				throw new Error(response.message || 'Failed to get current user');
			}
			if (!response.data) {
				throw new Error('User data is missing in response');
			}
			return response.data;
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to get current user');
		}
	}

	async updateProfile(request: UpdateProfileSchema): Promise<void | Error> {
		try {
			const response = await this.api.authRequest<void>('PUT', '/protected-admin/users/profile', request);
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to update profile');
			}
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to update profile');
		}
	}
	async updatePassword(request: UpdatePasswordSchema): Promise<void | Error> {
		try {
			const response = await this.api.authRequest<void>('PUT', '/protected-admin/users/password', request);
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to update password');
			}
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to update password');
		}
	}
	async updateAvatar(file: File): Promise<string | Error> {
		try {
			const formData = new FormData();
			formData.append('avatar', file);

			const response = await this.api.multipartAuthRequest<{ avatar_url: string }>('POST', '/protected-admin/users/avatar', formData);
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to update avatar');
			}
			return response.data?.avatar_url || '';
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to update avatar');
		}
	}

	async logout(): Promise<void | Error> {
		try {

			const response = await this.api.publicRequest<void>('POST', '/public-admin/auth/logout');

			if (!response.success) {
				console.warn('Backend logout failed', response.message);
			}
		} catch (error) {
			console.error('Logout error', error);
		}
	}

	async clientUpdateProfile(request: UpdateProfileSchema): Promise<void | Error> {
		try {
			const response = await this.api.authRequest<void>('PUT', '/web-client/protected-web/users/profile', request);
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to update profile');
			}
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to update profile');
		}
	}
	async clientUpdatePassword(request: UpdatePasswordSchema): Promise<void | Error> {
		try {
			const response = await this.api.authRequest<void>('PUT', '/web-client/protected-web/users/password', request);
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to update password');
			}
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to update password');
		}
	}
	async clientUpdateAvatar(file: File): Promise<string | Error> {
		try {
			const formData = new FormData();
			formData.append('avatar', file);

			const response = await this.api.multipartAuthRequest<{ avatar_url: string }>('POST', '/web-client/protected-web/users/avatar', formData);
			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to update avatar');
			}
			return response.data?.avatar_url || '';
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to update avatar');
		}
	}
}
