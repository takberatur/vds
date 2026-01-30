<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { MetaTags } from 'svelte-meta-tags';
	import { superForm } from 'sveltekit-superforms';
	import { handleSubmitLoading } from '@/stores';
	import { AdminSidebarLayout } from '@/components/admin';
	import { AppAlertDialog } from '@/components/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Spinner } from '@/components/ui/spinner/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import Icon from '@iconify/svelte';

	let { data } = $props();
	let metaTags = $derived(data.pageMetaTags);

	let errorMessage = $state<string | null>(null);
	let successMessage = $state<string | null>(null);

	// svelte-ignore state_referenced_locally
	const { form, errors, submitting, enhance } = superForm(data.form, {
		resetForm: false,
		onSubmit() {
			handleSubmitLoading(true);
			errorMessage = null;
		},
		async onUpdate(event) {
			if (event.result.type === 'failure') {
				errorMessage = event.result.data.message;
				handleSubmitLoading(false);
				return;
			}
			if (event.result.type === 'success') {
				handleSubmitLoading(false);
				successMessage = event.result.data.message;
				await invalidateAll();
			}
		},
		onError(event) {
			errorMessage = event.result.error.message;
			handleSubmitLoading(false);
		}
	});
</script>

<MetaTags {...metaTags} />
<AdminSidebarLayout page="Cookies Setting" user={data.user} setting={data.settings}>
	<div class="@container/main flex flex-col gap-4">
		<div class="flex-none px-4 py-4 sm:px-6">
			<div class="space-y-1">
				<h1 class="text-2xl font-bold tracking-tight sm:text-3xl">Edit Cookies</h1>
				<p class="text-sm text-muted-foreground">Edit cookies for chrome browser.</p>
			</div>
		</div>
		<div class="space-y-4 rounded-md border border-neutral-300 px-3 py-5 dark:border-neutral-700">
			{#if !data.cookies.valid}
				<Alert.Root variant="destructive">
					<Icon icon="material-symbols:info" />
					<Alert.Title>Invalid Cookies Format</Alert.Title>
					<Alert.Description>
						<ul class="list-inside list-disc text-sm">
							<li>Cookies must be in the format: name=value;domain=example.com;path=/</li>
							<li>Netscape HTTP Cookie File</li>
						</ul>
					</Alert.Description>
				</Alert.Root>
			{/if}
			<form method="POST" class="relative space-y-6" use:enhance>
				<Field.Group>
					<Field.Set>
						<Field.Field>
							<Field.Label for="cookies">
								Cookies
								<span class="text-sm text-red-500 dark:text-red-400"> * </span>
							</Field.Label>
							<p class="text-sm text-muted-foreground">
								(one cookie per line, each line in the format: name=value;domain=example.com;path=/)
							</p>
							<Textarea
								bind:value={$form.cookies}
								name="cookies"
								class="resize-none overflow-auto break-all whitespace-pre-wrap"
								placeholder="Enter cookies"
								aria-invalid={!!$errors.cookies}
								autocomplete="on"
								disabled={$submitting}
							/>
							{#if $errors.cookies}
								<Field.Error>{$errors.cookies}</Field.Error>
							{/if}
						</Field.Field>
					</Field.Set>
					<Field.Field>
						<Button type="submit" disabled={$submitting} class="w-full">
							{#if $submitting}
								<Spinner class="mr-2" />
								Please wait
							{:else}
								Update Cookies
							{/if}
						</Button>
					</Field.Field>
				</Field.Group>
			</form>
		</div>
	</div>
	{#if errorMessage}
		<AppAlertDialog
			open={true}
			title="Error"
			message={errorMessage}
			type="error"
			onclose={() => (errorMessage = null)}
		/>
	{/if}
	{#if successMessage}
		<AppAlertDialog
			open={true}
			title="Success"
			message={successMessage}
			type="success"
			onclose={() => (successMessage = null)}
		/>
	{/if}
</AdminSidebarLayout>
