<script lang="ts" module>
	interface QueryOptions {
		search: string;
		status: string;
		tag: string;
		series: string;
		page: number;
		limit: number;
		sort_by: string;
		order_by: 'asc' | 'desc';
		year: string;
		month: string;
	}
</script>

<script lang="ts">
	import { SvelteSet } from 'svelte/reactivity';
	import { Button } from '@/components/ui/button';
	import { Input } from '@/components/ui/input';
	import { CommandSelectInput } from '@/components';
	import { PostSchema } from '@/utils/schema.js';
	import { XIcon } from '@lucide/svelte';
	import * as i18n from '@/paraglide/messages.js';

	let {
		data,
		updateQuery,
		onreset
	}: {
		data?: PostSchema[] | null;
		updateQuery: (updates: Partial<QueryOptions>, resetPage: boolean) => Promise<void>;
		onreset: () => Promise<void>;
	} = $props();

	let searchTerms = $state<string | undefined>('');
	let selectedTag = $derived(new SvelteSet<string>([]));
	let selectedSeries = $derived(new SvelteSet<string>([]));

	async function handleReset() {
		await onreset();
		searchTerms = undefined;
		selectedTag.clear();
		selectedSeries.clear();
	}

	let searchTimer: ReturnType<typeof setTimeout>;
	async function handleSearch(value: string) {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(async () => {
			await updateQuery?.({ search: value || '' }, true);
		}, 500);
	}

	async function onTagChange(value: SvelteSet<string>) {
		await updateQuery?.({ tag: [...value].join(',') }, true);
	}

	function removeDuplicateOptions(options: { label: string; value: string }[]) {
		return options.filter(
			(item, index, arr) => arr.findIndex((t) => t.value === item.value) === index
		);
	}

	const tagOptions = $derived(
		removeDuplicateOptions(
			data?.flatMap((post) => (post.tags || []).map((tag) => ({ label: tag, value: tag }))) || []
		)
	);

	const seriesOptions = $derived(
		removeDuplicateOptions(
			data?.flatMap((post) => ({
				label: post.series?.title || '',
				value: post.series?.title || ''
			})) || []
		)
	);
</script>

<div class="flex flex-col items-center gap-4 p-2 lg:flex-row lg:justify-between">
	<div class="flex w-full flex-col items-center gap-2 gap-x-4 lg:flex-row">
		<Input
			bind:value={searchTerms}
			placeholder={i18n.pagination_search_posts()}
			oninput={(e) => {
				handleSearch(e.currentTarget.value);
			}}
			class="h-8 w-full bg-white lg:w-auto dark:bg-neutral-950"
		/>
	</div>
	{#if (searchTerms && searchTerms !== '') || selectedTag.size > 0 || selectedSeries.size > 0}
		<Button variant="ghost" class="h-8 px-2 lg:px-3" onclick={handleReset}>
			{i18n.pagination_reset_filter()}
			<XIcon />
		</Button>
	{/if}
	<CommandSelectInput
		bind:selectedValue={selectedTag}
		title={i18n.tag()}
		options={tagOptions}
		onchange={onTagChange}
		variant="default"
	/>
	<CommandSelectInput
		bind:selectedValue={selectedSeries}
		title={i18n.series()}
		options={seriesOptions}
		onchange={onTagChange}
		variant="default"
	/>
</div>
