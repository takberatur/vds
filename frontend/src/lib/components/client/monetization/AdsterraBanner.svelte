<script lang="ts">
	import type { ClassValue } from 'svelte/elements';
	import { onMount } from 'svelte';
	import { cn } from '@/utils';
	import { useAds } from '@/hooks/use-ads';

	let {
		ad,
		class: className = 'm-auto h-full w-full'
	}: {
		ad?: string | null;
		class?: ClassValue;
	} = $props();

	const { splitBannerValue } = useAds();

	let bannerData = $derived(splitBannerValue(ad));
	let adContainer = $state<HTMLDivElement | null>(null);

	onMount(() => {
		if (!adContainer || !bannerData) return;

		const iframe = document.createElement('iframe');
		iframe.width = String(bannerData.width);
		iframe.height = String(bannerData.height);
		iframe.frameBorder = '0';
		iframe.scrolling = 'no';
		iframe.style.border = 'none';
		iframe.style.overflow = 'hidden';
		iframe.style.width = '100%';
		iframe.style.height = '100%';
		iframe.setAttribute(
			'allow',
			'autoplay; fullscreen; clipboard-write; encrypted-media; picture-in-picture'
		);
		iframe.setAttribute('referrerpolicy', 'strict-origin-when-cross-origin');

		adContainer.appendChild(iframe);

		const doc = iframe.contentWindow?.document;
		if (doc) {
			doc.open();
			doc.write(`
            <!DOCTYPE html>
            <html>
            <head>
                <style>body { margin: 0; padding: 0; display: flex; justify-content: center; align-items: center; height: 100vh; }</style>
            </head>
            <body>
                <script type="text/javascript">
                    atOptions = {
                        'key' : '${bannerData.key}',
                        'format' : '${bannerData.format}',
                        'height' : ${bannerData.height},
                        'width' : ${bannerData.width},
                        'params' : {}
                    };
                <\/script>
                <script type="text/javascript" src="${bannerData.src}"><\/script>
            </body>
            </html>
        `);
			doc.close();
		}

		return () => {
			if (!adContainer) return;
			adContainer.innerHTML = '';
		};
	});
</script>

<div bind:this={adContainer} class={cn(className)}></div>
