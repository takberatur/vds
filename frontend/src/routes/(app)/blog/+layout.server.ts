import { locales as SUPPORTED_LOCALES } from '@/paraglide/runtime';
import { env } from '$env/dynamic/private';

export const prerender = true

export const config = {
	isr: {
		expiration: 60
	}
};

export const load = async ({ locals, url }) => {
	const { user, settings, lang, deps } = locals;

	const defaultOrigin = url.origin;
	let canonicalUrl = defaultOrigin;
	if (env.NODE_ENV === 'production' && canonicalUrl.startsWith('http://')) {
		canonicalUrl = canonicalUrl.replace('http://', 'https://');
	}

	const alternates = SUPPORTED_LOCALES.map((lang) => ({
		lang,
		href: `${canonicalUrl}/${lang}`
	}));
	const normalizedAlternates = alternates.map(alt => ({
		...alt,
		href: normalizeUrl(alt.href)
	}));

	try {
		const platforms = await deps.platformService.GetAll();
		if (platforms instanceof Error) {
			throw platforms
		}
		return {
			user,
			settings,
			lang,
			canonicalUrl,
			alternates: normalizedAlternates,
			platforms,
		};
	} catch (error) {
		console.error('Error fetching platforms:', error);
		return {
			user,
			settings,
			lang,
			canonicalUrl,
			alternates: normalizedAlternates,
			platforms: [],
		};

	}


}

function normalizeUrl(urlString: string): string {
	try {
		const url = new URL(urlString);
		if (env.NODE_ENV === 'production' && url.protocol === 'http:') {
			url.protocol = 'https:';
		}
		return url.href;
	} catch {
		return urlString;
	}
}
