<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import * as Empty from '$lib/components/ui/empty/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import type { DownloadTaskView, DownloadFormat } from '@/stores';
	import Icon from '@iconify/svelte';
	import * as i18n from '@/paraglide/messages.js';
	import { cn } from '@/utils';

	let {
		task,
		downloadVideo,
		deleteTask
	}: {
		task?: DownloadTaskView | null;
		onclose?: () => void;
		downloadVideo?: (
			data?: DownloadTaskView | null,
			formatId?: string | null,
			directUrl?: string | null
		) => Promise<void>;
		deleteTask?: (taskId?: string) => void;
	} = $props();

	// svelte-ignore state_referenced_locally
	console.log('download task', task);
	const twitterRegex = /^(https?:\/\/)?(www\.)?(twitter|x|twimg)\.com\/[a-zA-Z0-9_]+$/i;

	const excludeUrlType = (url?: string | null) => {
		return twitterRegex.test(url?.toLowerCase() ?? '');
	};

	function sortFormatsDesc(a: DownloadFormat, b: DownloadFormat) {
		const hA = a.height ?? 0;
		const hB = b.height ?? 0;
		if (hA !== hB) return hB - hA;

		const sA = a.filesize ?? 0;
		const sB = b.filesize ?? 0;
		if (sA !== sB) return sB - sA;

		return 0;
	}

	function isDownloadableFormat(format: DownloadFormat) {
		if (!format.url) return false;
		if (format.url.includes('.m3u8')) return false;
		// Relax vcodec check as backend might not send it for merged files
		// if (format.vcodec && format.vcodec !== 'h264') return false;
		return true;
	}

	function handleDownloadFormat(
		currentTask: DownloadTaskView | null | undefined,
		format: DownloadFormat
	) {
		if (!currentTask) return;
		downloadVideo?.(currentTask, format.format_id ?? null, format.url ?? null);
	}

	function handleOpenFile(url?: string | null, type?: string | null) {
		if (!url) return;

		// Clean URL first (remove spaces, backticks if any)
		const cleanUrl = url
			.trim()
			.replace(/^`+|`+$/g, '')
			.replace(/`/g, '');

		if (!validUrl(cleanUrl)) return;
		// check by type or domain regex

		window.open(cleanUrl, '_blank', 'noreferrer');
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

	function cleanVideoUrl(urlStr: string): string {
		try {
			const url = new URL(urlStr);
			// remove all spaces
			url.pathname = url.pathname.replace(/\s+/g, '');
			// Remove all query parameters
			url.search = '';
			// Remove hash if present
			url.hash = '';
			// (Optional) If you need to trim *after* the extension,
			// you might need regex on url.pathname if it contains extra data
			// Example: "video.mp4/ignored" -> "video.mp4"
			url.pathname = url.pathname.replace(/\.(mp4|m3u8|mkv|avi|mov).*$/i, '.$1');
			return url.toString();
		} catch (error) {
			console.error('Invalid URL', error);
			return urlStr;
		}
	}

	const excludePlatformType = ['tiktok', 'snackvideo'];
</script>

<div>
	<Dialog.Root>
		<Dialog.Trigger
			type="button"
			class={buttonVariants({
				variant: 'ghost',
				size: 'sm',
				class:
					'bg-green-600 text-sm text-white hover:bg-green-700 hover:text-white dark:bg-green-500 dark:hover:bg-green-600 dark:hover:text-white'
			})}
		>
			{i18n.text_open_result()}
		</Dialog.Trigger>
		<Dialog.Content class="sm:max-w-106.25 lg:max-w-lg">
			<Dialog.Title>{i18n.text_download_results()}</Dialog.Title>
			<Dialog.Description>
				{i18n.text_download_results_description()}
			</Dialog.Description>

			<ScrollArea class="max-h-[calc(100vh-160px)] space-y-4 px-2 py-5">
				{#if task}
					<div class="space-y-4">
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

								{#if task.status === 'completed'}
									{#if task.file_path}
										<div class="flex items-center gap-2">
											<Button
												type="button"
												variant="ghost"
												size="sm"
												class={cn(
													'bg-green-600 text-sm text-white hover:bg-green-700 hover:text-white dark:bg-green-500 dark:hover:bg-green-600 dark:hover:text-white hidden',
												)}
												onclick={() => handleOpenFile(task.file_path, task.type)}
											>
												{i18n.text_open_file()}
											</Button>
											<Button
												type="button"
												variant="ghost"
												size="sm"
												class="bg-blue-600 text-sm text-white hover:bg-blue-700 hover:text-white dark:bg-blue-500 dark:hover:bg-blue-600 dark:hover:text-white"
												onclick={() => downloadVideo?.(task)}
											>
												{i18n.text_download()}
											</Button>
											<Button
												type="button"
												variant="ghost"
												size="sm"
												class="bg-red-600 text-sm text-white hover:bg-red-700 hover:text-white dark:bg-red-500 dark:hover:bg-red-600 dark:hover:text-white"
												onclick={() => deleteTask?.(task?.id)}
											>
												{i18n.text_delete()}
											</Button>
										</div>
									{/if}

									{#if task.formats && task.formats.length > 0}
										<div class="mt-4 space-y-2">
											<p class="text-xs font-semibold text-neutral-700 dark:text-neutral-300">
												{i18n.text_available_formats()}
											</p>
											<div class="space-y-2">
												{#each task.formats
													.slice()
													.filter(isDownloadableFormat)
													.sort(sortFormatsDesc) as format, index}
													<div
														class="flex items-center justify-between rounded-md border border-neutral-200 bg-neutral-50 px-3 py-2 text-xs dark:border-neutral-700 dark:bg-neutral-900"
													>
														<div class="flex flex-col">
															<div class="flex items-center gap-2">
																<span class="font-medium text-neutral-800 dark:text-neutral-100">
																	{format.height
																		? `${format.height}p`
																		: (format.ext?.toUpperCase() ?? 'Format')}
																</span>
																{#if task?.file_path && format.url === task.file_path}
																	<span
																		class="rounded-full bg-emerald-600 px-2 py-0.5 text-[10px] font-semibold text-white uppercase"
																	>
																		Best
																	</span>
																{/if}
																{#if (format.vcodec === 'none' || format.vcodec === '') && format.acodec !== 'none' && format.acodec !== ''}
																	<span
																		class="rounded-full bg-amber-500 px-2 py-0.5 text-[10px] font-semibold text-white uppercase"
																	>
																		Audio Only
																	</span>
																{:else if (format.acodec === 'none' || format.acodec === '') && format.vcodec !== 'none' && format.vcodec !== ''}
																	<span
																		class="rounded-full bg-blue-500 px-2 py-0.5 text-[10px] font-semibold text-white uppercase"
																	>
																		Video Only
																	</span>
																{/if}
															</div>
															<span class="text-neutral-500 dark:text-neutral-400">
																{format.ext?.toUpperCase() ?? ''}
																{format.filesize
																	? ` • ${(format.filesize / (1024 * 1024)).toFixed(1)} MB`
																	: ''}
															</span>
														</div>
														<div class="flex flex-col items-center gap-2">
															<Button
																type="button"
																variant="ghost"
																size="sm"
																class="bg-blue-600 text-white hover:bg-blue-700 hover:text-white dark:bg-blue-500 dark:hover:bg-blue-600 dark:hover:text-white "
																onclick={() => handleDownloadFormat(task, format)}
															>
																{i18n.text_download()}
															</Button>
															<Button
																type="button"
																variant="ghost"
																size="sm"
																class={cn(
																	'bg-green-600 text-sm text-white hover:bg-green-700 hover:text-white dark:bg-green-500 dark:hover:bg-green-600 dark:hover:text-white hidden',
																)}
																onclick={() => handleOpenFile(format.url, task.type)}
															>
																{i18n.text_open_file()}
															</Button>
														</div>
													</div>
												{/each}
											</div>
										</div>
									{/if}
								{/if}
							</div>
						</div>
					</div>
				{:else}
					<Empty.Root>
						<Empty.Header>
							<Empty.Media variant="icon">
								<Icon icon="material-symbols:download" />
							</Empty.Media>
							<Empty.Title>{i18n.text_no_download_tasks_yet()}</Empty.Title>
							<Empty.Description>
								{i18n.text_no_download_tasks_yet_description()}
							</Empty.Description>
						</Empty.Header>
					</Empty.Root>
				{/if}
			</ScrollArea>
			<Dialog.Footer>
				<Dialog.Close>
					<Button type="button" variant="default">
						{i18n.text_close()}
					</Button>
				</Dialog.Close>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
</div>
