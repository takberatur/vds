import { superValidate } from 'sveltekit-superforms';
import { resetPasswordSchema } from '$lib/utils/schema';
import { defaultMetaTags } from '@/utils/meta-tags.js';
import { fail, redirect } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';
import { localizeHref } from '@/paraglide/runtime';
import type { SingleResponse } from "@siamf/google-translate";
import { capitalizeFirstLetter } from "@/utils/format.js";

export const load = async ({ locals, url, parent }) => {
	const { user, settings, deps, lang } = locals;

	const token = url.searchParams.get('token');
	if (!token) {
		throw redirect(302, localizeHref('/login'));
	}

	const defaultOrigin = await parent().then((data) => data.canonicalUrl || '');
	const alternates = await parent().then((data) => data.alternates || []);

	const title = await deps.languageHelper.singleTranslate('Reset Password', lang) as SingleResponse;
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

	const form = await superValidate(zod4(resetPasswordSchema));

	return {
		pageMetaTags,
		token,
		form,
		settings,
		user,
		lang
	};
};

export const actions = {
	default: async ({ request, locals }) => {
		const form = await superValidate(request, zod4(resetPasswordSchema));
		if (!form.valid) {
			return fail(400, {
				form,
				success: false,
				message: Object.values(form.errors).flat().join(', '),
				error: form.errors
			});
		}

		const response = await locals.deps.authService.resetPassword(form.data);

		if (response instanceof Error) {
			return fail(400, {
				form,
				success: false,
				message: response.message || 'Failed to reset password'
			});
		}

		throw redirect(303, localizeHref('/login'));
	}
};
