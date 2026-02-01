import { redirect } from '@sveltejs/kit';
import { localizeHref, locales as availableLocales } from '@/paraglide/runtime.js';
import { LanguageLabels } from '@/utils/localize-path.js';
import * as i18n from '@/paraglide/messages.js';

export async function GET({ url, request, locals }) {
	if (request.headers.get('accept')?.includes('application/json') ||
		url.pathname.includes('__data.json')) {
		throw redirect(307, localizeHref('/rss.xml', { locale: locals.lang }));
	}

	let origin = url.origin || request.headers.get('origin') || '';
	if (origin.startsWith('http://')) {
		origin = origin.replace('http://', 'https://');
	}

	const xml = `<?xml version="1.0" encoding="UTF-8"?>
    <rss version="2.0">
      <channel>
        <title>${i18n.rss_feed_list()}</title>
        <link>${origin}/rss.xml</link>
        <description>${i18n.rss_feed_list_description()}</description>
				${availableLocales.map(l => `<item>
          <title>RSS Feed - ${LanguageLabels[l] || l.toUpperCase()}</title>
          <link>${origin}/rss-${l}.xml</link>
        </item>`).join('')}
      </channel>
    </rss>`.trim();

	return new Response(xml, { headers: { 'Content-Type': 'application/xml' } });
}
