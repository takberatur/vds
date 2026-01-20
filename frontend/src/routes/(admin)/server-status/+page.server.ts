import { defaultMetaTags } from '@/utils/meta-tags.js';

export const load = async ({ locals, url, parent }) => {
	const { settings, deps } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Server Status - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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

	const page = Number(url.searchParams.get('page')) || 1;
	const limit = Number(url.searchParams.get('limit')) || 50;

	const serverHealth = await deps.serverStatusService.GetServerHealth();
	const serverLogs = await deps.serverStatusService.GetServerLogs(page, limit);

	return {
		pageMetaTags,
		settings,
		serverHealth: serverHealth as ServerHealthResponse || null,
		serverLogs: serverLogs as PaginatedResult<ServerLogsResponse> | null
	};
};
