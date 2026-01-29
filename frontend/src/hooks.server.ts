import { redirect, error as svelteError, type RequestEvent, type Handle } from '@sveltejs/kit';
import { building } from '$app/environment';
import { sequence } from '@sveltejs/kit/hooks';
import { Dependencies } from '$lib/server';
import { env } from '$env/dynamic/private';
import { paraglideMiddleware } from '$lib/paraglide/server';
import { localizeHref, locales as SUPPORTED_LOCALES, type Locale } from '@/paraglide/runtime';

const NODE_ENV = env.NODE_ENV || 'development';

const paraglideHandleBasic: Handle = ({ event, resolve }) => {
	return paraglideMiddleware(event.request, ({ request: localizedRequest, locale }) => {
		event.request = localizedRequest;
		event.locals.lang = detectLocale(event);

		return resolve(event, {
			transformPageChunk: ({ html }) => {
				return html.replace('%lang%', event.locals.lang)
			}
		});
	});
};

const paraglideHandleWithAutoDetectedLocale: Handle = ({ event, resolve }) => {
	const { url, request } = event;
	const pathname = event.url.pathname;

	if (
		pathname.startsWith('/api') ||
		pathname.startsWith('/_app') ||
		pathname.includes('.')
	) {
		return resolve(event);
	}

	const ua = request.headers.get('user-agent');
	const isBot = !!ua && /bot|crawl|spider|facebookexternalhit|twitterbot/i.test(ua);
	const pathLocale = getLocaleFromPath(pathname);

	if (isBot) {
		event.locals.lang = pathLocale ?? 'en'; { }

		return paraglideMiddleware(event.request, ({ request: localizedRequest, locale }) => {
			event.request = localizedRequest;
			return resolve(event, {
				transformPageChunk: ({ html, done }) => {
					return html.replace('%lang%', event.locals.lang)
				}
			});
		});
	}

	if (!pathLocale) {
		event.locals.lang = detectLocale(event);

		setCookie(event, event.locals.lang as Locale);
		throw redirect(302, `/${event.locals.lang}`);
	}

	if (pathLocale) {
		event.locals.lang = pathLocale;

		setCookie(event, pathLocale);

		return paraglideMiddleware(event.request, ({ request: localizedRequest, locale }) => {
			event.request = localizedRequest;

			return resolve(event, {
				// transformPageChunk: ({ html, done }) => {
				// 	if (done) {
				// 		return html
				// 			.replace(/%lang%/g, event.locals.lang)
				// 			.replace(/%canonical%/g, canonicalUrl)
				// 			.replace(/%hreflang%/g, event.request.headers.get('accept-language')?.split(',')[0] || 'en-US')
				// 			.replace(/<link rel="alternate"[^>]*%[^%]*%[^>]*>/g, '')
				// 			.replace(/<link rel="canonical"[^>]*%canonical%[^>]*>/g,
				// 				`<link rel="canonical" href="${canonicalUrl}" />`);
				// 	}
				// 	return html;
				// }
				transformPageChunk: ({ html, done }) => {
					return html.replace('%lang%', event.locals.lang)
				}
			});
		});
	}

	if (pathname === '/' || pathname === '') {
		const detected = detectLocale(event);

		setCookie(event, detected);

		throw redirect(302, `/${detected}`);
	}

	return resolve(event);

};

