
import { defaultMetaTags } from '@/utils/meta-tags.js';

export const load = async ({ locals, url, parent }) => {
	const { settings, deps, user } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Subscription - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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

	let queryParams = deps.queryHelper.parseQueryParams(url);

	const status = url.searchParams.get('status');
	if (status && status !== 'ALL') {
		queryParams.status = status
	}

	const subscription = await deps.subscriptionService.FindAll(queryParams) as PaginatedResult<Subscription>;

	return {
		pageMetaTags,
		settings,
		subscription,
		user
	};
};
