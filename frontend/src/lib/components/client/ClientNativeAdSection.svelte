<script lang="ts">
	import { CardSpotlight } from '@/components';
	import { AdsenseInArticleAd } from '@/components/client/monetization';

	let {
		monetize
	}: {
		monetize?: SettingMonetize | null;
	} = $props();
</script>

<CardSpotlight variant="none" shadow="none" spotlightIntensity="medium" useBorder class="p-0">
	<section class="py-16 md:py-24">
		<div class="container mx-auto px-4 md:max-w-6xl">
			<div class="h-full w-full">
				{#if monetize?.type_monetize === 'adsense'}
					<AdsenseInArticleAd
						publisher_id={monetize?.publisher_id}
						ad_slot={monetize?.native_ad_code}
					/>
				{/if}
			</div>
		</div>
	</section>
</CardSpotlight>
