import { json, type RequestHandler } from '@sveltejs/kit';
import { ApiClientHandler } from '$lib/helpers/api_helpers';

export const POST: RequestHandler = async (event) => {
	let payload: Record<string, any> | null = null;
	try {
		payload = await event.request.json();
	} catch {
		return json({ ok: false }, { status: 400 });
	}
	if (!payload) {
		return json({ ok: false }, { status: 400 });
	}

	const api = new ApiClientHandler(event);
	const data = {
		error: String(payload.error ?? ''),
		message: String(payload.message ?? ''),
		platform_id: String(payload.platform_id ?? ''),
		user_id: String(payload.user_id ?? ''),
		ip_address: '',
		user_agent: String(payload.user_agent ?? ''),
		url: String(payload.url ?? ''),
		method: String(payload.method ?? ''),
		request: String(payload.request ?? ''),
		status: Number.isFinite(payload.status) ? Number(payload.status) : 0,
		level: String(payload.level ?? 'error'),
		locale: String(payload.locale ?? ''),
		timestamp_ms: Number.isFinite(payload.timestamp_ms) ? Number(payload.timestamp_ms) : Date.now()
	};

	const resp = await api.publicRequest('POST', '/web-client/report/errors', data, true);
	if (!resp.success) {
		return json({ ok: false }, { status: 502 });
	}
	return json({ ok: true });
};

