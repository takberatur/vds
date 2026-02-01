import { defaultMetaTags } from '@/utils/meta-tags.js';
import type { SingleResponse } from "@siamf/google-translate";
import { capitalizeFirstLetter } from "@/utils/format.js";
import type { Dependencies } from '$lib/server';

export const load = async ({ locals, url, parent }) => {
	const { user, settings, lang, deps } = locals;

	const defaultOrigin = await parent().then((data) => data.canonicalUrl || '');
	const alternates = await parent().then((data) => data.alternates || []);
	const posts = await getPosts(url, deps);
	const pageMeta = await translatePageMeta(lang, deps, settings);

	const pageMetaTags = defaultMetaTags({
		path_url: defaultOrigin,
		title: `${capitalizeFirstLetter(pageMeta.title)} - ${capitalizeFirstLetter(pageMeta.siteName)}`,
		tagline: capitalizeFirstLetter(pageMeta.tagline),
		description: capitalizeFirstLetter(pageMeta.description),
		keywords: pageMeta.keywords.map((keyword: string) => capitalizeFirstLetter(keyword)),
		robots: 'index, follow',
		canonical: defaultOrigin,
		alternates,
		graph_type: 'website',
		language: lang,
	}, settings);

	const postsTranslated = await translateBlogPosts(posts.data, lang, deps);

	return {
		user,
		settings,
		lang,
		posts: {
			data: postsTranslated,
			pagination: posts.pagination,
		},
		pageMetaTags,
	};
}

async function getPosts(url: URL, deps: Dependencies) {
	let queryParams = deps.queryHelper.parseQueryParams(url);

	const status = url.searchParams.get('status');
	if (status && status !== 'ALL') {
		queryParams.status = status;

	}

	const tag = url.searchParams.get('tag');
	if (tag && tag !== 'ALL') {
		queryParams.tag = tag;

	}

	const series = url.searchParams.get('series');
	if (series && series !== 'ALL') {
		queryParams.series = series;
	}

	const year = url.searchParams.get('year');
	if (year) {
		queryParams.year = parseInt(year);
	}

	const month = url.searchParams.get('month');
	if (month) {
		queryParams.month = parseInt(month);
	}

	const posts = await deps.postHelper.getAllPosts(queryParams);


	return posts;
}

async function translateBlogPosts(posts: BlogPost[], lang: string, deps: Dependencies) {
	return await Promise.all(posts.map(async (post) => {
		const title = await deps.languageHelper.singleTranslate(post.meta.title || '', lang) as SingleResponse;
		const description = await deps.languageHelper.singleTranslate(post.meta.description || '', lang) as SingleResponse;

		return {
			...post,
			meta: {
				...post.meta,
				title: title.data.target.text,
				description: description.data.target.text,
			},
		};
	}));
}

async function translatePageMeta(lang: string, deps: Dependencies, settings?: SettingsValue | null) {
	const siteName = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_name || '', lang) as SingleResponse;
	const title = await deps.languageHelper.singleTranslate('Blog', lang) as SingleResponse;
	const description = await deps.languageHelper.singleTranslate(settings?.WEBSITE?.site_description || '', lang) as SingleResponse;
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
