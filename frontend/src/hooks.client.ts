import type { HandleClientError } from '@sveltejs/kit';

let lastSentAt = 0;
let lastKey = '';

function safeString(v: unknown, max = 2000) {
	try {
		const s = typeof v === 'string' ? v : JSON.stringify(v);
		return s.length > max ? s.slice(0, max) + '...' : s;
	} catch {
		const s = String(v);
		return s.length > max ? s.slice(0, max) + '...' : s;
	}
}

async function report(payload: Record<string, any>) {
	try {
		const now = Date.now();
		const key = `${payload.level}|${payload.url}|${payload.message}|${payload.error}`.slice(0, 500);
		if (key === lastKey && now - lastSentAt < 15_000) return;
		if (now - lastSentAt < 3_000) return;
		lastKey = key;
		lastSentAt = now;

		const body = JSON.stringify(payload);
		if (navigator.sendBeacon) {
			const ok = navigator.sendBeacon('/api/report/errors', new Blob([body], { type: 'application/json' }));
			if (ok) return;
		}
		await fetch('/api/report/errors', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body
		});
	} catch {
		return;
	}
}

export const handleError: HandleClientError = async ({ error, event, status, message }) => {
	const url = event.url?.toString?.() ? event.url.toString() : '';
	const stack = (error instanceof Error ? error.stack : undefined) ?? safeString(error, 8000);
	const msg = (error instanceof Error ? error.message : undefined) ?? message ?? safeString(error, 1000);
	const locale = document?.documentElement?.lang ?? '';

	await report({
		level: 'error',
		message: msg,
		error: stack,
		url,
		status: status ?? 0,
		method: 'CLIENT',
		locale,
		user_agent: navigator.userAgent,
		timestamp_ms: Date.now()
	});
};

window.addEventListener('unhandledrejection', (ev) => {
	const reason = (ev as PromiseRejectionEvent).reason;
	const stack = reason instanceof Error ? reason.stack : safeString(reason, 8000);
	const msg = reason instanceof Error ? reason.message : safeString(reason, 1000);
	void report({
		level: 'error',
		message: msg,
		error: stack,
		url: location.href,
		status: 0,
		method: 'CLIENT',
		locale: document?.documentElement?.lang ?? '',
		user_agent: navigator.userAgent,
		timestamp_ms: Date.now()
	});
});
