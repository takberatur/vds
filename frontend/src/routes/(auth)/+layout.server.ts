import { redirect } from '@sveltejs/kit';

export const load = async ({ locals, url }) => {
	const { user, settings } = locals;

	return {
		user,
		settings
	};
};
