<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { browser } from '$app/environment';
	import { afterNavigate } from '$app/navigation';

	let {
		publisher_id,
		ad_slot
	}: {
		publisher_id?: string | null;
		ad_slot?: string | null;
	} = $props();

	let adElement = $state<any | null>(null);
	let isVisible = $state<boolean>(false);

	async function refreshAd() {
		if (!browser || !window.adsbygoogle) return;

		await tick(); // Tunggu DOM stabil

		try {
			// Hapus status lama agar Google mau mengisi ulang slot ini
			if (adElement) {
				adElement.removeAttribute('data-adsbygoogle-status');
				adElement.innerHTML = ''; // Bersihkan konten iklan lama
				(window.adsbygoogle = window.adsbygoogle || []).push({});
			}
		} catch (e) {
			console.warn('AdSense Refresh Note:', e);
		}
	}

	onMount(() => {
		refreshAd();
	});
	afterNavigate(() => {
		refreshAd();
	});
</script>

{#if isVisible}
	<div class="h-full w-full">
		<ins
			bind:this={adElement}
			class="adsbygoogle"
			style="display:block"
			data-ad-format="autorelaxed"
			data-ad-client={publisher_id ?? 'ca-pub-4603244057078716'}
			data-ad-slot={ad_slot ?? '7476019270'}
		></ins>
	</div>
{/if}
