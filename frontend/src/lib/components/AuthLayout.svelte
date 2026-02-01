<script lang="ts">
	import type { Snippet } from 'svelte';
	import { LightSwitch } from '@/components/ui-extras/light-switch';
	import { localizeHref } from '@/paraglide/runtime';

	let {
		children,
		webSetting,
		lang
	}: { children?: Snippet; webSetting?: SettingWeb; lang: string } = $props();

	let logo = $derived(webSetting?.site_logo || '/images/logo.png');
</script>

<main
	class="relative flex min-h-screen w-full items-center justify-center bg-linear-to-br from-blue-500 via-purple-500 to-pink-500"
>
	<div class="absolute top-4 right-4 z-10">
		<div class="flex items-center gap-2">
			<LightSwitch />
		</div>
	</div>
	<div class="mx-auto h-full w-full px-4 md:max-w-7xl md:px-12 lg:px-20">
		<div class="relative">
			<div class="mx-auto flex w-full max-w-lg flex-col items-center pt-4 md:pt-20">
				<div class="flex w-full flex-col items-center justify-center py-5">
					<a href={localizeHref('/')} class="flex items-center gap-x-2">
						<img src={logo} alt={webSetting?.site_name || 'Video Downloader'} class="h-14 w-auto" />
					</a>
					<h2 class="text-2xl font-bold text-white">
						{webSetting?.site_name || 'Video Downloader'}
					</h2>
				</div>

	{#if children}
		{@render children()}
	{/if}
			</div>
		</div>
	</div>
</main>
