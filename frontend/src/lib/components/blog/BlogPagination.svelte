<script lang="ts" module>
	interface QueryOptions {
		ssearch: string;
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
	import * as Select from '@/components/ui/select';
	import { Button } from '@/components/ui/button';
	import {
		ChevronRightIcon,
		ChevronLeftIcon,
		ChevronsLeftIcon,
		ChevronsRightIcon
	} from '@lucide/svelte';
	import * as i18n from '@/paraglide/messages.js';

	let {
		data,
		updateQuery
	}: {
		data?: PaginatedResult<BlogPost> | null;
		updateQuery: (updates: Partial<QueryOptions>, resetPage: boolean) => Promise<void>;
	} = $props();

	// svelte-ignore state_referenced_locally
	let pagination = $derived(
		data?.pagination || {
			current_page: 0,
			total_pages: 0,
			total_items: 0,
			has_next: false,
			has_prev: false,
			limit: 10,
			next_page: undefined,
			prev_page: undefined
		}
	);

	let maxVisiblePages = 5;

	let pageNumbers = $derived(() => {
		const { current_page, total_pages } = pagination;
		const half = Math.floor(maxVisiblePages / 2);
		let start = Math.max(current_page - half, 1);
		let end = Math.min(start + maxVisiblePages - 1, total_pages);

		if (end - start + 1 < maxVisiblePages) {
			start = Math.max(end - maxVisiblePages + 1, 1);
		}

		return Array.from({ length: end - start + 1 }, (_, i) => start + i);
	});

	async function goToPage(pageNumber: number) {
		await updateQuery({ page: pageNumber }, false);
	}

	async function goToFirstPage() {
		await updateQuery({ page: 1 }, false);
	}

	async function goToLastPage() {
		await updateQuery({ page: pagination.total_pages }, false);
	}

	async function goToPrevPage() {
		if (pagination.prev_page) {
			await updateQuery({ page: pagination.prev_page }, false);
		}
	}

	async function goToNextPage() {
		if (pagination.next_page) {
			await updateQuery({ page: pagination.next_page }, false);
		}
	}
</script>

<div class="flex items-center justify-between px-4">
	<div class="hidden flex-1 text-sm text-muted-foreground lg:flex">
		{#if data?.pagination}
			<span class="ml-2">
				{i18n.pagination_showing()}
				{(pagination.current_page - 1) * pagination.limit + 1}
				{i18n.pagination_to()}
				{Math.min(pagination.current_page * pagination.limit, pagination.total_items)}
				{i18n.pagination_of()}
				{pagination.total_items}
				{i18n.pagination_posts()}
				{i18n.pagination_results()}
			</span>
		{/if}
	</div>
	<div class="flex w-full items-center justify-between gap-8 lg:w-fit">
		<div class="flex items-center gap-2">
			<p class="hidden text-sm font-medium lg:flex">
				{i18n.pagination_rows_per_page()}
			</p>
			<Select.Root
				type="single"
				value={String(pagination.limit)}
				onValueChange={(value) => {
					if (value) {
						updateQuery({ limit: Number(value) }, true);
					}
				}}
			>
				<Select.Trigger class="h-8 w-17.5">
					{String(pagination.limit)}
				</Select.Trigger>
				<Select.Content side="top">
					{#each [3, 5, 10, 20, 30, 40, 50, 100] as limit (limit)}
						<Select.Item value={String(limit)}>
							{limit}
						</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>

		<!-- Page info -->
		<div class="flex w-fit items-center justify-center text-sm font-medium">
			{i18n.pagination_pages()}
			{pagination.current_page}
			{i18n.pagination_of()}
			{pagination.total_pages}
		</div>

		<!-- Navigation buttons -->
		<div class="flex items-center gap-2">
			<Button
				variant="outline"
				class="hidden size-8 p-0 lg:flex"
				onclick={() => {
					goToFirstPage();
				}}
				disabled={!pagination.has_prev}
			>
				<span class="sr-only">{i18n.pagination_first()}</span>
				<ChevronsLeftIcon class="size-4" />
			</Button>
			<Button
				variant="outline"
				class="size-8 p-0"
				onclick={() => {
					goToPrevPage();
				}}
				disabled={!pagination.has_prev}
			>
				<span class="sr-only">{i18n.pagination_previous()}</span>
				<ChevronLeftIcon class="size-4" />
			</Button>
			{#each pageNumbers() as page (page)}
				<Button
					variant={page === pagination.current_page ? 'secondary' : 'default'}
					class="size-8 p-0"
					disabled={page === pagination.current_page}
					onclick={() => {
						goToPage(page);
					}}
				>
					{page}
				</Button>
			{/each}
			<Button
				variant="outline"
				class="size-8 p-0"
				onclick={() => {
					goToNextPage();
				}}
				disabled={!pagination.has_next}
			>
				<span class="sr-only">{i18n.pagination_next()}</span>
				<ChevronRightIcon class="size-4" />
			</Button>
			<Button
				variant="outline"
				class="hidden size-8 p-0 lg:flex"
				onclick={() => {
					goToLastPage();
				}}
				disabled={!pagination.has_next}
			>
				<span class="sr-only">{i18n.pagination_last()}</span>
				<ChevronsRightIcon class="size-4" />
			</Button>
		</div>
	</div>
</div>
