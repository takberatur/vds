<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { MetaTags } from 'svelte-meta-tags';
	import { superForm } from 'sveltekit-superforms';
	import { handleSubmitLoading } from '@/stores';
	import { AdminSidebarLayout, AdminPlatformUploadThumbnail } from '@/components/admin';
	import { AppAlertDialog } from '@/components/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Spinner } from '@/components/ui/spinner/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Switch } from '@/components/ui/switch/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import Icon from '@iconify/svelte';
	import { formatDate } from '@/utils/time.js';
	import { localizeHref } from '@/paraglide/runtime';

	let { data } = $props();
	let metaTags = $derived(data.pageMetaTags);
	let platform = $derived(data.platform);
	let errorMessage = $state<string | null>(null);

	// svelte-ignore state_referenced_locally
	const { form, errors, submitting, enhance } = superForm(data.form, {
		id: `edit-platform-form-${platform.slug}`,
		dataType: 'json',
		resetForm: false,
		onSubmit() {
			handleSubmitLoading(true);
			errorMessage = null;
		},
		async onUpdate(event) {
			if (event.result.type === 'failure') {
				errorMessage = event.result.data.message;
				return;
			}
			if (event.result.type === 'success') {
				handleSubmitLoading(false);
				await goto(localizeHref('/platform'));
				await invalidateAll();
			}
		},
		onError(event) {
			errorMessage = event.result.error.message;
			handleSubmitLoading(false);
		}
	});

	function generateSlug(name: string) {
		if (!name) return;

		const slug = name
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/^-+|-+$/g, '');

		$form.slug = slug;

		return slug;
	}
	function validateSlug(slug: string): boolean {
		return !slug || /^[a-z0-9-]+$/.test(slug);
	}
	function validUrl(url?: string | null) {
		if (!url) return false;
		try {
			new URL(url);
			return true;
		} catch (_) {
			return false;
		}
	}
	function onUrlPatternChange(value: string) {
		if (!validUrl(value)) {
			$errors.url_pattern = ['Invalid URL pattern'];
		}
	}

	let configEntries = $state<{ key: string; value: string }[]>(
		Object.entries($form.config || {}).map(([key, value]) => ({ key, value: String(value) }))
	);

	function addConfigEntry() {
		configEntries.push({ key: '', value: '' });
	}

	function removeConfigEntry(index: number) {
		configEntries.splice(index, 1);
	}

	$effect(() => {
		const newConfig: Record<string, any> = {};
		for (const entry of configEntries) {
			if (entry.key) {
				newConfig[entry.key] = entry.value;
			}
		}
		$form.config = newConfig;
	});

	const typeOptions = [
		{ value: 'youtube', label: 'Youtube' },
		{ value: 'tiktok', label: 'Tiktok' },
		{ value: 'instagram', label: 'Instagram' },
		{ value: 'facebook', label: 'Facebook' },
		{ value: 'twitter', label: 'Twitter' },
		{ value: 'vimeo', label: 'Vimeo' },
		{ value: 'dailymotion', label: 'Dailymotion' },
		{ value: 'rumble', label: 'Rumble' },
		{ value: 'any-video-downloader', label: 'Any Video Downloader' },
		{ value: 'youtube-to-mp3', label: 'Youtube to MP3' },
		{ value: 'snackvideo', label: 'Snack Video' }
	];
</script>

<MetaTags {...metaTags} />
<AdminSidebarLayout
	page={`Edit platform ${platform.name}`}
	user={data.user}
	setting={data.settings}
