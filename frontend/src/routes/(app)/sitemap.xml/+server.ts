import { redirect } from '@sveltejs/kit';

let pages: string[] = [
	'about',
	'contact',
	'faq',
	'privacy',
	'terms',
];

const sitemap = (site: string, pages: string[]) => `<?xml version="1.0" encoding="UTF-8" ?>
<urlset
  xmlns="https://www.sitemaps.org/schemas/sitemap/0.9"
  xmlns:news="https://www.google.com/schemas/sitemap-news/0.9"
  xmlns:xhtml="https://www.w3.org/1999/xhtml"
  xmlns:mobile="https://www.google.com/schemas/sitemap-mobile/1.0"
  xmlns:image="https://www.google.com/schemas/sitemap-image/1.1"
  xmlns:video="https://www.google.com/schemas/sitemap-video/1.1"
>
  ${pages
		.map(
			(page) => `
  <url>
    <loc>${site}/${page}</loc>
    <changefreq>daily</changefreq>
    <priority>0.5</priority>
  </url>
  `
		)
		.join('')}
</urlset>`;

export async function GET({ url, request, locals }) {
	if (request.headers.get('accept')?.includes('application/json') ||
		url.pathname.includes('__data.json')) {
		throw redirect(307, '/sitemap.xml');
	}

	const origin = url.origin || request.headers.get('origin') || '';
	const site = origin.replace(/^https?:\/\//, '');

	const platforms = await locals.deps.platformService.GetAll() as Platform[] | Error;
	if (platforms instanceof Error) {
		const body = sitemap(origin, pages);
		const response = new Response(body);
		response.headers.set('Cache-Control', 'max-age=0, s-maxage=3600');
		response.headers.set('Content-Type', 'application/xml');
		return response;
	}

	platforms.forEach((platform) => {
		pages.push(platform.slug);
	});

	const body = sitemap(origin, pages);
	const response = new Response(body);
	response.headers.set('Cache-Control', 'max-age=0, s-maxage=3600');
	response.headers.set('Content-Type', 'application/xml');
	return response;
}
