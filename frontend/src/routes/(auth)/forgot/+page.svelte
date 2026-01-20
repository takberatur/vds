<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { MetaTags } from 'svelte-meta-tags';
	import { superForm } from 'sveltekit-superforms';
	import { AuthLayout } from '@/components/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import Icon from '@iconify/svelte';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { localizeHref } from '@/paraglide/runtime';

	let { data } = $props();
	let metaTags = $derived(data.pageMetaTags);

	let errorMessage = $state<string | undefined>(undefined);
	let successMessage = $state<string | undefined>(undefined);

	// svelte-ignore state_referenced_locally
	const { form, enhance, errors, submitting } = superForm(data.form, {
		async onSubmit(input) {
			errorMessage = undefined;
			successMessage = undefined;
		},
		async onUpdate(event) {
			if (event.result.type === 'failure') {
				errorMessage = event.result.data.message;
				return;
			}
			successMessage = event.result.data.message;
			await invalidateAll();
		},
		onError(event) {
			errorMessage = event.result.error.message;
		}
	});
</script>

<MetaTags {...metaTags} />
<AuthLayout webSetting={data.settings?.WEBSITE}>
	<div class="flex w-full flex-col items-start gap-y-6 rounded-lg bg-accent px-5 py-8">
		<div class="w-full">
			<h1 class="text-2xl font-bold text-white">Forgot Password</h1>
			<p class="text-sm text-muted-foreground">Enter your email address to reset your password.</p>
		</div>
		{#if errorMessage}
			<Alert.Root variant="destructive">
				<Icon icon="mingcute:warning-line" class="size-4" />
				<Alert.Title>Error</Alert.Title>
				<Alert.Description>{errorMessage}</Alert.Description>
			</Alert.Root>
		{/if}
		{#if successMessage}
			<Alert.Root variant="default">
				<Icon icon="mingcute:check-line" class="size-4" />
				<Alert.Title>Success</Alert.Title>
				<Alert.Description>{successMessage}</Alert.Description>
			</Alert.Root>
		{:else}
			<form method="POST" class="w-full" use:enhance>
				<Field.Group class="Root">
					<Field.Field>
						<Field.Label for="email" class="capitalize">
							Email
							<span class="text-red-500 dark:text-red-400">*</span>
						</Field.Label>
						<div class="relative">
							<Icon icon="mdi:email" class="absolute top-1/2 left-3 -translate-y-1/2" />
							<Input
								bind:value={$form.email}
								name="email"
								type="email"
								class="ps-10"
								placeholder="Enter your email"
								aria-invalid={!!$errors.email}
								autocomplete="email"
							/>
						</div>
						{#if $errors.email}
							<Field.Error>{$errors.email}</Field.Error>
						{/if}
					</Field.Field>
					<Field.Field>
						<Button type="submit" disabled={$submitting}>
							{#if $submitting}
								<Spinner />
							{/if}
							{$submitting ? 'Please wait' : 'Reset Password'}
						</Button>
					</Field.Field>
				</Field.Group>
			</form>
		{/if}
		<div class="flex w-full flex-col items-center gap-2 pt-2">
			<div class="text-sm text-muted-foreground">Remember me?</div>
			<Button href={localizeHref('/login')} type="button" variant="outline" class="w-full text-sm"
				>Sign In</Button
			>
		</div>
	</div>
	<div class="rounded-lg px-5 py-8 text-neutral-300">
		<div class="flex w-full flex-col items-start">
			<p class="text-sm">
				By resetting your password, you agree to our
				<a href={localizeHref('/terms')}> Terms of Service </a>
				and
				<a href={localizeHref('/privacy')}> Privacy Policy </a>
			</p>
		</div>
	</div>
</AuthLayout>
