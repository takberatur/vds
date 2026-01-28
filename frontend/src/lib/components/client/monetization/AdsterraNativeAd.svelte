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

	const { splitNativeAdsValue } = useAds();

	let bannerData = $derived(splitNativeAdsValue(ad));
	let adContainer = $state<HTMLDivElement | null>(null);

	onMount(() => {
		if (!adContainer || !bannerData) return;

		const iframe = document.createElement('iframe');
		iframe.style.width = '100%';
		iframe.style.height = '100%';
		iframe.style.border = 'none';
		iframe.style.overflow = 'hidden';
		iframe.scrolling = 'no';
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
                <style>
                    body { margin: 0; padding: 0; display: flex; justify-content: center; align-items: center; height: 100vh; overflow: hidden; }
                    img { max-width: 100%; height: auto; }
                </style>
            </head>
            <body>
								<script async="async" data-cfasync="${bannerData.dataCfasync}" src="${bannerData.src}"><\/script>
								<div id="{bannerData.id}"></div>
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
