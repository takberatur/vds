import { defaultMetaTags } from '@/utils/meta-tags.js';
import * as i18n from '@/paraglide/messages.js';
import type { SingleResponse } from "@siamf/google-translate";
import { capitalizeFirstLetter } from "@/utils/format.js";

export const load = async ({ locals, url, parent }) => {
	const { user, settings, deps, lang } = locals;

	const defaultOrigin = await parent().then((data) => data.canonicalUrl || '');
	const alternates = await parent().then((data) => data.alternates || []);
	const title = await deps.languageHelper.singleTranslate(i18n.terms_of_service(), lang) as SingleResponse;
	const siteName = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_name || '', lang) as SingleResponse;
	const tagline = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_tagline || '', lang) as SingleResponse;
	const description = await deps.languageHelper.singleTranslate(i18n.terms_of_service_description({ site_name: settings?.WEBSITE.site_name || 'Video Downloader' }), lang) as SingleResponse;
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
			alternates,
			graph_type: 'website',
			use_tagline: false,
			language: lang,
		},
		settings
	);

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
			lang
		};
	} catch (error) {
		console.error('Failed to get platforms:', error);
		return {
			pageMetaTags,
			user,
			settings,
			platforms: [],
			lang
		};
	}

};
