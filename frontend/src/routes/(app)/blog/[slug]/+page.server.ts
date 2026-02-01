import { defaultMetaTags } from '@/utils/meta-tags.js';
import { redirect } from '@sveltejs/kit';
import { localizeHref } from '@/paraglide/runtime.js';
import type { SingleResponse } from "@siamf/google-translate";
import { capitalizeFirstLetter } from "@/utils/format.js";
import type { Dependencies } from '$lib/server';

export const load = async ({ locals, url, parent, params }) => {
	const { user, settings, lang, deps } = locals;

	const slugParams = params.slug || '';
	if (!slugParams) {
		throw redirect(302, localizeHref('/blog'));
	}

	const posts = await deps.postHelper.getPostBySlug(slugParams) as BlogPost | null;
	if (!posts) {
		throw redirect(302, localizeHref('/blog'));
	}

	const defaultOrigin = await parent().then((data) => data.canonicalUrl || '');
	const alternates = await parent().then((data) => data.alternates || []);

	const postsTranslated = await translatePageMeta(lang, deps, posts, settings);

	const pageMetaTags = defaultMetaTags({
		path_url: defaultOrigin,
		title: `${capitalizeFirstLetter(postsTranslated.title)} - ${capitalizeFirstLetter(postsTranslated.siteName)}`,
		tagline: capitalizeFirstLetter(postsTranslated.tagline),
		description: capitalizeFirstLetter(postsTranslated.description),
		keywords: postsTranslated.keywords.map((keyword: string) => capitalizeFirstLetter(keyword)),
		robots: 'index, follow',
		canonical: defaultOrigin,
		alternates,
		graph_type: 'website',
		language: lang,
	}, settings);

	return {
		user,
		settings,
		lang,
		posts,
		pageMetaTags,
	};
}

async function translatePageMeta(lang: string, deps: Dependencies, data: BlogPost, settings?: SettingsValue | null) {
	const siteName = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_name || '', lang) as SingleResponse;
	const title = await deps.languageHelper.singleTranslate(data.meta.title || '', lang) as SingleResponse;
	const description = await deps.languageHelper.singleTranslate(data.meta.description || '', lang) as SingleResponse;
	const tagline = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_tagline || '', lang) as SingleResponse;
	const keywords = await Promise.all((settings?.WEBSITE?.site_keywords?.split(',') || ['']).map(async (keyword) => await deps.languageHelper.singleTranslate(keyword.trim(), lang) as SingleResponse));

	return {
		siteName: siteName.data.target.text,
		title: title.data.target.text,
		tagline: tagline.data.target.text,
		description: description.data.target.text,
		keywords: keywords.map((keyword: SingleResponse) => keyword.data.target.text),
	};
}
