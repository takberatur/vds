
import { defaultMetaTags } from '@/utils/meta-tags.js';
import { superValidate } from 'sveltekit-superforms';
import { updateApplicationSchema, type UpdateApplicationSchema } from '$lib/utils/schema';
import { fail, redirect } from '@sveltejs/kit';
import { zod4 } from 'sveltekit-superforms/adapters';
import { localizeHref } from '@/paraglide/runtime';

export const load = async ({ locals, url, params }) => {
	const { settings, deps } = locals;

	const id = params.id;
	if (!id) {
		throw redirect(303, localizeHref(`/application`));
	}

	const application = await deps.applicationService.findByID(id) as Error | Application;
	if (application instanceof Error) {
		throw redirect(303, localizeHref(`/application/${encodeURIComponent(application.message)}`));
	}

	const defaultOrigin = new URL(url.pathname, url.origin).href;
	const pageMetaTags = defaultMetaTags(
		{
			path_url: defaultOrigin,
			title: `Edit Application ${application.name} - ${settings?.WEBSITE.site_name || 'Video Downloader'}`,
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
		id: application.id,
		name: application.name,
		package_name: application.package_name,
		version: application.version,
		platform: application.platform,
		enable_monetization: application.enable_monetization,
		enable_admob: application.enable_admob,
		enable_unity_ad: application.enable_unity_ad,
		enable_start_app: application.enable_start_app,
		enable_in_app_purchase: application.enable_in_app_purchase,
		is_active: application.is_active,
		admob_ad_unit_id: application.admob_ad_unit_id,
		unity_ad_unit_id: application.unity_ad_unit_id,
		start_app_ad_unit_id: application.start_app_ad_unit_id,
		admob_banner_ad_unit_id: application.admob_banner_ad_unit_id,
		admob_interstitial_ad_unit_id: application.admob_interstitial_ad_unit_id,
		admob_native_ad_unit_id: application.admob_native_ad_unit_id,
		admob_rewarded_ad_unit_id: application.admob_rewarded_ad_unit_id,
		unity_banner_ad_unit_id: application.unity_banner_ad_unit_id,
		unity_interstitial_ad_unit_id: application.unity_interstitial_ad_unit_id,
		unity_native_ad_unit_id: application.unity_native_ad_unit_id,
		unity_rewarded_ad_unit_id: application.unity_rewarded_ad_unit_id,
	} as UpdateApplicationSchema, zod4(updateApplicationSchema))

	return {
		pageMetaTags,
		form,
		settings,
		application
	};
};

export const actions = {
	default: async ({ request, locals }) => {
		const { deps } = locals;
		const form = await superValidate(request, zod4(updateApplicationSchema));

		if (!form.valid) {
			return fail(400, {
				form,
				message: Object.values(form.errors).flat().join(', ')
			});
		}

		const application = await deps.applicationService.update(form.data.id, form.data);

		if (application instanceof Error) {
			return fail(500, {
				form,
				message: application.message || 'Failed to update application'
			});
		}

		throw redirect(303, localizeHref(`/application`));
	}
}
