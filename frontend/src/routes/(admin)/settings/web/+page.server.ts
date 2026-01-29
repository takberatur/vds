
import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { updateSettingWeb } from '$lib/utils/schema';
import { fail } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';

export const load = async ({ locals, url }) => {
	const { settings, user } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Web Setting - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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

	const form = await superValidate({
		site_name: settings?.WEBSITE.site_name || '',
		site_tagline: settings?.WEBSITE.site_tagline || '',
		site_description: settings?.WEBSITE.site_description || '',
		site_keywords: settings?.WEBSITE.site_keywords || '',
		site_email: settings?.WEBSITE.site_email || '',
		site_phone: settings?.WEBSITE.site_phone || '',
		site_url: settings?.WEBSITE.site_url || ''
	}, zod4(updateSettingWeb));

	return {
		pageMetaTags,
		settings,
		user,
		form
	};
};

export const actions = {
	default: async ({ request, locals }) => {
		const form = await superValidate(request, zod4(updateSettingWeb));
		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', '),
			});
		}

		// console.log(form.data);

		const settingsToUpdate = [
			{ key: 'site_name', value: form.data.site_name, group_name: 'WEBSITE' },
			{ key: 'site_tagline', value: form.data.site_tagline, group_name: 'WEBSITE' },
			{ key: 'site_description', value: form.data.site_description, group_name: 'WEBSITE' },
			{ key: 'site_keywords', value: form.data.site_keywords, group_name: 'WEBSITE' },
			{ key: 'site_email', value: form.data.site_email, group_name: 'WEBSITE' },
			{ key: 'site_phone', value: form.data.site_phone, group_name: 'WEBSITE' },
			{ key: 'site_url', value: form.data.site_url, group_name: 'WEBSITE' }
		].filter(s => s.value !== undefined) as { key: string; value: string; group_name: string }[];

		const updateResponse = await locals.deps.settingService.updateBulkSetting(settingsToUpdate);
		if (updateResponse instanceof Error) {
			return fail(500, {
				form,
				message: updateResponse.message || 'Failed to update settings.'
			});
		}

		return {
			form,
			message: 'Settings updated successfully.'
		};
	}
};
