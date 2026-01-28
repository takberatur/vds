import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { downloadVideoSchema } from '$lib/utils/schema';
import { fail } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';
import { redirect } from '@sveltejs/kit';
import { localizeHref } from '@/paraglide/runtime.js';
import type { SingleResponse } from "@siamf/google-translate";
import { capitalizeFirstLetter } from "@/utils/format.js";

export async function load({ locals, url, params, parent }) {
	const { user, settings, deps, lang } = locals;

	const { slug } = params;
	if (!slug) {
		throw redirect(302, localizeHref('/'))
	}

	const platform = await deps.platformService.PublicGetPlatformBySlug(slug) as Platform | Error;
	if (platform instanceof Error) {
		throw redirect(302, localizeHref('/'))
	}

	const defaultOrigin = await parent().then((data) => data.canonicalUrl || '');
	const alternates = await parent().then((data) => data.alternates || []);

	const title = await deps.languageHelper.singleTranslate(`Download ${platform.name} Video Without Watermark Full HD Free`, lang) as SingleResponse;
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

	const form = await superValidate({
		url: '',
		type: platform.type,
		user_id: user?.id || '',
		platform_id: platform.id,
		app_id: undefined,
	}, zod4(downloadVideoSchema));

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
			platform,
			lang,
			form,
		};
	} catch (error) {
		console.error('Failed to get platforms:', error);
		return {
			pageMetaTags,
			user,
			settings,
			platforms: [],
			platform: null,
			lang,
			form,
		};
	}
}
export const actions = {
	default: async ({ request, locals }) => {
		const { deps } = locals;
		const form = await superValidate(request, zod4(downloadVideoSchema));
		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', '),
				data: null,
			});
		}
		try {
			const response = await deps.webService.DownloadVideo(form.data) as ApiResponse<Download>;

			if (!response.success) {
				return fail(500, {
					form,
					message: response.message,
					data: null,
				});
			}

			return {
				form,
				message: response.message,
				data: response.data,
			};
		} catch (error: any) {
			console.error('❌ Download action failed unexpectedly', error);
			return fail(500, {
				form,
				message: error?.message || 'Failed to process download request',
				data: null,
			});
		}
	}
}
