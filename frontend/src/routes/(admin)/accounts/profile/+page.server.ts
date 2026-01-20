import { superValidate } from 'sveltekit-superforms';
import { updateProfileSchema } from '$lib/utils/schema';
import { fail } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';
import { defaultMetaTags } from '@/utils/meta-tags.js';

export const load = async ({ locals, url }) => {
	const { settings, user } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Profile - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
			tagline: settings?.WEBSITE.site_tagline || '',
			description: settings?.WEBSITE.site_description || '',
			keywords: settings?.WEBSITE.site_keywords?.split(', ') || [''],
			robots: 'noindex, nofollow',
			canonical: defaultOrigin,
			graph_type: 'website',
			use_tagline: false
		},
		settings
	);

	const form = await superValidate({
		email: user?.email || '',
		full_name: user?.full_name || '',
	}, zod4(updateProfileSchema));

	return {
		pageMetaTags,
		settings,
		user,
		form
	};
};
export const actions = {
	default: async ({ locals, request }) => {
		const { user } = locals;
		const form = await superValidate(request, zod4(updateProfileSchema));

		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', ')
			});
		}

		const result = await locals.deps.userService.updateProfile(form.data);
		if (result instanceof Error) {
			return fail(500, {
				form,
				message: result.message || 'Failed to update profile'
			});
		}

		return {
			form,
			success: true,
			message: 'Profile updated successfully'
		};
	}
};
