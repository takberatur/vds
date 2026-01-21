<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { MetaTags } from 'svelte-meta-tags';
	import { superForm } from 'sveltekit-superforms';
	import { AuthLayout } from '@/components/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import * as Password from '$lib/components/ui-extras/password';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import type { ZxcvbnResult } from '@zxcvbn-ts/core';
	import Icon from '@iconify/svelte';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { localizeHref } from '@/paraglide/runtime';

	let { data } = $props();
	let metaTags = $derived(data.pageMetaTags);

	let showConfirmPassword = $state(false);
	let errorMessage = $state<string | undefined>(undefined);
	let successMessage = $state<string | undefined>(undefined);
	let passwordInput = $state<string | undefined>('');
	const SCORE_NAMING = ['Poor', 'Weak', 'Average', 'Strong', 'Secure'];
	let strength = $state<ZxcvbnResult>();

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
			await goto(localizeHref('/login'));
		},
		onError(event) {
			errorMessage = event.result.error.message;
		}
	});

	$effect(() => {
		if (!data.token || data.token?.trim().length === 0) {
			errorMessage = 'Token is required';
			return;
		}
		$form.token = data.token;
	});
</script>

<MetaTags {...metaTags} />
<AuthLayout webSetting={data.settings?.WEBSITE} lang={data.lang}>
	<div class="flex w-full flex-col items-start gap-y-6 rounded-lg bg-accent px-5 py-8">
		<div class="w-full">
			<h1 class="text-2xl font-bold text-white">Reset Password</h1>
			<p class="text-sm text-muted-foreground">Enter your new password below.</p>
		</div>
		{#if errorMessage}
			<Alert.Root variant="destructive">
				<Icon icon="mingcute:warning-line" class="size-4" />
				<Alert.Title>Error</Alert.Title>
				<Alert.Description>{errorMessage}</Alert.Description>
			</Alert.Root>
		{/if}
		<form method="POST" class="w-full" use:enhance>
			<Field.Group class="Root">
				<Field.Field>
					<Input
						bind:value={$form.token}
						name="token"
						type="hidden"
						aria-invalid={!!$errors.token}
						autocomplete="on"
					/>
				</Field.Field>
				<Field.Field>
					<Field.Label for="new_password" class="capitalize">
						New Password
						<span class="text-red-500 dark:text-red-400">*</span>
					</Field.Label>
					<div class="relative">
						<Icon icon="material-symbols:key" class="absolute top-4.5 left-3 -translate-y-1/2" />
						<Password.Root minScore={2}>
							<Password.Input
								bind:value={passwordInput}
								name="new_password"
								class="ps-10 pe-10"
								disabled={$submitting}
								placeholder="New Password"
								autocomplete="new-password"
								oninput={(e) => {
									$form.new_password = (e.target as HTMLInputElement).value;
								}}
							>
								<Password.ToggleVisibility />
							</Password.Input>
							<div class="flex flex-col gap-1">
								<Password.Strength bind:strength />
							</div>
						</Password.Root>
					</div>
					{#if $errors.new_password}
						<Field.Error>{$errors.new_password}</Field.Error>
					{/if}
				</Field.Field>
				<Field.Field>
					<Field.Label for="confirm_password" class="capitalize">
						Confirm Password
						<span class="text-red-500 dark:text-red-400">*</span>
					</Field.Label>
					<div class="relative">
						<Icon icon="material-symbols:key" class="absolute top-1/2 left-3 -translate-y-1/2" />
						<Input
							bind:value={$form.confirm_password}
							name="confirm_password"
							type={showConfirmPassword ? 'text' : 'password'}
							class="ps-10 pe-10"
							placeholder="Confirm Password"
							aria-invalid={!!$errors.confirm_password}
							autocomplete="new-password"
						/>

						<Button
							variant="ghost"
							size="icon"
							class="absolute top-1/2 right-1 size-8 -translate-y-1/2 cursor-pointer"
							onclick={() => (showConfirmPassword = !showConfirmPassword)}
						>
							<Icon
								icon={showConfirmPassword ? 'mdi:eye' : 'mdi:eye-off'}
								class="absolute top-1/2 right-3 -translate-y-1/2 cursor-pointer"
							/>
						</Button>
					</div>
					{#if $errors.confirm_password}
						<Field.Error>{$errors.confirm_password}</Field.Error>
					{/if}
				</Field.Field>
				<Field.Field>
					<Button type="submit" disabled={$submitting}>
						{#if $submitting}
							<Spinner />
						{/if}
						{$submitting ? 'Please wait' : 'Change Password'}
					</Button>
				</Field.Field>
			</Field.Group>
		</form>
	</div>
	<div class="rounded-lg px-5 py-8 text-neutral-300">
		<div class="flex w-full flex-col items-start">
			<p class="text-sm">
				Reset Password Terms
				<a href={localizeHref('/terms')}> Terms of Service </a>
				{' and '}
				<a href={localizeHref('/privacy')}> Privacy Policy </a>
			</p>
		</div>
	</div>
</AuthLayout>
