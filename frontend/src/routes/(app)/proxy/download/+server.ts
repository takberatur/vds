import { json } from '@sveltejs/kit';

export async function POST({ request }) {
	try {
		const { videoUrl, filename } = await request.json();

		if (!videoUrl) {
			return json({ error: 'Video URL is required' }, {
				status: 400,
				headers: { 'Content-Type': 'application/json' }
			});
		}

		const rawUrl = String(videoUrl);
		const cleanedUrl = rawUrl.trim().replace(/^`+|`+$/g, '').replace(/`/g, '');

		let targetUrl: string;

		try {
			targetUrl = new URL(cleanedUrl).toString();
		} catch {
			return json({ error: 'Invalid video URL' }, {
				status: 400,
				headers: { 'Content-Type': 'application/json' }
			});
		}

		const urlObj = new URL(targetUrl);
		const isTikTok = urlObj.hostname.includes('tiktok.com');

		const fetchOptions: RequestInit = {};

		if (isTikTok) {
			// Try minimal headers - sometimes CDN blocks complex headers
			fetchOptions.headers = {
				'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36',
				// Remove Referer/Origin as they can trigger hotlink protection on some CDNs
			};
		}

		const response = await fetch(targetUrl, fetchOptions);

		if (!response.ok) {
			return json({
				error: `Failed to fetch video: ${response.status}`
			}, {
				status: response.status,
				headers: { 'Content-Type': 'application/json' }
			});
		}

		const buffer = await response.arrayBuffer();
		const contentType = response.headers.get('content-type') || 'video/mp4';

		const baseName = filename || 'video';
		const asciiName = baseName.replace(/[^a-zA-Z0-9\-\._ ]/g, '_');
		const encodedName = encodeURIComponent(baseName);

		const headers = new Headers();
		headers.set('Content-Type', contentType);
		headers.set('Content-Disposition', `attachment; filename="${asciiName}.mp4"; filename*=UTF-8''${encodedName}.mp4`);

		return new Response(buffer, { headers });

	} catch (error) {
		console.error('Download error:', error);
		return json({
			error: 'Download failed',
			details: error instanceof Error ? error.message : 'Unknown error'
		}, {
			status: 500,
			headers: { 'Content-Type': 'application/json' }
		});
	}
}
