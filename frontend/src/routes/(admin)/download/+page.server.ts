
import { defaultMetaTags } from '@/utils/meta-tags.js';

export const load = async ({ locals, url, parent }) => {
	const { settings, deps } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Download - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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

	const download = await deps.downloadService.GetDownloads(queryParams) as PaginatedResult<Download>;
	console.log(JSON.stringify(download, null, 2));

	return {
		pageMetaTags,
		download,
		settings
	};
};
