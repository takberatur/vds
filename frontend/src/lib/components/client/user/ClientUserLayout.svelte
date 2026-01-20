<script lang="ts">
	import type { Snippet } from 'svelte';
	import {
		localizeHref,
		getLocale,
		setLocale,
		type Locale,
		locales as availableLocales,
		isLocale
	} from '@/paraglide/runtime';
	import { LightSwitch } from '@/components/ui-extras/light-switch';
	import { LanguageSwitcher } from '$lib/components/ui-extras/language-switcher/index.js';

	let {
		children,
		setting
	}: {
		children?: Snippet;
		setting?: SettingsValue | null;
	} = $props();

	const languageLabels: Partial<Record<Locale, string>> = {
		en: 'English',
		es: 'Español',
		id: 'Bahasa Indonesia'
	};
	const languages = availableLocales.map((code) => ({
		code,
		label: languageLabels[code] ?? code.toUpperCase()
	}));
	let currentLang = $derived(getLocale());
</script>

<div
	class="min-h-screen bg-linear-to-br from-blue-600 to-purple-600 pt-10 text-white dark:bg-linear-to-br dark:from-neutral-950 dark:to-neutral-700 dark:text-white"
>
	<div class="m-auto w-full">
		{@render children?.()}
	</div>
	<div class="fixed top-2 right-2 z-50 rounded-md p-1">
		<div class="flex items-center gap-2 text-neutral-900 dark:text-neutral-50">
			<LightSwitch />
			<LanguageSwitcher
				{languages}
				bind:value={currentLang}
				onChange={(code: string) => {
					if (isLocale(code)) setLocale(code);
				}}
			/>
		</div>
	</div>
</div>
