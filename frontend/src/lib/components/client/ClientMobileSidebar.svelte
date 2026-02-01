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
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Avatar from '$lib/components/ui/avatar';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { LightSwitch } from '@/components/ui-extras/light-switch';
	import { LanguageSwitcher } from '$lib/components/ui-extras/language-switcher/index.js';
	import Icon from '@iconify/svelte';
	import { LanguageLabels } from '@/utils/localize-path.js';
	import { smoothScroll } from '$lib/stores';
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
	let isOpen = $state(false);

	const languages = availableLocales.map((code) => ({
		code,
		label: LanguageLabels[code] ?? code.toUpperCase()
	}));
	let currentLang = $derived(getLocale());
	let isMp3Path = $derived(page.url.pathname.startsWith(localizeHref('/mp3', { locale: lang })));
	let isScrolling = $state(false);

	const handleScroll = (id: string, offset: number = 500) => {
		smoothScroll.scrollToAnchor(id, offset);
		isOpen = false;
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
		if (parts[1] && availableLocales.includes(parts[1] as Locale)) {
			return '/' + parts.slice(2).join('/');
		}
		return path;
	}
</script>

<Sheet.Root bind:open={isOpen} onOpenChange={(val) => (isOpen = val)}>
	<Sheet.Trigger
		class={buttonVariants({ variant: 'ghost', size: 'icon' })}
		onclick={() => (isOpen = !isOpen)}
	>
		<Icon icon="heroicons-outline:menu" />
	</Sheet.Trigger>
	<Sheet.Content side="left">
		<a href={localizeHref('/')} class="flex items-center gap-2 px-6 py-3">
			<div class="flex h-10 w-10 items-center justify-center rounded-xl">
				<Avatar.Root>
					<Avatar.Image src={logo} alt={webSetting?.site_name} />
					<Avatar.Fallback>
						{webSetting?.site_name?.slice(0, 2)}
					</Avatar.Fallback>
				</Avatar.Root>
			</div>
			<span
				class="bg-linear-to-r from-blue-600 to-purple-600 bg-clip-text text-xl font-bold text-transparent dark:bg-linear-to-r dark:from-purple-400 dark:to-blue-400"
			>
				{webSetting?.site_name}
			</span>
		</a>

		<Separator />
		<div class="flex flex-col items-start gap-4 ps-6">
			<a
				href={localizeHref('/')}
				class="inline-flex cursor-pointer items-center gap-2 text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
			>
				<Icon icon="heroicons-outline:desktop-computer" />
				{i18n.home()}
			</a>
			{#if !menuHidden()}
				<button
					type="button"
					class="inline-flex cursor-pointer items-center gap-2 text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
					onclick={() => handleScroll('#platforms', 600)}
				>
					<Icon icon="heroicons-outline:desktop-computer" />
					{i18n.text_platforms()}
				</button>
				<button
					type="button"
					class="inline-flex cursor-pointer items-center gap-2 text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
					onclick={() => handleScroll('#features', 100)}
				>
					<Icon icon="heroicons-outline:light-bulb" />
					{i18n.text_feature()}
				</button>
				<button
					type="button"
					class="inline-flex cursor-pointer items-center gap-2 text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
					onclick={() => handleScroll('#how-to', 80)}
				>
					<Icon icon="heroicons-outline:question-mark-circle" />
					{i18n.text_how_to_use()}
				</button>
				<button
					type="button"
					class="inline-flex cursor-pointer items-center gap-2 text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
					onclick={() => handleScroll('#supported-formats', 80)}
				>
					<Icon icon="heroicons-outline:check-circle" />
					{i18n.text_supported_formats()}
				</button>
				<a
					href={isMp3Path ? localizeHref('/') : localizeHref('/mp3')}
					target="_blank"
					class="inline-flex cursor-pointer items-center gap-2 text-sm font-medium text-neutral-800 transition-colors hover:text-blue-600 dark:text-neutral-100 dark:hover:text-blue-400"
					onclick={() => (isOpen = false)}
				>
					<Icon icon={isMp3Path ? 'tdesign:video-filled' : 'tdesign:music-filled'} />
					{isMp3Path ? i18n.video_downloader() : i18n.mp3_downloader()}
				</a>
			{/if}
		</div>
		<Separator />
		<div class="flex items-center justify-center gap-2 ps-6">
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
	</Sheet.Content>
</Sheet.Root>
