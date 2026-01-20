import { redirect, error as svelteError, type RequestEvent, type Handle } from '@sveltejs/kit';
import { sequence } from '@sveltejs/kit/hooks';
import { Dependencies } from '$lib/server';
import { env } from '$env/dynamic/private';
import { paraglideMiddleware } from '$lib/paraglide/server';
import { localizeHref } from '@/paraglide/runtime';

const NODE_ENV = env.NODE_ENV || 'development';

const paraglideHandle: Handle = ({ event, resolve }) =>
	paraglideMiddleware(event.request, ({ request: localizedRequest, locale }) => {
		event.request = localizedRequest;
		event.locals.lang = locale;
		return resolve(event, {
			transformPageChunk: ({ html }) => {
				return html.replace('%lang%', locale);
			}
		});
	});
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

export const authHandle: Handle = async ({ event, resolve }) => {
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

export const adminMiddleware: Handle = async ({ event, resolve }) => {
	const { url, locals } = event;

	const protectedPaths = [
		'/dashboard',
		'/download',
		'/settings',
		'/accounts',
		'/application',
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
			throw redirect(303, localizeHref(`/login?redirect=${encodeURIComponent(url.pathname)}`));
		}

		const user = locals.user;
		const isAdmin = user.role?.name === 'admin' || user.role?.name === 'Admin';

		if (!isAdmin) {
			throw redirect(303, localizeHref('/user'));
		}
	}

	const isAuthRoute = authPath.some(path => url.pathname.startsWith(path));

	if (isAuthRoute) {
		if (locals.user && locals.user.role?.name === 'admin') {
			throw redirect(303, localizeHref('/dashboard'));
		} else if (locals.user && locals.user.role?.name === 'user') {
			throw redirect(303, localizeHref('/user'));
		}
	}

	return resolve(event);
};

const errorHandling: Handle = async ({ event, resolve }) => {
	try {
		const response = await resolve(event);

		if (response.status === 404) {
			// console.log('404 Not Found:', event.url.pathname);
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

		// Don't redirect to root on server error, let SvelteKit handle the error page
		throw error;
	}
};

export const handle: Handle = sequence(
	paraglideHandle,
	dependenciesInject,
	authHandle,
	adminMiddleware,
	errorHandling
);