const paraglideHandleWithCloudflareWorker: Handle = async ({ event, resolve }) => {
	const { pathname } = event.url;

	const imagesRegex = /\.(png|jpg|jpeg|gif|webp|svg)$/;

	if (
		pathname.startsWith('/api') ||
		pathname.startsWith('/_app') ||
		pathname.startsWith('/favicon.ico') ||
		imagesRegex.test(pathname) ||
		pathname.startsWith('/images/') ||
		pathname.startsWith('/robots.txt') ||
		pathname.includes('.')
	) {
		return resolve(event);
	}

	const ua = event.request.headers.get('user-agent') ?? '';
	const isBot = /bot|crawl|spider|facebookexternalhit|twitterbot/i.test(ua);

	const pathLocale = getLocaleFromPath(pathname);

	if (isBot) {
		event.locals.lang = pathLocale ?? 'en';
		setCookie(event, event.locals.lang as Locale);

		return paraglideMiddleware(event.request, () =>
			resolve(event)
		);
	}

	if (!pathLocale) {
		event.locals.lang = detectLocale(event);

		setCookie(event, event.locals.lang as Locale);
		throw redirect(302, `/${event.locals.lang}${pathname === '/' ? '' : pathname}`);
	}

	event.locals.lang = pathLocale;
	setCookie(event, pathLocale);

	return paraglideMiddleware(event.request, ({ locale }) =>
		resolve(event, {
			transformPageChunk: ({ html, done }) =>
				done ? html.replace('%lang%', locale) : html
		})
	);
};

const dependenciesInject: Handle = async ({ event, resolve }) => {
	event.locals.deps = new Dependencies(event);
	event.locals.session = {
		set: (key, value, maxAge?: number) => {
			const isProduction = NODE_ENV === 'production';

			event.cookies.set(key, value, {
				httpOnly: true,
				secure: isProduction,
				sameSite: 'lax',
				path: '/',
				maxAge: maxAge ?? 60 * 60 * 24 * 7 // default 1 week
			});
		},
		get: (key) => {
			const val = event.cookies.get(key);
			return val ?? null;
		},
		delete: (key) => {
			const isProduction = NODE_ENV === 'production';
			event.cookies.delete(key, {
				path: '/',
				httpOnly: true,
				secure: isProduction,
				sameSite: 'lax'
			});
		}
	};
	event.locals.safeGetUser = async () => {
		try {
			const accessToken = event.locals.session?.get('access_token');

			if (!accessToken) {
				return null;
			}

			const user = await event.locals.deps.userService.getCurrentUser();
			if (user instanceof Error) {
				return null;
			}
			return user;
		} catch (e) {
			console.error("Error fetching user:", e);
			return null;
		}
	};
	event.locals.safeGetSettings = async () => {
		try {
			const settings = await event.locals.deps.settingService.getPublicSettings();
			if (settings instanceof Error) {
				return null;
			}
			return settings;
		} catch (e) {
			console.error("Error fetching settings:", e);
			return null;
		}
	};

	const response = await resolve(event);
	return response;
};

const authHandle: Handle = async ({ event, resolve }) => {
	const { locals } = event;

	try {
		if (!locals.settings) {
			locals.settings = await event.locals.safeGetSettings();
		}
	} catch (error: any) {
		console.error('Error fetching settings:', error);
	}

	try {
		if (!locals.user) {
			locals.user = await event.locals.safeGetUser();
		}

		return resolve(event);
	} catch (error: any) {
		console.error('Auth Handle Error:', error);
		if (error?.status === 302 || error?.status === 301 || error?.status === 303) {
			throw error;
		}
		// If it's a critical error, we might want to throw, but for auth check failure we usually just proceed as guest
		return resolve(event);
	}
}

const adminMiddleware: Handle = async ({ event, resolve }) => {
	const { url, locals } = event;

	const protectedPaths = [
		'/dashboard',
		'/download',
		'/settings',
		'/accounts',
		'/application',
		'/cookies',
		'/users',
		"/platform",
		"/subscription",
		"/transaction"
	];

	const authPath = [
		'/login',
		'/forgot-password',
		'/reset-password',
	]

	const isProtected = protectedPaths.some(path => url.pathname.startsWith(path));

	if (isProtected) {
		if (!locals.user) {
			throw redirect(303, localizeHref(`/login?redirect=${encodeURIComponent(url.pathname)}`, { locale: event.locals.lang }));
		}

		const user = locals.user;
		const isAdmin = user.role?.name === 'admin' || user.role?.name === 'Admin';

		if (!isAdmin) {
			throw redirect(303, localizeHref('/user', { locale: event.locals.lang }));
		}
	}

	const isAuthRoute = authPath.some(path => url.pathname.startsWith(path));

	if (isAuthRoute) {
		if (locals.user && locals.user.role?.name === 'admin') {
			throw redirect(303, localizeHref('/dashboard', { locale: event.locals.lang }));
		} else if (locals.user && locals.user.role?.name === 'user') {
			throw redirect(303, localizeHref('/user', { locale: event.locals.lang }));
		}
	}

	return resolve(event);
};

