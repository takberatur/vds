import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { updateSettingRobotTxt } from '$lib/utils/schema';
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
			title: `Robots.txt Setting - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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

	const robotTxt = await GetRobotTxtContent();

	const form = await superValidate(
		{
			content: robotTxt
		},
		zod4(updateSettingRobotTxt)
	);

	return {
		pageMetaTags,
		settings,
		user,
		form
	};
};

export const actions = {
	default: async ({ request }) => {
		const form = await superValidate(request, zod4(updateSettingRobotTxt));
		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', '),
				error: form.errors
			});
		}


		const error = await UpdateRobotTxtContent(form.data.content || '');
		if (error instanceof Error) {
			return fail(500, {
				form,
				message: error.message || 'Failed to update robots.txt settings.'
			});
		}

		return {
			form,
			message: 'robots.txt updated successfully.'
		};
	}
};


async function GetRobotTxtContent(): Promise<string> {
	try {

		const staticPath = join(process.cwd(), 'static', 'robots.txt');
		let adsContent: string;


		if (existsSync(staticPath)) {
			adsContent = readFileSync(staticPath, 'utf-8');
		} else {
			adsContent = generateDefaultRobotTxt();

			writeFileSync(staticPath, adsContent, 'utf-8');
		}
		return adsContent;
	} catch (error) {
		console.error('Error reading robots.txt:', error);
		const fallbackContent = generateDefaultRobotTxt();
		return fallbackContent;
	}
}

async function UpdateRobotTxtContent(content: string): Promise<string | Error> {
	try {
		const staticPath = join(process.cwd(), 'static', 'robots.txt');
		writeFileSync(staticPath, content, 'utf-8');
		return 'success';
	} catch (error) {
		console.error('Error writing robots.txt:', error);
		return error instanceof Error ? error : new Error('Unknown error');
	}
}


function generateDefaultRobotTxt(): string {
	return `# Default robots.txt for ${process.env.ORIGIN || 'your-domain.com'}
User-agent: *
Disallow: /
# Add your own robot.txt entries here
# Contact: admin@your-domain.com`;
}
