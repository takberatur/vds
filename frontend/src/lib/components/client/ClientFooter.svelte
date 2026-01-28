<script lang="ts">
	import { onMount } from 'svelte';
	import { localizeHref } from '@/paraglide/runtime';
	import * as i18n from '@/paraglide/messages.js';
	import { translateStore } from '@/stores';

	let {
		setting,
		platforms,
		lang = 'en'
	}: {
		setting?: SettingsValue | null;
		platforms?: Platform[];
		lang?: string;
	} = $props();

	let webSetting = $derived(setting?.WEBSITE);
	let systemSetting = $derived(setting?.SYSTEM);
	// svelte-ignore state_referenced_locally
	let translatedSiteDescription = $state(webSetting?.site_description || '');
	let translateLoading = $state(false);

	let logo = $derived(
		setting?.SYSTEM?.source_logo_favicon === 'remote' &&
			setting?.WEBSITE?.site_logo?.startsWith('http') &&
			setting?.WEBSITE?.site_logo !== ''
			? setting?.WEBSITE?.site_logo
			: '/images/logo.png'
	);

	const handleTranslate = async (value: string) => {
		translateLoading = true;
		try {
			const result = await translateStore.singleTranslate(value, {
				targetLang: lang,
				useCache: true
			});
			translatedSiteDescription = result.data.target.text;
		} catch (error) {
			console.error(error);
			return value;
		} finally {
			translateLoading = false;
		}
	};

	onMount(async () => {
		if (webSetting?.site_description) {
			await handleTranslate(webSetting?.site_description);
		}
	});
</script>

<footer class="border-t border-b border-border py-10">
	<div class="container mx-auto px-6 md:max-w-7xl">
		<div class="grid gap-8 md:grid-cols-4">
			<div class="space-y-4 md:col-span-2">
				<a href={localizeHref('/')} class="flex items-center gap-2">
					<div class="flex h-10 w-10 items-center justify-center rounded-xl">
						<img src={logo} alt={webSetting?.site_name} class="h-full w-full rounded-xl" />
					</div>
					<span
						class="block bg-linear-to-r from-blue-600 to-purple-600 bg-clip-text text-xl font-bold text-transparent dark:bg-linear-to-r dark:from-purple-400 dark:to-blue-400"
					>
						{webSetting?.site_name}
					</span>
				</a>
				{#if translateLoading}
					<p class="text-sm">{webSetting?.site_description || ''}</p>
				{:else}
					<p class="text-sm">{translatedSiteDescription || ''}</p>
				{/if}
				<a
					href={systemSetting?.play_store_app_url || 'https://play.google.com/'}
					target="_blank"
					rel="noopener noreferrer"
					class="flex max-w-max rounded-md bg-neutral-100 p-2 shadow-xl/30 shadow-neutral-500 backdrop-blur-md dark:bg-neutral-800 dark:shadow-neutral-400"
				>
					<img src="/images/play-store.png" alt="" class="h-10 w-auto" />
				</a>
				<a
					href={systemSetting?.app_store_app_url || 'https://apps.apple.com/'}
					target="_blank"
					rel="noopener noreferrer"
					class="flex max-w-max rounded-md bg-neutral-100 p-2 shadow-xl/30 shadow-neutral-500 backdrop-blur-md dark:bg-neutral-800 dark:shadow-neutral-400"
				>
					<img src="/images/app-store.png" alt="" class="h-10 w-auto" />
				</a>
			</div>

			<div>
				<h3 class="mb-4 font-semibold">{i18n.text_platforms()}</h3>
				<ul class="space-y-2 text-sm text-muted-foreground">
					{#each platforms as platform}
						<li>
							<a
								href={localizeHref(`/${platform.slug}`)}
								class="text-muted-foreground transition-colors hover:text-blue-600 dark:hover:text-blue-400"
							>
								{platform.name}
							</a>
						</li>
					{/each}
				</ul>
			</div>

			<div>
				<h3 class="mb-4 font-semibold">{i18n.text_footer_info()}</h3>
				<ul class="space-y-2 text-sm font-semibold text-muted-foreground">
					<li>
						<a
							href={localizeHref('/about')}
							class="transition-colors hover:text-blue-600 dark:hover:text-blue-400"
							>{i18n.text_footer_about_us()}</a
						>
					</li>
					<li>
						<a
							href={localizeHref('/contact')}
							class="transition-colors hover:text-blue-600 dark:hover:text-blue-400"
							>{i18n.text_footer_contact_us()}</a
						>
					</li>
					<li>
						<a
							href={localizeHref('/faq')}
							class="transition-colors hover:text-blue-600 dark:hover:text-blue-400"
							>{i18n.text_footer_faq()}</a
						>
					</li>
					<li>
						<a
							href={localizeHref('/privacy')}
							class="transition-colors hover:text-blue-600 dark:hover:text-blue-400"
							>{i18n.text_footer_privacy_policy()}</a
						>
					</li>
					<li>
						<a
							href={localizeHref('/terms')}
							class="transition-colors hover:text-blue-600 dark:hover:text-blue-400"
							>{i18n.text_footer_terms_of_service()}</a
						>
					</li>
					<li>
						<a
							href="/sitemap.xml"
							data-sveltekit-preload-data="off"
							rel="sitemap"
							class="transition-colors hover:text-blue-600 dark:hover:text-blue-400"
						>
							{i18n.text_footer_sitemap()}
						</a>
					</li>
				</ul>
			</div>
		</div>

		<div class="mt-12 border-t border-border pt-8 text-center text-sm text-muted-foreground">
			<p>
				{i18n.text_footer_copyright()} © {new Date().getFullYear()}
				<a href={localizeHref('/')} class="font-bold text-primary underline hover:text-blue-400">
					{webSetting?.site_name}
				</a>. All rights reserved.
			</p>
		</div>
	</div>
</footer>
