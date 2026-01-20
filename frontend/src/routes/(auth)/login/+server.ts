import { json } from '@sveltejs/kit';

export const POST = async ({ request, locals }) => {

	try {

		const { credential } = await request.json();

		if (!credential) {
			throw new Error('Credential is required');
		}

		const response = await locals.deps.authService.loginGoogle(credential);

		if (response instanceof Error) {
			throw response;
		}

		if (!response.access_token || !response.user) {
			throw new Error('Login failed: Missing token or user data');
		}

		locals.session?.set('access_token', response.access_token);
		locals.user = response.user;

		return json({
			success: true,
			message: 'Login successful',
			data: {
				full_name: response.user.full_name,
				email: response.user.email,
				avatar: response.user.avatar,
				role: response.user.role,
			}
		}, {
			status: 200
		});

	} catch (error) {
		return json({
			success: false,
			message: error instanceof Error ? error.message : 'Login failed'
		}, {
			status: 400
		});
	}
}
