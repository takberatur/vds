
import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { updateSettingSystem } from '$lib/utils/schema';
import { fail } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';

export const load = async ({ locals, url, parent }) => {
	const { settings, user } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `System Setting - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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

	const form = await superValidate(
		{
			enable_documentation: settings?.SYSTEM.enable_documentation ?? true,
			maintenance_mode: settings?.SYSTEM.maintenance_mode ?? false,
			maintenance_message: settings?.SYSTEM.maintenance_message ?? '',
			source_logo_favicon: settings?.SYSTEM.source_logo_favicon ?? 'local',
			histats_tracking_code: settings?.SYSTEM.histats_tracking_code ?? ''
		},
		zod4(updateSettingSystem)
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
		const form = await superValidate(request, zod4(updateSettingSystem));
		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', '),
				error: form.errors
			});
		}

		const settingsToUpdate = [
			{ key: 'enable_documentation', value: String(form.data.enable_documentation), group_name: 'SYSTEM' },
			{ key: 'maintenance_mode', value: String(form.data.maintenance_mode), group_name: 'SYSTEM' },
			{ key: 'maintenance_message', value: form.data.maintenance_message, group_name: 'SYSTEM' },
			{ key: 'source_logo_favicon', value: form.data.source_logo_favicon, group_name: 'SYSTEM' },
			{ key: 'histats_tracking_code', value: form.data.histats_tracking_code, group_name: 'SYSTEM' }
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
