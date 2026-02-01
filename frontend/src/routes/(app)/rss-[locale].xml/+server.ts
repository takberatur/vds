
import { type Locale } from '@/paraglide/runtime.js';

export async function GET({ url, request, locals, params }) {
	const { deps, settings } = locals

	const locale = params.locale as Locale;
	let origin = url.origin || request.headers.get('origin') || '';
	if (origin.startsWith('http://')) {
		origin = origin.replace('http://', 'https://');
	}

	const posts: PaginatedResult<BlogPost> = await deps.postHelper.getAllPosts({
		page: 1,
		limit: 1000,
	})

	const body = `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
	<title>${settings?.WEBSITE.site_name}</title>
	<link>${origin}/${locale}</link>
	<description>${settings?.WEBSITE.site_description}</description>

	${posts
			.data.map(
				(post) => `
	<item>
		<title><![CDATA[${post.meta.title}]]></title>
        <link>${origin}/${locale}/blog/${post.meta.slug}</link>
        <guid isPermaLink="true">${origin}/${locale}/blog/${post.meta.slug}</guid>
        <pubDate>${new Date(post.meta.publishedDate).toUTCString()}</pubDate>
        <description><![CDATA[${post.meta.description}]]></description>
	</item>
	`
			)
			.join('')}
</channel>
</rss>`;

	return new Response(body, {
		headers: {
			'Content-Type': 'application/xml',
			'Cache-Control': 'max-age=0, s-maxage=3600'
		}
	});
}
