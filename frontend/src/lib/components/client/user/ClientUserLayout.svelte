<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';
	import {
		localizeHref,
		getLocale,
		setLocale,
		type Locale,
		locales as availableLocales,
		isLocale
	} from '@/paraglide/runtime';
	import { LanguageLabels } from '@/utils/localize-path.js';
	import { LightSwitch } from '@/components/ui-extras/light-switch';
	import { LanguageSwitcher } from '$lib/components/ui-extras/language-switcher/index.js';

	let {
		children,
		setting
	}: {
		children?: Snippet;
		setting?: SettingsValue | null;
	} = $props();

	const languages = availableLocales.map((code) => ({
		code,
		label: LanguageLabels[code] ?? code.toUpperCase()
	}));
	let currentLang = $derived(getLocale());

	async function handleLanguageChange(code?: string) {
		if (!code || !isLocale(code)) return;

		setLocale(code);

		const rawPath = removeLocaleFromPath(page.url.pathname);
		const localized = localizeHref(rawPath, { locale: code });

		await goto(localized);
	}

	function removeLocaleFromPath(path: string) {
		const parts = path.split('/');
		if (parts[1] && availableLocales.includes(parts[1] as Locale)) {
			return '/' + parts.slice(2).join('/');
		}
		return path;
	}
</script>

<div
	class="min-h-screen bg-linear-to-br from-blue-600 to-purple-600 pt-10 text-white dark:bg-linear-to-br dark:from-neutral-950 dark:to-neutral-700 dark:text-white"
>
	<div class="m-auto w-full">
	{#if children}
		{@render children()}
	{/if}
	</div>
	<div class="fixed top-2 right-2 z-50 rounded-md p-1">
		<div class="flex items-center gap-2 text-neutral-900 dark:text-neutral-50">
			<LightSwitch />
			<LanguageSwitcher {languages} bind:value={currentLang} onChange={handleLanguageChange} />
		</div>
	</div>
</div>
