import { goto } from '$app/navigation';

export function localizePath(path: string, lang: string) {
	return `/${lang}${path.startsWith('/') ? path : `/${path}`}`;
}

export function gotoLocale(path: string, lang: string) {
	const href = `/${lang}${path.startsWith('/') ? path : `/${path}`}`;
	goto(href);
}
