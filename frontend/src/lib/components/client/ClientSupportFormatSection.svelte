<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { scrollAnimation } from '@/stores';
	import { CardSpotlight } from '@/components';
	import * as i18n from '@/paraglide/messages.js';
	import { CheckCircle2 } from '@lucide/svelte';

	let elementSection = $state<HTMLElement | null>(null);
	let scrollAnimationSection = $state<ReturnType<typeof scrollAnimation.registerElement>>();
	let textElement = $state<HTMLElement | null>(null);
	let typewriter = $state<ReturnType<typeof scrollAnimation.createTypewriter>>();

	onMount(() => {
		if (elementSection) {
			scrollAnimationSection = scrollAnimation.registerElement(elementSection, {
				animationType: 'slideInRight',
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

<CardSpotlight variant="none" shadow="none" class="p-0">
	<section
		id="supported-formats"
		bind:this={elementSection}
		class="w-full bg-linear-to-br from-blue-600 to-purple-600 py-16 text-white md:py-24"
	>
		<div class="container mx-auto flex justify-center px-4 md:max-w-6xl">
			<div class="w-full text-center">
				<h2 class="mb-6 text-3xl font-bold md:text-4xl">{i18n.text_supported_formats_title()}</h2>
				<p bind:this={textElement} class="mb-8 text-lg text-blue-100">
					{i18n.text_supported_formats_description()}
				</p>

				<div class="grid w-full gap-4 md:grid-cols-3">
					<div
						class="rounded-xl bg-white/10 px-6 py-10 shadow-lg backdrop-blur-md dark:bg-white/5 dark:backdrop-blur-xl"
					>
						<CheckCircle2 class="mx-auto mb-3 h-10 w-10" />
						<h3 class="mb-2 text-xl font-bold">{i18n.text_supported_formats_format()}</h3>
						<p class="text-mute-foreground text-sm">
							{i18n.text_supported_formats_format_description()}
						</p>
					</div>
					<div
						class="rounded-xl bg-white/10 px-6 py-10 shadow-lg backdrop-blur-md dark:bg-white/5 dark:backdrop-blur-xl"
					>
						<CheckCircle2 class="mx-auto mb-3 h-10 w-10" />
						<h3 class="mb-2 text-xl font-bold">{i18n.text_supported_formats_quality()}</h3>
						<p class="text-mute-foreground text-sm">
							{i18n.text_supported_formats_quality_description()}
						</p>
					</div>
					<div
						class="rounded-xl bg-white/10 px-6 py-10 shadow-lg backdrop-blur-md dark:bg-white/5 dark:backdrop-blur-xl"
					>
						<CheckCircle2 class="mx-auto mb-3 h-10 w-10" />
						<h3 class="mb-2 text-xl font-bold">{i18n.text_supported_formats_audio()}</h3>
						<p class="text-mute-foreground text-sm">
							{i18n.text_supported_formats_audio_description()}
						</p>
					</div>
				</div>
			</div>
		</div>
	</section>
</CardSpotlight>
