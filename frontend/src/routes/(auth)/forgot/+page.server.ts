import { superValidate } from 'sveltekit-superforms';
import { forgotSchema } from '$lib/utils/schema';
import { defaultMetaTags } from '@/utils/meta-tags.js';
import { fail } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';
import type { SingleResponse } from "@siamf/google-translate";
import { capitalizeFirstLetter } from "@/utils/format.js";

export const load = async ({ locals, url, parent }) => {
	const { user, settings, deps, lang } = locals;

	const defaultOrigin = await parent().then((data) => data.canonicalUrl || '');
	const alternates = await parent().then((data) => data.alternates || []);

	const title = await deps.languageHelper.singleTranslate('Forgot Password', lang) as SingleResponse;
	const siteName = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_name || '', lang) as SingleResponse;
	const tagline = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_tagline || '', lang) as SingleResponse;
	const description = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_description || '', lang) as SingleResponse;
	const keywords = await Promise.all((settings?.WEBSITE?.site_keywords?.split(',') || ['']).map(async (keyword) => await deps.languageHelper.singleTranslate(keyword.trim(), lang) as SingleResponse));

	const pageMetaTags = defaultMetaTags({
		path_url: defaultOrigin,
		title: `${capitalizeFirstLetter(title.data.target.text || '')} - ${capitalizeFirstLetter(siteName.data.target.text || '')}`,
		tagline: capitalizeFirstLetter(tagline.data.target.text || ''),
		description: capitalizeFirstLetter(description.data.target.text || ''),
		keywords: keywords.map((keyword: SingleResponse) => capitalizeFirstLetter(keyword.data.target.text || '')),
		robots: 'index, follow',
		canonical: defaultOrigin,
		alternates,
		graph_type: 'website'
	});

	const form = await superValidate(zod4(forgotSchema));

	return {
		pageMetaTags,
		form,
		settings,
		user,
		lang
	};
};

export const actions = {
	default: async ({ request, locals }) => {
		const form = await superValidate(request, zod4(forgotSchema));
		if (!form.valid) {
			return fail(400, {
				form,
				success: false,
				message: Object.values(form.errors).flat().join(', '),
				error: form.errors
			});
		}

		const response = await locals.deps.authService.forgotPassword(form.data.email);

		if (response instanceof Error) {
			return fail(400, {
				form,
				success: false,
				message: response.message || 'Failed to send reset email'
			});
		}
		return {
			form,
			success: true,
			message: response || 'Reset email sent successfully',
		};
	}
};
