
import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { registerAppSchema } from '$lib/utils/schema';
import { fail, redirect } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';
import { localizeHref } from '@/paraglide/runtime';

export const load = async ({ locals, url }) => {
	const { settings } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Create Application - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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

	const form = await superValidate(zod4(registerAppSchema))

	return {
		pageMetaTags,
		form,
		settings
	};
};

export const actions = {
	default: async ({ request, locals }) => {
		const { deps } = locals;
		const form = await superValidate(request, zod4(registerAppSchema));

		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', ')
			});
		}

		const application = await deps.applicationService.create(form.data);

		if (application instanceof Error) {
			return fail(500, {
				form,
				message: application.message || 'Failed to create application'
			});
		}

		throw redirect(303, localizeHref(`/application`));
	}
}
