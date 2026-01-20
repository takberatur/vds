import { redirect } from '@sveltejs/kit';
import { localizeHref } from '@/paraglide/runtime';

export const load = async ({ locals, url }) => {
	const { user, settings } = locals;

	if (!user) {
		throw redirect(302, localizeHref('/login'));
	}
	if (user.role?.name !== 'admin') {
		throw redirect(302, localizeHref('/user'));
	}

	return {
		user,
		settings
	};
};
