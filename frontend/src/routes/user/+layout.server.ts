import { redirect } from '@sveltejs/kit';
import { localizeHref } from '@/paraglide/runtime';

export const load = async ({ locals, url }) => {
	const { user, settings } = locals;

	if (!user) {
		throw redirect(302, localizeHref('/login'));
	}

	return {
		user,
		settings
	};
};
