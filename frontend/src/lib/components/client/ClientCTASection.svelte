<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { scrollAnimation } from '@/stores';
	import { CardSpotlight } from '@/components';
	import { Button } from '@/components/ui/button';
	import { ArrowRight } from '@lucide/svelte';
	import { smoothScroll } from '$lib/stores';
	import * as i18n from '@/paraglide/messages.js';

	let isScrolling = $state(false);
	let elementSection = $state<HTMLElement | null>(null);
	let scrollAnimationSection = $state<ReturnType<typeof scrollAnimation.registerElement>>();
	let textElement = $state<HTMLElement | null>(null);
	let typewriter = $state<ReturnType<typeof scrollAnimation.createTypewriter>>();

	const handleScroll = () => {
		smoothScroll.scrollToAnchor('#hero', 100);
	};

	onMount(() => {
		const unsubscribe = smoothScroll.subscribe((state) => {
			isScrolling = state.isScrolling;
		});

		return () => {
			unsubscribe();
		};
	});
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
</script>

<CardSpotlight
	variant="success"
	shadow="large"
	spotlightIntensity="medium"
	spotlight
	useBorder
	class="p-0"
>
	<section id="cta" bind:this={elementSection} class="py-16 md:py-24">
		<div class="container mx-auto px-4 md:max-w-6xl">
			<div
				class="mx-auto max-w-3xl rounded-3xl bg-linear-to-br from-white to-neutral-200 p-8 text-center text-neutral-900 shadow-2xl md:p-12 dark:bg-linear-to-br dark:from-neutral-700 dark:to-neutral-800 dark:text-neutral-100"
			>
				<h2 class="mb-4 text-3xl font-bold md:text-4xl">{i18n.text_cta_title()}</h2>
				<p bind:this={textElement} class="mb-8 text-lg text-muted-foreground">
					{i18n.text_cta_description()}
				</p>
				<Button
					variant="default"
					size="lg"
					class="text-base font-semibold shadow-lg hover:bg-neutral-700 dark:hover:bg-neutral-400"
					disabled={isScrolling}
					onclick={handleScroll}
				>
					{i18n.text_cta_button()}
					<ArrowRight class="ml-2 h-5 w-5" />
				</Button>
			</div>
		</div>
	</section>
</CardSpotlight>
