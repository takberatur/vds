<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { scrollAnimation } from '@/stores';
	import { CardSpotlight } from '@/components';
	import * as i18n from '@/paraglide/messages.js';
	import { ArrowRight } from '@lucide/svelte';

	let elementSection = $state<HTMLElement | null>(null);
	let scrollAnimationSection = $state<ReturnType<typeof scrollAnimation.registerElement>>();
	let textElement = $state<HTMLElement | null>(null);
	let typewriter = $state<ReturnType<typeof scrollAnimation.createTypewriter>>();

	onMount(() => {
		if (elementSection) {
			scrollAnimationSection = scrollAnimation.registerElement(elementSection, {
				animationType: 'flipIn',
				threshold: 0.2,
				delay: 0.3,
				offset: 50,
				once: false
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
		if (scrollAnimationSection) {
			scrollAnimationSection.destroy();
		}
		if (typewriter) {
			typewriter.destroy();
		}
	});

	const steps = [
		{
			number: '1',
			title: i18n.text_how_to_download_step_copy_url(),
			description: i18n.text_how_to_download_step_copy_url_description()
		},
		{
			number: '2',
			title: i18n.text_how_to_download_step_paste_url(),
			description: i18n.text_how_to_download_step_paste_url_description()
		},
		{
			number: '3',
			title: i18n.text_how_to_download_step_save(),
			description: i18n.text_how_to_download_step_save_description()
		}
	];
</script>

<CardSpotlight
	variant="purple"
	shadow="large"
	spotlightIntensity="medium"
	spotlight
	useBorder
	class="p-0"
>
	<section
		id="how-to"
		bind:this={elementSection}
		class="bg-white/40 py-16 md:py-24 dark:bg-black/40"
	>
		<div class="container mx-auto px-4 md:max-w-6xl">
			<div class="mb-12 text-center">
				<h2 class="mb-4 text-3xl font-bold md:text-4xl">
					{i18n.text_how_to_download_title()}
				</h2>
				<p bind:this={textElement} class="text-lg text-muted-foreground">
					{i18n.text_how_to_download_description()}
				</p>
			</div>

			<div class="grid gap-8 md:grid-cols-3">
				{#each steps as step}
					<div class="relative">
						<div class="flex flex-col items-center text-center">
							<div
								class="mb-6 flex h-20 w-20 items-center justify-center rounded-full bg-linear-to-br from-blue-600 to-purple-600 text-3xl font-bold text-white shadow-lg shadow-blue-500/30 dark:bg-linear-to-br dark:from-blue-500 dark:to-purple-500 dark:text-white dark:shadow-lg dark:shadow-blue-400/30"
							>
								{step.number}
							</div>
							<h3 class="mb-3 text-xl font-bold">{step.title}</h3>
							<p class="text-muted-foreground">{step.description}</p>
						</div>
						{#if step.number !== '3'}
							<ArrowRight
								class="absolute top-10 right-0 hidden h-6 w-6 translate-x-1/2 -translate-y-1/2 text-blue-600 md:block lg:-right-8 dark:text-blue-500"
							/>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	</section>
</CardSpotlight>
