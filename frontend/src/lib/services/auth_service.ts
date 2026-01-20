import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';
import type { ResetPasswordSchema } from '@/utils/schema';

export class AuthServiceImpl extends BaseService implements AuthService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}

	async loginEmail(email: string, password: string): Promise<{ access_token: string, user: User } | Error> {
		try {
			const response = await this.api.publicRequest<{ access_token: string, user: User }>('POST', '/public-admin/auth/email', {
				email,
				password
			});



			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to login');
			}
			if (!response.data) {
				throw new Error(response.error?.message || response.message || 'Failed to login');
			}

			return response.data;
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to login');
		}

	}
	async loginGoogle(token: string): Promise<{ access_token: string, user: User } | Error> {
		try {
			const response = await this.api.publicRequest<{ access_token: string, user: User }>('POST', '/public-admin/auth/google', {
				credential: token
			});

			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to login');
			}
			if (!response.data) {
				throw new Error(response.error?.message || response.message || 'Failed to login');
			}

			return response.data;
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to login');
		}
	}

	async forgotPassword(email: string): Promise<string | Error> {
		try {
			const response = await this.api.publicRequest('POST', '/public-admin/auth/forgot-password', {
				email
			});

			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to send reset email');
			}

			return response.message || 'If your email is registered, you will receive a reset link';
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to send reset email');
		}
	}

	async resetPassword(data: ResetPasswordSchema): Promise<string | Error> {
		try {
			const response = await this.api.publicRequest('POST', '/public-admin/auth/reset-password', {
				token: data.token,
				new_password: data.new_password, // Changed from password to new_password to match backend
			});

			if (!response.success) {
				throw new Error(response.error?.message || response.message || 'Failed to reset password');
			}

			return response.message || 'Password reset successfully';
		} catch (error) {
			return error instanceof Error ? error : new Error('Failed to reset password');
		}
	}
}
