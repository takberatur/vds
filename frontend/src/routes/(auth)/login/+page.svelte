<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { MetaTags } from 'svelte-meta-tags';
	import { superForm } from 'sveltekit-superforms';
	import { AuthLayout, GoogleSignIn } from '@/components/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '@/components/ui/checkbox';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import Icon from '@iconify/svelte';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { localizeHref } from '@/paraglide/runtime';

	let { data } = $props();
	let { pageMetaTags, loginForm, settings } = $derived(data);
	let metaTags = $derived(pageMetaTags);

	let passwordType = $state('password');
	let errorMessage = $state<string | undefined>(undefined);
	let isProcessing = $state(false);

	// svelte-ignore state_referenced_locally
	const {
		form: loginFormData,
		enhance: loginEnhance,
		errors: loginErrors,
		submitting: loginSubmitting
	} = superForm(loginForm, {
		id: 'login-form',
		warnings: { duplicateId: false },
		onSubmit: async (event) => {
			isProcessing = true;
			errorMessage = undefined;
		},
		async onUpdate(event) {
			if (event.result.type === 'failure') {
				errorMessage = event.result.data.message;
				return;
			}
			if (event.result.type === 'success') {
				await invalidateAll();
				isProcessing = false;
				await goto(localizeHref(data.user?.role?.name === 'admin' ? '/dashboard' : '/user'));
			}
		},
		onError(event) {
			errorMessage = event.result.error.message;
			isProcessing = false;
		}
	});
	// svelte-ignore state_referenced_locally

	// Function to handle Google Sign-In success
	const handleGoogleSuccess = async (response: google.accounts.id.CredentialResponse) => {
		isProcessing = true;
		errorMessage = undefined;

		try {
			const res = await fetch('/login', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					credential: response.credential
				})
			});

			const result = await res.json();

			if (!res.ok || !result.success) {
				errorMessage = result.message;
				return;
			}

			await goto(localizeHref(data.user?.role?.name === 'admin' ? '/dashboard' : '/user'));
			await invalidateAll();
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Google Sign-In failed';
		} finally {
			isProcessing = false;
		}
	};
</script>

<MetaTags {...metaTags} />
<AuthLayout webSetting={data.settings?.WEBSITE} lang={data.lang}>
	<div class="flex w-full flex-col items-start gap-y-6 rounded-lg bg-accent px-5 py-8">
		<div class="w-full">
			<h2 class="text-2xl font-semibold">Sign In</h2>
			<p class="text-sm text-muted-foreground">Sign in to continue</p>
		</div>
		{#if errorMessage}
			<Alert.Root variant="destructive">
				<Icon icon="mingcute:warning-line" class="size-4" />
				<Alert.Title>Error</Alert.Title>
				<Alert.Description>{errorMessage}</Alert.Description>
			</Alert.Root>
		{/if}
		<form method="POST" action="?/email" class="w-full" use:loginEnhance>
			<Field.Group>
				<Field.Field>
					<Field.Label for="identifier" class="capitalize">
						Email
						<span class="text-red-500 dark:text-red-400">*</span>
					</Field.Label>
					<div class="relative">
						<Icon icon="mdi:email" class="absolute top-1/2 left-3 -translate-y-1/2" />
						<Input
							bind:value={$loginFormData.email}
							name="email"
							type="email"
							class="ps-10"
							placeholder="Enter your email"
							aria-invalid={!!$loginErrors.email}
							autocomplete="email"
							disabled={isProcessing || $loginSubmitting}
						/>
					</div>
					{#if $loginErrors.email}
						<Field.Error>{$loginErrors.email}</Field.Error>
					{/if}
				</Field.Field>
				<Field.Field>
					<div class="flex items-center">
						<Field.Label for="password" class="capitalize">
							Password
							<span class="text-red-500 dark:text-red-400">*</span>
						</Field.Label>
						<a
							href={localizeHref('/forgot')}
							class="ml-auto text-xs underline-offset-4 hover:underline"
						>
							Forgot password?
						</a>
					</div>
					<div class="relative">
						<Icon icon="material-symbols:key" class="absolute top-1/2 left-3 -translate-y-1/2" />
						<Input
							bind:value={$loginFormData.password}
							name="password"
							type={passwordType}
							class="ps-10"
							placeholder="Enter your password"
							aria-invalid={!!$loginErrors.password}
							autocomplete="new-password"
							disabled={isProcessing || $loginSubmitting}
						/>

						<Button
							variant="ghost"
							size="icon"
							class="absolute top-1/2 right-1 size-8 -translate-y-1/2 cursor-pointer"
							onclick={() => (passwordType = passwordType === 'password' ? 'text' : 'password')}
							disabled={isProcessing || $loginSubmitting}
						>
							<Icon
								icon={passwordType === 'password' ? 'mdi:eye' : 'mdi:eye-off'}
								class="absolute top-1/2 right-3 -translate-y-1/2 cursor-pointer"
							/>
						</Button>
					</div>
					{#if $loginErrors.password}
						<Field.Error>{$loginErrors.password}</Field.Error>
					{/if}
				</Field.Field>
				<Field.Field orientation="horizontal">
					<Checkbox
						id="remember_me"
						name="remember_me"
						bind:checked={$loginFormData.remember_me}
						onCheckedChange={(value) => ($loginFormData.remember_me = value)}
						disabled={isProcessing || $loginSubmitting}
					/>
					<Field.Label for="remember_me" class="font-normal capitalize">Remember me</Field.Label>
				</Field.Field>
				<Field.Field>
					<Button type="submit" disabled={$loginSubmitting || isProcessing}>
						{#if $loginSubmitting || isProcessing}
							<Spinner />
						{/if}
						{$loginSubmitting || isProcessing ? 'Please wait' : 'Sign in'}
					</Button>
				</Field.Field>
				<Field.Separator>OR</Field.Separator>
			</Field.Group>
		</form>

		<Field.Group>
			<Field.Field class="flex w-full items-center justify-center">
				<GoogleSignIn
					disabled={isProcessing || $loginSubmitting}
					onSuccess={handleGoogleSuccess}
					onError={() => {
						errorMessage = 'Google Sign-In failed';
						isProcessing = false;
					}}
				/>
			</Field.Field>
		</Field.Group>
	</div>

	<div class="rounded-lg px-5 py-8 text-neutral-300">
		<div class="flex w-full flex-col items-start">
			<p class="text-sm">
				By signing in, you agree to our
				<a href={localizeHref('/terms')}> Terms of Service </a>
				And
				<a href={localizeHref('/privacy')}> Privacy Policy </a>
			</p>
		</div>
	</div>
</AuthLayout>
