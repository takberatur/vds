import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { updateSettingAdsTxt } from '$lib/utils/schema';
import { fail } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';
import { readFileSync, writeFileSync, existsSync } from 'fs';
import { join } from 'path';

export const load = async ({ locals, url, parent }) => {
	const { settings, user } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Ads.txt Setting - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
			tagline: settings?.WEBSITE.site_tagline || '',
			description: settings?.WEBSITE.site_description || '',
			keywords: settings?.WEBSITE.site_keywords?.split(', ') || [''],
			robots: 'noindex, nofollow',
			canonical: defaultOrigin,
			graph_type: 'website',
			use_tagline: false
		},
		settings
	);

	const adsTxt = await GetAdsTxtContent();

	const form = await superValidate(
		{
			content: adsTxt
		},
		zod4(updateSettingAdsTxt)
	);

	return {
		pageMetaTags,
		settings,
		user,
		form
	};
};

export const actions = {
	default: async ({ request, locals }) => {
		const form = await superValidate(request, zod4(updateSettingAdsTxt));
		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', '),
				error: form.errors
			});
		}


		const error = await UpdateAdsTxtContent(form.data.content || '');
		if (error instanceof Error) {
			return fail(500, {
				form,
				message: error.message || 'Failed to update ads.txt settings.'
			});
		}

		return {
			form,
			message: 'ads.txt updated successfully.'
		};
	}
};


async function GetAdsTxtContent(): Promise<string> {
	try {

		const staticPath = join(process.cwd(), 'static', 'ads.txt');
		let adsContent: string;


		if (existsSync(staticPath)) {
			adsContent = readFileSync(staticPath, 'utf-8');
		} else {
			adsContent = generateDefaultAdsTxt();

			writeFileSync(staticPath, adsContent, 'utf-8');
		}
		return adsContent;
	} catch (error) {
		console.error('Error reading ads.txt:', error);
		const fallbackContent = generateDefaultAdsTxt();
		return fallbackContent;
	}
}

async function UpdateAdsTxtContent(content: string): Promise<string | Error> {
	try {
		const staticPath = join(process.cwd(), 'static', 'ads.txt');
		writeFileSync(staticPath, content, 'utf-8');
		return 'success';
	} catch (error) {
		console.error('Error writing ads.txt:', error);
		return error instanceof Error ? error : new Error('Unknown error');
	}
}


function generateDefaultAdsTxt(): string {
	return `# Default ads.txt for ${process.env.ORIGIN || 'your-domain.com'}
google.com, pub-0000000000000000, DIRECT, f08c47fec0942fa0
google.com, pub-0000000000000001, RESELLER
# Add your own ad network entries here
# Contact: admin@your-domain.com`;
}
