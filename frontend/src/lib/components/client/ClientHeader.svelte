<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
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
	import { ClientMobileSidebar } from '$lib/components/client';
	import { Button } from '$lib/components/ui/button';
	import Icon from '@iconify/svelte';
	import { smoothScroll } from '$lib/stores';
	import { cn } from '@/utils';
	import * as i18n from '@/paraglide/messages.js';

	let {
		user,
		setting,
		lang = 'en'
	}: { user?: User | null; setting?: SettingsValue | null; lang?: Locale } = $props();

	let webSetting = $derived(setting?.WEBSITE);
	let logo = $derived(
		setting?.SYSTEM?.source_logo_favicon === 'remote' &&
			setting?.WEBSITE?.site_logo?.startsWith('http') &&
			setting?.WEBSITE?.site_logo !== ''
			? setting?.WEBSITE?.site_logo
			: '/images/logo.png'
	);

	const languageLabels: Partial<Record<Locale, string>> = {
		en: 'English',
		es: 'Español',
		de: 'German',
		pt: 'Português',
		fr: 'Français',
		id: 'Bahasa Indonesia',
		hi: 'हिन्दी',
		ar: 'العربية',
		zh: '中文',
		ru: 'Русский',
		ja: '日本語',
		tr: 'Türkçe',
		vi: 'Tiếng Việt',
		th: 'ไทย',
		el: 'Ελληνικά',
		it: 'Italiano'
	};
	const languages = availableLocales.map((code) => ({
		code,
		label: languageLabels[code] ?? code.toUpperCase()
	}));
	let currentLang = $derived(getLocale());
	let menuHidden = $derived(() => {
		const pagesRoutes = [
			localizeHref('/terms'),
			localizeHref('/privacy'),
			localizeHref('/about'),
			localizeHref('/contact'),
			localizeHref('/faq')
		];
		return pagesRoutes.some((route) => page.url.pathname.startsWith(route));
	});
	let isScrolling = $state(false);

	const handleScroll = (id: string, offset: number = 500) => {
		smoothScroll.scrollToAnchor(id, offset);
	};

	onMount(() => {
		const unsubscribe = smoothScroll.subscribe((state) => {
			isScrolling = state.isScrolling;
		});

		return () => unsubscribe();
	});

	async function handleLanguageChange(code?: string) {
		if (!code || !isLocale(code)) return;

		setLocale(code);

		const rawPath = removeLocaleFromPath(page.url.pathname);
		const localized = localizeHref(rawPath, { locale: code });

		await goto(localized);
	}

	function removeLocaleFromPath(path: string) {
		const parts = path.split('/');
		if (parts[1] && availableLocales.includes(parts[1].toLowerCase() as Locale)) {
			return '/' + parts.slice(2).join('/');
		}
		console.log(path);
		return path;
	}
</script>

<header class="sticky top-0 z-50 w-full shadow-md backdrop-blur-md dark:backdrop-blur-md">
	<div class="container mx-auto flex h-16 items-center justify-between px-4 md:max-w-7xl md:px-6">
		<a href={localizeHref('/')} class="flex items-center gap-2">
			<div class="flex h-10 w-10 items-center justify-center rounded-xl">
				<img src={logo} alt={webSetting?.site_name} class="h-full w-full rounded-xl" />
			</div>
			<span
				class="bg-linear-to-r from-blue-600 to-purple-600 bg-clip-text text-xl font-bold text-transparent dark:bg-linear-to-r dark:from-purple-400 dark:to-blue-400"
			>
				{webSetting?.site_name}
			</span>
		</a>
		<nav class={cn('hidden items-center gap-6 md:flex')}>
			{#if !menuHidden()}
				<button
					type="button"
					class="cursor-pointer text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
					onclick={() => handleScroll('#platforms', 600)}
				>
					{i18n.text_platforms()}
				</button>
				<button
					type="button"
					class="cursor-pointer text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
					onclick={() => handleScroll('#features', 100)}
				>
					{i18n.text_feature()}
				</button>
				<button
					type="button"
					class="cursor-pointer text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
					onclick={() => handleScroll('#how-to', 80)}
				>
					{i18n.text_how_to_use()}
				</button>
				<button
					type="button"
					class="cursor-pointer text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
					onclick={() => handleScroll('#supported-formats', 80)}
				>
					{i18n.text_supported_formats()}
				</button>
			{/if}
		</nav>
		<div class="hidden items-center gap-2 md:flex">
			<LightSwitch />
			<LanguageSwitcher {languages} bind:value={currentLang} onChange={handleLanguageChange} />
			{#if user}
				<Button
					href={localizeHref(`${user?.role?.name === 'admin' ? '/dashboard' : '/user'}`)}
					variant="outline"
					size="icon"
				>
					<Icon icon="material-symbols:account-circle" />
				</Button>
			{:else}
				<Button href={localizeHref('/login')} variant="outline" size="icon">
					<Icon icon="icon-park-outline:login" />
				</Button>
			{/if}
		</div>
		<div class="flex items-center md:hidden">
			<ClientMobileSidebar {user} {setting} />
		</div>
	</div>
</header>
