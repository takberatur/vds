import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { contactSchema } from '$lib/utils/schema';
import { fail } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';
import * as i18n from '@/paraglide/messages.js';
import type { SingleResponse } from "@siamf/google-translate";
import { capitalizeFirstLetter } from "@/utils/format.js";

export const load = async ({ locals, url, parent }) => {
	const { user, settings, deps, lang } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const title = await deps.languageHelper.singleTranslate(i18n.contact_us(), lang) as SingleResponse;
	const siteName = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_name || '', lang) as SingleResponse;
	const tagline = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_tagline || '', lang) as SingleResponse;
	const description = await deps.languageHelper.singleTranslate(i18n.contact_us_description({ site_name: settings?.WEBSITE.site_name || 'Video Downloader' }), lang) as SingleResponse;
	const keywords = await Promise.all((settings?.WEBSITE?.site_keywords?.split(',') || ['']).map(async (keyword) => await deps.languageHelper.singleTranslate(keyword.trim(), lang) as SingleResponse));


	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `${capitalizeFirstLetter(title.data.target.text || '')} - ${capitalizeFirstLetter(siteName.data.target.text || '')}`,
			tagline: capitalizeFirstLetter(tagline.data.target.text || ''),
			description: capitalizeFirstLetter(description.data.target.text || ''),
			keywords: keywords.map((keyword: SingleResponse) => capitalizeFirstLetter(keyword.data.target.text || '')),
			robots: 'index, follow',
			canonical: defaultOrigin,
			graph_type: 'website',
			use_tagline: false
		},
		settings
	);

	const form = await superValidate(zod4(contactSchema));

	try {
		const platforms = await deps.platformService.GetAll();
		if (platforms instanceof Error) {
			throw platforms
		}

		return {
			pageMetaTags,
			user,
			settings,
			platforms,
			form,
			lang
		};
	} catch (error) {
		console.error('Failed to get platforms:', error);
		return {
			pageMetaTags,
			user,
			settings,
			platforms: [],
			form,
			lang
		};
	}

};
export const actions = {
	default: async ({ locals, request }) => {
		const form = await superValidate(request, zod4(contactSchema));

		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', ')
			});
		}

		const result = await locals.deps.webService.Contact(form.data);
		if (result instanceof Error) {
			return fail(500, {
				form,
				message: result.message || i18n.contact_error()
			});
		}

		return {
			form,
			message: i18n.contact_success()
		}
	}
};
