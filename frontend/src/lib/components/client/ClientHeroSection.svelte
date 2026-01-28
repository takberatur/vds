<script lang="ts">
	import { page } from '$app/state';
	import { onMount, onDestroy, getContext } from 'svelte';
	import { derived } from 'svelte/store';
	import type { SuperValidated } from 'sveltekit-superforms';
	import type { DownloadVideoSchema } from '@/utils/schema';
	import { superForm } from 'sveltekit-superforms';
	import {
		scrollAnimation,
		type AnimationOptions,
		type WebsocketStore,
		handleSubmitLoading,
		type DownloadTaskView,
		customPageLoading
	} from '@/stores';
	import { Button } from '@/components/ui/button';
	import { Input } from '@/components/ui/input';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import { ClientDialogDownloadResults } from '@/components/client';
	import { Download, Zap } from '@lucide/svelte';
	import { localizeHref } from '@/paraglide/runtime';
	import * as i18n from '@/paraglide/messages.js';
	import { cn } from '@/utils';
	import Icon from '@iconify/svelte';
	import { PUBLIC_API_URL } from '$env/static/public';

	let {
		id,
		setting,
		platforms,
		user,
		form: formData
	}: {
		id?: string;
		setting?: SettingsValue | null;
		platforms?: Platform[];
		user?: User | null;
		form: SuperValidated<DownloadVideoSchema>;
	} = $props();

	const ws = getContext<WebsocketStore>('websocket');
	const tasks = derived(ws.state, ($state) => $state.tasks);

	let errorMessage = $state<string | null>(null);
	let successMessage = $state<string | null>(null);
	let element = $state<HTMLElement | null>(null);
	let textElement = $state<HTMLElement | null>(null);
	let animation = $state<ReturnType<typeof scrollAnimation.registerElement>>();
	let typewriter = $state<ReturnType<typeof scrollAnimation.createTypewriter>>();

	// svelte-ignore state_referenced_locally
	const { form, enhance, errors, submitting } = superForm(formData, {
		resetForm: false,
		onSubmit: (input) => {
			errorMessage = null;
			successMessage = null;
		},
		async onUpdate({ result }) {
			if (result.type === 'failure') {
				handleSubmitLoading(false);
				errorMessage = result.data.message || i18n.text_failed_to_process_download();
				return;
			}
			if (result.type === 'success') {
				handleSubmitLoading(false);
				successMessage = result.data.message;
				const task = result.data.data;
				ws.upsertTaskFromApi(task);
			}
		},
		onError: ({ result }) => {
			handleSubmitLoading(false);
			errorMessage = result.error?.message || i18n.text_failed_to_process_download();
		}
	});

	const animationOptions: AnimationOptions = {
		animationType: 'fadeIn',
		threshold: 0.2,
		delay: 0.3,
		offset: 50,
		once: false
	};

	onMount(() => {
		if (element) {
			animation = scrollAnimation.registerElement(element, animationOptions);
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
		if (animation) {
			animation.destroy();
		}
		if (typewriter) {
			typewriter.destroy();
		}
	});

	function validUrl(url?: string | null) {
		if (!url) return false;
		try {
			new URL(url);
			return true;
		} catch (_) {
			return false;
		}
	}

	function onUrlVideoChange(value: string) {
		if (!validUrl(value)) {
			$errors.url = ['Invalid URL'];
		}
	}

	async function downloadVideo(
		data?: DownloadTaskView | null,
		formatId?: string | null,
		directUrl?: string | null
	) {
		if (!data) return;

		try {
			// If we have a direct URL (e.g. MinIO), use it directly
			if (directUrl) {
				const link = document.createElement('a');
				// Clean URL (remove spaces, backticks if any)
				const cleanUrl = directUrl
					.trim()
					.replace(/^`+|`+$/g, '')
					.replace(/`/g, '');
				link.href = cleanUrl;
				link.target = '_blank';
				link.rel = 'noreferrer'; // Important for some CDNs (TikTok, etc.) to accept the request
				link.download = `${data.title || 'download'}.mp4`; // Browser might ignore this for cross-origin but worth trying
				document.body.appendChild(link);
				link.click();
				document.body.removeChild(link);
				return;
			}

			customPageLoading.show();

			const params = new URLSearchParams();
			params.set('task_id', data.id);
			params.set('filename', data.title || 'download');
			if (formatId) {
				params.set('format_id', formatId);
			}

			const downloadUrl = `${PUBLIC_API_URL}/public-proxy/downloads/file/video?${params.toString()}`;
			window.location.href = downloadUrl;
		} catch (error) {
			errorMessage =
				error instanceof Error ? error.message : i18n.text_failed_to_process_download();
			console.error('Download error:', error);
		} finally {
			customPageLoading.hide();
		}
	}

	const deleteTask = (taskId?: string) => {
		if (!taskId) return;

		errorMessage = null;
		successMessage = null;
		ws.state.update((current) => {
			const updatedTasks: Record<string, DownloadTaskView> = { ...current.tasks };
			if (updatedTasks[taskId]) {
				delete updatedTasks[taskId];
			}
			return {
				...current,
				tasks: updatedTasks
			};
		});
	};

	const clearAllTasks = () => {
		errorMessage = null;
		successMessage = null;

		ws.state.update((current) => ({
			...current,
			tasks: {}
		}));
	};

	const platformIconItems = [
		{
			images: '/images/platforms/youtube.svg',
			name: 'Youtube',
			class: 'pos-1'
		},
		{
			images: '/images/platforms/facebook.svg',
			name: 'Facebook',
			class: 'pos-2'
		},
		{
			images: '/images/platforms/twitter.svg',
			name: 'Twitter',
			class: 'pos-3'
		},
		{
			images: '/images/platforms/instagram.svg',
			name: 'Instagram',
			class: 'pos-4'
		},
		{
			images: '/images/platforms/tiktok.svg',
			name: 'Tiktok',
			class: 'pos-5'
		},
		{
			images: '/images/platforms/vimeo.png',
			name: 'Vimeo',
			class: 'pos-6'
		},
		{
			images: '/images/platforms/dailymotion.svg',
			name: 'Dailymotion',
			class: 'pos-7'
		},
		{
			images: '/images/platforms/rumble.svg',
			name: 'Rumble',
			class: 'pos-8'
		},
		{
			images: '/images/platforms/any-video.svg',
			name: 'Any Video Downloader',
			class: 'pos-9'
		},
		{
			images: '/images/platforms/snackvideo.svg',
			name: 'Snack Video',
			class: 'pos-10'
		},
		{
			images: '/images/platforms/linkedin.svg',
			name: 'LinkedIn',
			class: 'pos-11'
		},
		{
			images: '/images/platforms/twitch.svg',
			name: 'Twitch',
			class: 'pos-12'
		},
		{
			images: '/images/platforms/snapchat.svg',
			name: 'Snapchat',
			class: 'pos-13'
		},
		{
			images: '/images/platforms/pinterest.svg',
			name: 'Pinterest',
			class: 'pos-14'
		},
		{
			images: '/images/platforms/baidu.svg',
			name: 'Baidu',
			class: 'pos-15'
		}
	];
</script>

<div class="relative flex h-full w-full flex-col items-center justify-center overflow-x-clip">
	{#each platformIconItems as item, i}
		<img
			src={item.images}
			alt={item.name}
			class={cn('absolute rounded-lg opacity-0 fade-in', item.class)}
		/>
	{/each}
	<section
		id={id ?? 'hero'}
		bind:this={element}
		class="relative z-10 container mx-auto px-4 py-16 md:max-w-6xl md:py-24"
	>
		<div class="mx-auto max-w-4xl text-center">
			<div
				class="mb-6 inline-flex items-center gap-2 rounded-full bg-blue-100 px-4 py-2 text-sm font-medium text-blue-700 dark:bg-blue-900 dark:text-blue-200"
			>
				<Zap class="h-4 w-4" />
				{i18n.text_hero_tagline()}
			</div>

			<h1
				class="mb-6 text-4xl font-bold tracking-tight text-neutral-900 md:text-6xl dark:text-neutral-100"
			>
				{i18n.text_hero_header()}
				<span
					class="block bg-linear-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent dark:bg-linear-to-r dark:from-purple-400 dark:to-blue-400"
				>
					{i18n.text_hero_sub_header()}
				</span>
			</h1>

			<p
				bind:this={textElement}
				class="mb-10 text-lg text-neutral-600 md:text-xl dark:text-neutral-400"
			>
				{i18n.text_hero_describtion()}
			</p>

			<!-- Download Form -->
			<div class="mx-auto mb-8 max-w-3xl">
				<div
					class="flex flex-col gap-3 rounded-2xl border border-border bg-blue-200 p-4 shadow-xl shadow-blue-100/50 md:p-6 dark:bg-blue-800/50 dark:shadow-blue-900/50"
				>
					<div class="mb-4 flex items-center justify-center gap-2">
						<Icon
							icon="mingcute:download-3-fill"
							class="h-4 w-4 text-neutral-800 dark:text-white"
						/>
						<h2 class="text-center text-xl font-semibold">
							{i18n.text_input_link_download_label_name({ name: 'Any Video Downloader' })}
						</h2>
					</div>
					<form method="POST" class="flex flex-col items-center gap-3 md:flex-row" use:enhance>
						<Input type="hidden" name="type" value={$form.type} />
						<Input type="hidden" name="user_id" value={$form.user_id} />
						<Input type="hidden" name="platform_id" value={$form.platform_id} />
						<Input type="hidden" name="app_id" value={$form.app_id} />
						<Input
							bind:value={$form.url}
							name="url"
							type="url"
							autocomplete="url"
							placeholder={i18n.text_input_link_download_placeholder({
								url: 'https://youtube.com/watch?v=...'
							})}
							class="inline-block h-14 border-neutral-200 bg-white text-base text-neutral-900 focus-visible:ring-blue-600 dark:border-neutral-600 dark:bg-black dark:text-neutral-100 dark:focus-visible:ring-purple-600"
							disabled={$submitting}
							oninput={() => onUrlVideoChange($form.url)}
						/>
						<Button
							type="submit"
							disabled={$submitting}
							class="h-10 bg-linear-to-r from-blue-600 to-purple-600 px-8 text-base font-semibold text-white shadow-lg shadow-blue-500/30 hover:from-blue-700 hover:to-purple-700 disabled:cursor-not-allowed disabled:opacity-70 md:h-12 dark:bg-linear-to-r dark:from-purple-600 dark:to-blue-600 dark:shadow-purple-500/30 dark:hover:from-purple-700 dark:hover:to-blue-700"
						>
							{#if $submitting}
								<Spinner class="mr-2 size-5" />
							{:else}
								<Download class="mr-2 h-5 w-5" />
							{/if}
							{$submitting ? i18n.text_processing() : i18n.text_download()}
						</Button>
					</form>
				</div>
				{#if $submitting}
					<p class="mt-3 text-sm text-blue-600 dark:text-blue-400">
						<Spinner class="mr-2 size-5" />
						{i18n.text_processing()}
					</p>
				{:else if successMessage}
					<p class="mt-3 text-sm text-green-500 dark:text-green-400">{successMessage}</p>
				{:else if errorMessage}
					<p class="mt-3 text-sm text-red-500 dark:text-red-400">{errorMessage}</p>
				{:else}
					<p class="mt-4 text-sm">
						✨ {i18n.text_hero_input_description()}
					</p>
				{/if}
			</div>

			{#if Object.keys($tasks).length > 0}
				<div
					class="mx-auto mt-6 max-w-3xl rounded-2xl bg-muted p-4 shadow-xl shadow-blue-100/50 md:p-6 dark:bg-neutral-900 dark:shadow-blue-900/50"
				>
					<h2 class="mb-4 text-left text-sm font-semibold text-neutral-700 dark:text-neutral-300">
						{i18n.text_progress_download_title()}
					</h2>
					<div class="space-y-4">
						<div class="flex w-full items-center justify-end">
							<Button
								type="button"
								variant="ghost"
								size="sm"
								class="text-red-500 dark:text-red-400"
								onclick={() => clearAllTasks?.()}
							>
								{i18n.text_clear_all()}
							</Button>
						</div>
						{#each Object.values($tasks).toReversed() as task (task.id)}
							<div
								class="rounded-xl border border-neutral-200 bg-white p-3 md:p-4 dark:border-neutral-600 dark:bg-neutral-950"
							>
								<div class="flex items-start gap-3">
									{#if task.thumbnail_url}
										<img
											src={task.thumbnail_url}
											alt={task.title ?? 'Thumbnail'}
											crossorigin="anonymous"
											class="h-12 w-20 shrink-0 rounded-md object-cover"
										/>
									{/if}
									<div class="flex-1 space-y-2">
										<p
											class="line-clamp-2 text-start text-sm font-semibold text-neutral-800 dark:text-neutral-200"
										>
											{task.title ?? i18n.text_progress_download_loading()}
										</p>
										<div
											class="mt-2 h-2 w-full overflow-hidden rounded-full bg-neutral-200 dark:bg-neutral-700"
										>
											<div
												class="h-full bg-linear-to-r from-blue-600 to-purple-600 transition-[width]"
												style={`width: ${Math.min(Math.max(task.progress ?? 0, 0), 100)}%`}
											></div>
										</div>
										<div
											class="mt-1 flex items-center justify-between text-xs text-neutral-500 dark:text-neutral-400"
										>
											<span class="capitalize">
												{task.status === 'processing'
													? i18n.text_processing()
													: task.status === 'queued'
														? i18n.text_queue()
														: task.status === 'completed'
															? i18n.text_complete()
															: task.status === 'failed'
																? i18n.text_failed()
																: task.status}
											</span>
											<span>{task.progress ?? 0}%</span>
										</div>

										{#if task.status === 'completed' && task.file_path}
											<div class="flex items-center gap-2">
												<ClientDialogDownloadResults {task} {downloadVideo} {deleteTask} />
											</div>
										{/if}
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<div id="platforms" class="mt-4 flex flex-wrap items-center justify-center gap-4 md:gap-6">
				{#each platforms?.filter((platform) => platform.category === 'video') as platform}
					<a
						href={localizeHref(`/video/${platform.slug}`)}
						class={cn(
							'flex cursor-pointer flex-col items-center gap-2 rounded-xl bg-muted p-4 shadow-md transition-shadow hover:shadow-lg',
							page.url.pathname === localizeHref(`/video/${platform.slug}`)
								? 'bg-sky-500 text-white dark:bg-sky-600 dark:text-white'
								: '',
							!platform.is_active ? 'hidden' : ''
						)}
					>
						<img
							src={platform.thumbnail_url}
							alt={platform.name}
							class="h-10 w-auto rounded-lg bg-transparent object-cover object-center"
						/>
						<span class="text-xs font-medium">{platform.name}</span>
					</a>
				{/each}
			</div>
		</div>
	</section>
</div>
