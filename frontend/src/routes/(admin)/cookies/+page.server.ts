import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { updateSettingCookie } from '$lib/utils/schema';
import { fail } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';

export const load = async ({ locals, url, parent }) => {
	const { settings, user } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Cookies Setting - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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

	const cookies = await locals.deps.adminService.getCookies() as CookieItem;

	const form = await superValidate(
		{
			cookies: Array.isArray(cookies?.lines) ? cookies.lines.join('\n') : (cookies?.lines || '')
		},
		zod4(updateSettingCookie)
	);

	return {
		pageMetaTags,
		settings,
		user,
		form,
		cookies
	};
};

export const actions = {
	default: async ({ request, locals }) => {
		const { deps } = locals;
		const form = await superValidate(request, zod4(updateSettingCookie));

		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', ')
			});
		}

		const cookies = await deps.adminService.updateCookies(form.data.cookies?.split('\n') || []);

		if (!cookies) {
			return fail(500, {
				form,
				message: 'Failed to update cookies'
			});
		}

		return {
			form,
			message: 'Cookies updated successfully'
		}
	}
}
