
import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { updatePlatformSchema } from '$lib/utils/schema';
import { fail, redirect } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';
import { localizeHref } from '@/paraglide/runtime';

export const load = async ({ locals, url, params }) => {
	const { settings, deps } = locals;

	const id = params.id;
	if (!id) {
		throw redirect(302, localizeHref('/platform'));
	}

	const platform = await deps.platformService.GetPlatformByID(id);
	if (!platform) {
		throw redirect(302, localizeHref('/platform'));
	}

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Edit Platform ${platform.name} - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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
		id: platform.id,
		slug: platform.slug,
		type: platform.type,
		name: platform.name,
		url_pattern: platform.url_pattern || '',
		is_active: platform.is_active,
		is_premium: platform.is_premium,
		config: platform.config || {}
	}, zod4(updatePlatformSchema));

	return {
		pageMetaTags,
		platform: platform as Platform,
		settings,
		form
	};
};

export const actions = {
	default: async ({ request, locals, params }) => {
		const { deps } = locals;
		const id = params.id;
		if (!id) {
			throw redirect(302, localizeHref('/platform'));
		}

		const form = await superValidate(request, zod4(updatePlatformSchema));
		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', ')
			});
		}

		const response = await deps.platformService.UpdatePlatform(form.data);
		if (response instanceof Error) {
			return fail(400, {
				form,
				message: response.message || 'Failed to update platform'
			});
		}
		throw redirect(302, localizeHref('/platform'));
	}
}
