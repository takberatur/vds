<script lang="ts" module>
	type CTAProps = {
		title: string;
		description?: string;
		buttonText: string;
		class?: ClassValue;
	};
</script>

<script lang="ts">
	import type { ClassValue } from 'svelte/elements';
	import { cn } from '@/utils';
	import { Input } from '@/components/ui/input';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Button } from '@/components/ui/button';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import Icon from '@iconify/svelte';
	import { toast } from '@/stores';

	let { title, description, buttonText, class: className }: CTAProps = $props();

	let form = $state({
		email: ''
	});
	let isSubmitting = $state(false);
	let errorMessage = $state<string | undefined>(undefined);

	async function submitForm(event: SubmitEvent) {
		event.preventDefault();

		try {
			isSubmitting = true;
			errorMessage = undefined;

			const response = await fetch('/blog/newsletter', {
				method: 'POST',
				body: JSON.stringify({
					email: form.email
				})
			});
			const result = await response.json();
			if (!response.ok) {
				throw new Error(result.message || 'Failed to subscribe');
			}

			toast.success(result.message || 'Thank you for subscribing!');
			form.email = '';
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'An unknown error occurred';
		} finally {
			isSubmitting = false;
		}
	}
</script>

<section
	class={cn(
		'relative isolate my-24 overflow-hidden bg-primary py-6 text-primary-foreground',
		className
	)}
>
	<div class="p-8 md:p-12">
		<div class="mx-auto max-w-lg text-center">
			<h2 class="font-heading text-2xl font-bold md:text-3xl">{title}</h2>

			<p class="hidden text-muted-foreground sm:mt-4 sm:block">{description}</p>
		</div>

		<div class="mx-auto mt-8 max-w-xl">
			<form onsubmit={submitForm}>
				<Field.Set>
					<Field.Group>
						<Field.Field>
							<Field.Label for="identifier" class="capitalize">
								Email
								<span class="text-red-500 dark:text-red-400">*</span>
							</Field.Label>
							<div class="relative">
								<Icon icon="mdi:email" class="absolute top-1/2 left-3 -translate-y-1/2" />
								<Input
									bind:value={form.email}
									name="email"
									type="email"
									class="ps-10"
									placeholder="Enter your email"
									aria-invalid={!!errorMessage}
									autocomplete="email"
									disabled={isSubmitting}
								/>
							</div>
							{#if errorMessage}
								<Field.Error>{errorMessage}</Field.Error>
							{/if}
						</Field.Field>
						<Field.Field>
							<Button type="submit" disabled={isSubmitting}>
								{#if isSubmitting}
									<Spinner />
								{/if}
								{isSubmitting ? 'Please wait' : buttonText}
							</Button>
						</Field.Field>
					</Field.Group>
				</Field.Set>
			</form>
		</div>
	</div>

	<svg
		viewBox="0 0 1024 1024"
		class="absolute top-1/2 left-1/2 -z-10 h-256 w-5xl -translate-x-1/2"
		aria-hidden="true"
	>
		<circle cx={512} cy={512} r={512} fill="url(#gradient)" fill-opacity="0.7" />
		<defs>
			<radialGradient
				id="gradient"
				cx={0}
				cy={0}
				r={1}
				gradientUnits="userSpaceOnUse"
				gradientTransform="translate(512 512) rotate(90) scale(512)"
			>
				<stop stop-color="#7775D6" />
				<stop offset={1} stop-color="#E935C1" stop-opacity={0} />
			</radialGradient>
		</defs>
	</svg>
</section>
