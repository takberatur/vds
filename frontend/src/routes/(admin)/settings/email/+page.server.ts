
import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { updateSettingEmail } from '$lib/utils/schema';
import { fail } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';

export const load = async ({ locals, url, parent }) => {
	const { settings, user } = locals;

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Email Setting - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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
			smtp_enabled: settings?.EMAIL.smtp_enabled ?? false,
			smtp_service: settings?.EMAIL.smtp_service ?? '',
			smtp_host: settings?.EMAIL.smtp_host ?? '',
			smtp_port: settings?.EMAIL.smtp_port ?? 0,
			smtp_user: settings?.EMAIL.smtp_user ?? '',
			smtp_password: settings?.EMAIL.smtp_password ?? '',
			from_email: settings?.EMAIL.from_email ?? '',
			from_name: settings?.EMAIL.from_name ?? '',
		},
		zod4(updateSettingEmail)
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
		const form = await superValidate(request, zod4(updateSettingEmail));
		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', '),
				error: form.errors
			});
		}

		const settingsToUpdate = [
			{ key: 'smtp_enabled', value: String(form.data.smtp_enabled), group_name: 'EMAIL' },
			{ key: 'smtp_service', value: form.data.smtp_service, group_name: 'EMAIL' },
			{ key: 'smtp_host', value: form.data.smtp_host, group_name: 'EMAIL' },
			{ key: 'smtp_port', value: String(form.data.smtp_port), group_name: 'EMAIL' },
			{ key: 'smtp_user', value: form.data.smtp_user, group_name: 'EMAIL' },
			{ key: 'smtp_password', value: form.data.smtp_password, group_name: 'EMAIL' },
			{ key: 'from_email', value: form.data.from_email, group_name: 'EMAIL' },
			{ key: 'from_name', value: form.data.from_name, group_name: 'EMAIL' }
		].filter(s => s.value !== undefined) as { key: string; value: string; group_name: string }[];

		const error = await locals.deps.settingService.updateBulkSetting(settingsToUpdate);
		if (error) {
			return fail(500, {
				form,
				message: error.message || 'Failed to update settings.'
			});
		}

		return {
			form,
			message: 'Settings updated successfully.'
		};
	}
};