>
	<div class="@container/main flex flex-col gap-4">
		<div class="flex-none px-4 py-4 sm:px-6">
			<Card.Root class="mb-5">
				<Card.Content>
					<div class="flex flex-col items-start gap-6 md:flex-row md:items-center">
						<div class="relative">
							<img
								src={platform.thumbnail_url}
								alt={platform.name || 'N/A'}
								class="h-38 w-auto rounded-md object-fill shadow-lg"
								onerror={() => '/default-cover.png'}
							/>
							<AdminPlatformUploadThumbnail {platform} />
						</div>
						<div class="flex-1 space-y-2">
							<div class="flex flex-col">
								<h1 class="text-2xl font-bold text-neutral-900 dark:text-white">
									{platform.name || 'N/A'}
								</h1>
								<p class="text-sm text-neutral-500 dark:text-neutral-400">
									{formatDate(new Date(platform.created_at)) || 'N/A'}
								</p>
							</div>
							<div class="mt-1 flex items-center gap-2">
								<Badge
									variant={platform.is_active ? 'default' : 'destructive'}
									class="px-4 text-xs font-semibold"
								>
									{platform.is_active ? 'Active' : 'Inactive'}
								</Badge>
								<Badge
									variant={platform.is_premium ? 'default' : 'destructive'}
									class="px-4 text-xs font-semibold"
								>
									{platform.is_premium ? 'Premium' : 'Free'}
								</Badge>
							</div>
						</div>
					</div>
				</Card.Content>
			</Card.Root>
		</div>
		<div class="space-y-4 bg-transparent">
			<ScrollArea
				class="h-[calc(100vh-350px)] space-y-4 rounded-md border border-neutral-300 px-3 py-5 dark:border-neutral-700"
			>
				<form method="POST" class="relative space-y-6" use:enhance>
					<input type="hidden" name="id" value={$form.id} />
					<Field.Group>
						<Field.Set>
							<Field.Field>
								<Field.Label for="name">Name</Field.Label>
								<div class="relative">
									<Icon
										icon="material-symbols:title"
										class="absolute top-1/2 left-3 -translate-y-1/2"
									/>
									<Input
										bind:value={$form.name}
										id="name"
										name="name"
										type="text"
										class="ps-10"
										placeholder="Enter platform name"
										autocomplete="off"
										disabled={$submitting}
										oninput={() => generateSlug($form.name)}
									/>
								</div>
								{#if $errors.name}
									<Field.Error>{$errors.name}</Field.Error>
								{/if}
							</Field.Field>
							<Field.Field>
								<Field.Label for="slug">Slug</Field.Label>
								<div class="relative">
									<Icon
										icon="material-symbols:link"
										class="absolute top-1/2 left-3 -translate-y-1/2"
									/>
									<Input
										bind:value={$form.slug}
										id="slug"
										name="slug"
										type="text"
										class="ps-10"
										placeholder="Enter platform slug"
										aria-invalid={!validateSlug($form.slug)}
										autocomplete="off"
										disabled={$submitting}
									/>
								</div>
								{#if $errors.slug}
									<Field.Error>{$errors.slug}</Field.Error>
								{:else if !validateSlug($form.slug)}
									<Field.Error
										>Slug must contain only lowercase letters, numbers, and hyphens.</Field.Error
									>
								{/if}
							</Field.Field>
							<Field.Field>
								<Field.Label for="type">Type</Field.Label>
								<div class="relative">
									<Icon
										icon="material-symbols:link"
										class="absolute top-1/2 left-3 -translate-y-1/2"
									/>
									<Select.Root type="single" name="type" bind:value={$form.type} disabled>
										<Select.Trigger class="w-full ps-10 capitalize">
											{$form.type ? $form.type : 'Select Type'}
										</Select.Trigger>
										<Select.Content>
											<Select.Group>
												<Select.Label>Type</Select.Label>
												{#each typeOptions as option (option.value)}
													<Select.Item
														value={option.value}
														label={option.label}
														disabled={option.value === 'grapes'}
													>
														{option.label}
													</Select.Item>
												{/each}
											</Select.Group>
										</Select.Content>
									</Select.Root>
								</div>
								{#if $errors.url_pattern}
									<Field.Error>{$errors.url_pattern}</Field.Error>
								{/if}
							</Field.Field>
							<Field.Field>
								<Field.Label for="url_pattern">URL Pattern</Field.Label>
								<div class="relative">
									<Icon
										icon="material-symbols:link"
										class="absolute top-1/2 left-3 -translate-y-1/2"
									/>
									<Input
										bind:value={$form.url_pattern}
										id="url_pattern"
										name="url_pattern"
										type="text"
										class="ps-10"
										placeholder="Enter platform URL pattern"
										autocomplete="off"
										disabled
									/>
								</div>
								{#if $errors.url_pattern}
									<Field.Error>{$errors.url_pattern}</Field.Error>
								{/if}
							</Field.Field>
						</Field.Set>
						<Field.Set>
							<Field.Field orientation="horizontal">
								<Field.Content>
									<Field.Label for="is_active">
										Active Status : <Badge variant={$form.is_active ? 'default' : 'destructive'}>
											{$form.is_active ? 'Active' : 'Inactive'}
										</Badge>
									</Field.Label>
									<Field.Description>
										{$form.is_active
											? 'Platform is active. Users can use this platform to download videos.'
											: 'Platform is inactive. Users cannot use this platform to download videos.'}
									</Field.Description>
								</Field.Content>
								<input type="hidden" name="is_active" value={$form.is_active} />
								<Switch
									id="is_active"
									bind:checked={$form.is_active}
									name="is_active"
									class="cursor-pointer"
									onCheckedChange={(val) => ($form.is_active = val)}
								/>
							</Field.Field>
						</Field.Set>
						<Field.Set>
							<Field.Field orientation="horizontal">
								<Field.Content>
									<Field.Label for="is_premium">
										Premium Status : <Badge variant={$form.is_premium ? 'default' : 'destructive'}>
											{$form.is_premium ? 'Premium' : 'Free'}
										</Badge>
									</Field.Label>
									<Field.Description>
										{$form.is_premium
											? 'Platform is premium. Users can use this platform to download videos when they are subscribed.'
											: 'Platform is free. Users can use this platform to download videos without any subscription.'}
									</Field.Description>
								</Field.Content>
								<input type="hidden" name="is_premium" value={$form.is_premium} />
								<Switch
									id="is_premium"
									bind:checked={$form.is_premium}
									name="is_premium"
									class="cursor-pointer"
									onCheckedChange={(val) => ($form.is_premium = val)}
								/>
							</Field.Field>
						</Field.Set>
						<Field.Set>
							<div class="space-y-3 py-2">
								<div class="flex items-center justify-between">
									<Field.Label>Configuration</Field.Label>
									<Button
										size="sm"
										variant="outline"
										onclick={addConfigEntry}
										type="button"
										class="h-8"
									>
										<Icon icon="material-symbols:add" class="mr-2" />
										Add Config
									</Button>
								</div>

								{#if configEntries.length === 0}
									<div
										class="rounded-md border border-dashed py-6 text-center text-sm text-neutral-500"
									>
										No configuration set
									</div>
								{:else}
									<div class="space-y-2">
										{#each configEntries as entry, i}
											<div class="flex items-start gap-2">
												<div class="flex-1">
													<Input placeholder="Key" bind:value={entry.key} disabled={$submitting} />
												</div>
												<div class="flex-1">
													<Input
														placeholder="Value"
														bind:value={entry.value}
														disabled={$submitting}
													/>
												</div>
												<Button
													variant="ghost"
													size="icon"
													onclick={() => removeConfigEntry(i)}
													type="button"
													class="text-red-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
													disabled={$submitting}
												>
													<Icon icon="material-symbols:delete-outline" />
												</Button>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						</Field.Set>
						<Field.Field>
							<Button type="submit" disabled={$submitting} class="w-full">
								{#if $submitting}
									<Spinner class="mr-2" />
									Please wait
								{:else}
									Update platform
								{/if}
							</Button>
						</Field.Field>
					</Field.Group>
				</form>
			</ScrollArea>
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
</AdminSidebarLayout>
