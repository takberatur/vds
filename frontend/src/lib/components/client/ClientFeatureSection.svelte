<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { repeatScrollAnimation, scrollAnimation } from '@/stores';
	import { CardSpotlight } from '@/components';
	import { Zap, Shield, Video, Smartphone } from '@lucide/svelte';
	import * as i18n from '@/paraglide/messages.js';

	let elementSection = $state<HTMLElement | null>(null);
	let scrollAnimationFeature = $state<ReturnType<typeof repeatScrollAnimation.registerElement>>();
	let textElement = $state<HTMLElement | null>(null);
	let typewriter = $state<ReturnType<typeof scrollAnimation.createTypewriter>>();

	onMount(() => {
		if (elementSection) {
			scrollAnimationFeature = repeatScrollAnimation.registerElement(elementSection, {
				animationType: 'slideInOutLeft',
				threshold: 0.3,
				offset: 50,
				direction: 'both',
				delay: 0.1
			});
		}
		if (textElement) {
			typewriter = scrollAnimation.createTypewriter(textElement, {
				speed: 30,
				delay: 500,
				cursor: true,
				infinite: true
			});
		}
	});

	onDestroy(() => {
		if (scrollAnimationFeature) {
			scrollAnimationFeature.destroy();
		}
		if (typewriter) {
			typewriter.destroy();
		}
	});

	const features = [
		{
			icon: Zap,
			title: i18n.text_feature_title_download_fast(),
			description: i18n.text_feature_title_download_fast_description()
		},
		{
			icon: Shield,
			title: i18n.text_feature_title_secure(),
			description: i18n.text_feature_title_secure_description()
		},
		{
			icon: Smartphone,
			title: i18n.text_feature_title_support_all_devices(),
			description: i18n.text_feature_title_support_all_devices_description()
		},
		{
			icon: Video,
			title: i18n.text_feature_title_quality(),
			description: i18n.text_feature_title_quality_description()
		}
	];
</script>

<CardSpotlight
	variant="info"
	shadow="large"
	spotlightIntensity="medium"
	spotlight
	useBorder
	class="p-0"
>
	<section id="features" bind:this={elementSection} class="py-16 md:py-24">
		<div class="container mx-auto px-4 md:max-w-6xl">
			<div class="mb-12 text-center">
				<h2 class="mb-4 text-3xl font-bold text-neutral-900 md:text-4xl dark:text-neutral-100">
					{i18n.text_feature_why_choose_us()}
				</h2>
				<p bind:this={textElement} class="text-lg text-neutral-600 dark:text-neutral-400">
					{i18n.text_feature_why_choose_us_description()}
				</p>
			</div>

			<div class="grid gap-8 md:grid-cols-2 lg:grid-cols-4">
				{#each features as feature}
					<div
						class="group flex flex-col items-center rounded-2xl border border-border bg-linear-to-br from-white to-neutral-50 p-6 text-center transition-all hover:border-blue-300 hover:shadow-xl dark:border-neutral-700 dark:bg-linear-to-br dark:from-neutral-950 dark:to-neutral-900 dark:hover:border-blue-500"
					>
						<div
							class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-xl bg-linear-to-br from-blue-600 to-purple-600 text-white transition-transform group-hover:scale-110 dark:bg-blue-500"
						>
							<feature.icon class="h-6 w-6" />
						</div>
						<h3 class="mb-2 text-center text-xl font-bold text-neutral-900 dark:text-neutral-100">
							{feature.title}
						</h3>
						<p class="text-center text-neutral-600 dark:text-neutral-400">{feature.description}</p>
					</div>
				{/each}
			</div>
		</div>
	</section>
</CardSpotlight>