const errorHandling: Handle = async ({ event, resolve }) => {
	try {
		const response = await resolve(event);

		if (response.status === 404) {
			const locale = getLocaleFromPath(event.url.pathname);

			let redirectPath = '/';
			if (locale) {
				redirectPath = `/${locale}`;
			}


			const isStaticAsset = /\.(css|js|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$/i.test(event.url.pathname);
			const isApiRoute = event.url.pathname.startsWith('/api/');
			const isImagesRoute = event.url.pathname.startsWith('/images/');
			const isInternalRoute = event.url.pathname.startsWith('/_') || event.url.pathname.includes('__');

			if (!isStaticAsset && !isApiRoute && !isInternalRoute && !isImagesRoute) {
				return new Response(null, {
					status: 302,
					headers: {
						'Location': redirectPath,
						'Cache-Control': 'no-cache'
					}
				});
			}
		}

		const authRoute = [
			'/login',
			'/forgot-password',
			'/reset-password',
		]


		if (authRoute.includes(event.url.pathname)) {
			response.headers.set('Cross-Origin-Opener-Policy', 'unsafe-none');
		}
		return response;
	} catch (error) {
		console.error('Server error:', error);

		if (
			(error as any)?.status === 301 ||
			(error as any)?.status === 302 ||
			(error as any)?.status === 303 ||
			(error as any)?.status === 307 ||
			(error as any)?.status === 308
		) {
			throw error;
		}

		if (event.url.pathname.startsWith('/api/')) {
			return new Response(
				JSON.stringify({
					error: "Internal Server Error",
					message: error instanceof Error ? error.message : "An unknown error occurred"
				}),
				{
					status: 500,
					headers: { 'Content-Type': 'application/json' }
				}
			);
		}

		throw error;
	}
};

export const handle: Handle = sequence(
	paraglideHandleWithAutoDetectedLocale,
	dependenciesInject,
	authHandle,
	adminMiddleware,
	errorHandling
);

function hasLocalePrefix(path: string): boolean {
	return SUPPORTED_LOCALES.some(
		(l) => path === `/${l}` || path.startsWith(`/${l}/`)
	);
}

function getLocaleFromPath(pathname: string): Locale | null {
	const match = pathname.match(/^\/(en|id|es|ru|pt|fr|de|zh|hi|ar|ja|tr|vi|th|el|it)(\/|$)/);
	return match ? (match[1] as Locale) : null;
}

function detectLocale(event: RequestEvent): Locale {
	// const cookie = event.cookies.get('PARAGLIDE_LOCALE') as Locale | null;
	// if (cookie && SUPPORTED_LOCALES.includes(cookie)) return cookie;

	const accept = event.request.headers.get('accept-language');
	const l = accept?.split(',')[0].split('-')[0] as Locale;
	const supported = SUPPORTED_LOCALES.includes(l);

	return supported ? l : 'en';
}

function setCookie(event: RequestEvent, locale: Locale) {
	event.cookies.set('PARAGLIDE_LOCALE', locale, {
		httpOnly: true,
		secure: NODE_ENV === 'production',
		sameSite: 'lax',
		path: '/',
		maxAge: 60 * 60 * 24 * 7
	});
}
