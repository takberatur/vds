<script lang="ts" module>
	type QueryOption = {
		page: number;
		limit: number;
		search: string;
		status: string;
		sort_by: string;
		order_by: 'asc' | 'desc';
		date_from: string;
		date_to: string;
	};
</script>

<script lang="ts">
	import type { HTMLAttributes, ClassValue } from 'svelte/elements';
	import {
		type ColumnDef,
		type ColumnFiltersState,
		type PaginationState,
		type Row,
		type RowSelectionState,
		type SortingState,
		type VisibilityState,
		type Table as TableType,
		getCoreRowModel,
		getFacetedRowModel,
		getFacetedUniqueValues,
		getFilteredRowModel,
		getPaginationRowModel,
		getSortedRowModel,
		type Column
	} from '@tanstack/table-core';
	import {
		createSvelteTable,
		FlexRender,
		renderSnippet,
		renderComponent
	} from '@/components/ui/data-table';
	import * as Table from '@/components/ui/table';
	import * as DropdownMenu from '@/components/ui/dropdown-menu';
	import * as Select from '@/components/ui/select';
	import { Badge } from '@/components/ui/badge';
	import { Button } from '@/components/ui/button';
	import * as Empty from '$lib/components/ui/empty/index.js';
	import * as Tooltip from '@/components/ui/tooltip';
	import * as Avatar from '@/components/ui/avatar';
	import * as Card from '$lib/components/ui/card/index.js';
	import { AdminPlatformTableToolbar } from '@/components/admin';
	import { cn } from '@/utils';
	import {
		EyeOffIcon,
		ArrowDownIcon,
		ArrowUpIcon,
		EllipsisIcon,
		ChevronRightIcon,
		ChevronLeftIcon,
		ChevronsUpDownIcon,
		ChevronsLeftIcon,
		ChevronsRightIcon,
		XCircleIcon
	} from '@lucide/svelte';
	import Icon from '@iconify/svelte';
	import { formatTimeAgo } from '@/utils/time';

	let {
		data,
		updateQuery,
		openAdd = $bindable(),
		onEdit,
		onView,
		onreset,
		class: className = ''
	}: {
		data?: PaginatedResult<Platform> | null;
		updateQuery: (updates: Partial<QueryOption>, resetPage: boolean) => Promise<void>;
		openAdd?: boolean;
		onEdit?: (data?: Platform) => void;
		onView?: (data?: Platform) => void;
		onreset: () => Promise<void>;
		class?: ClassValue;
	} = $props();

	let rowSelection = $state<RowSelectionState>({});
	let columnVisibility = $state<VisibilityState>({});
	let columnFilters = $state<ColumnFiltersState>([]);
	let sorting = $state<SortingState>([]);

	// svelte-ignore state_referenced_locally
	let pagination = $state<PaginationState>({
		pageIndex: data?.pagination?.current_page ? data.pagination.current_page - 1 : 0,
		pageSize: data?.pagination?.limit || 10
	});

	const columns: ColumnDef<Platform>[] = [
		{
			accessorKey: 'name',
			header: ({ column }) => {
				return renderSnippet(ColumnHeader, { column, title: 'Name' });
			},
			cell: ({ row }) => {
				return renderSnippet(NameCell, {
					id: row.original.id,
					value: row.original.name
				});
			},
			filterFn: (row, id, value) => {
				return value.includes(row.getValue(id));
			}
		},
		{
			accessorKey: 'slug',
			header: ({ column }) => {
				return renderSnippet(ColumnHeader, { column, title: 'Slug' });
			},
			cell: ({ row }) => {
				return renderSnippet(SlugCell, {
					value: row.original.slug
				});
			},
			filterFn: (row, id, value) => {
				return value.includes(row.getValue(id));
			}
		},
		{
			accessorKey: 'type',
			header: ({ column }) => {
				return renderSnippet(ColumnHeader, { column, title: 'Type' });
			},
			cell: ({ row }) => {
				return renderSnippet(TypeCell, {
					// replace all regex special characters with empty string include '-'
					value: row.original.type.replace(/[.*+?^${}()|[\]\\-]/g, ' ')
				});
			},
			filterFn: (row, id, value) => {
				return value.includes(row.getValue(id));
			}
		},
		{
			accessorKey: 'status',
			header: ({ column }) => {
				return renderSnippet(ColumnHeader, { column, title: 'Status' });
			},
			cell: ({ row }) => {
				return renderSnippet(StatusCell, {
					value: row.original.is_active ? 'Active' : 'Inactive'
				});
			},
			filterFn: (row, id, value) => {
				return value.includes(row.getValue(id));
			}
		},
		{
			accessorKey: 'is_premium',
			header: ({ column }) => {
				return renderSnippet(ColumnHeader, { column, title: 'Premium' });
			},
			cell: ({ row }) => {
				return renderSnippet(PremiumCell, {
					value: row.original.is_premium
				});
			},
			filterFn: (row, id, value) => {
				return value.includes(row.getValue(id));
			}
		},
		{
			accessorKey: 'created_at',
			header: ({ column }) => {
				return renderSnippet(ColumnHeader, { column, title: 'Created At' });
			},
			cell: ({ row }) => {
				return renderSnippet(DateCell, {
					value: row.original.created_at || ''
				});
			},
			filterFn: (row, id, value) => {
				return value.includes(row.getValue(id));
			}
		},
		{
			id: 'actions',
			cell: ({ row }) => {
				return renderSnippet(RowActions, { row });
			}
		}
	];

	const table = createSvelteTable({
		get data() {
			return data?.data || [];
		},
		manualPagination: true,
		// svelte-ignore state_referenced_locally
		pageCount: data?.pagination?.total_pages || 0,

		manualSorting: true,

		manualFiltering: true,

		state: {
			get sorting() {
				return sorting;
			},
			get columnVisibility() {
				return columnVisibility;
			},
			get rowSelection() {
				return rowSelection;
			},
			get columnFilters() {
				return columnFilters;
			},
			get pagination() {
				return pagination;
			}
		},
		columns,
		enableRowSelection: true,
		onRowSelectionChange: (updater) => {
			rowSelection = typeof updater === 'function' ? updater(rowSelection) : updater;
		},
		onSortingChange: (updater) => {
			const newSorting = typeof updater === 'function' ? updater(sorting) : updater;
			sorting = newSorting;

			if (updateQuery && newSorting.length > 0) {
				const sortBy = newSorting[0]?.id || '';
				const sortOrder = newSorting[0]?.desc ? 'desc' : 'asc';

				updateQuery(
					{
						sort_by: sortBy,
						order_by: sortOrder
					},
					false
				);
			}
		},
		onColumnFiltersChange: (updater) => {
			columnFilters = typeof updater === 'function' ? updater(columnFilters) : updater;
		},
		onColumnVisibilityChange: (updater) => {
			columnVisibility = typeof updater === 'function' ? updater(columnVisibility) : updater;
		},
		onPaginationChange: (updater) => {
			const newPagination = typeof updater === 'function' ? updater(pagination) : updater;
			pagination = newPagination;

			if (updateQuery) {
				updateQuery(
					{
						page: newPagination.pageIndex + 1,
						limit: newPagination.pageSize
					},
					false
				);
			}
		},
		getCoreRowModel: getCoreRowModel(),
		getFilteredRowModel: getFilteredRowModel(),
		getPaginationRowModel: getPaginationRowModel(),
		getSortedRowModel: getSortedRowModel(),
		getFacetedRowModel: getFacetedRowModel(),
		getFacetedUniqueValues: getFacetedUniqueValues()
	});

	$effect(() => {
		if (data?.pagination) {
			pagination = {
				pageIndex: (data.pagination.current_page || 1) - 1, // Convert to 0-based
				pageSize: data.pagination.limit || 10
			};
		}
	});

	const statusOptions = [
		{
			label: 'Active',
			value: 'active'
		},
		{
			label: 'Inactive',
			value: 'inactive'
		}
	];
</script>

{#snippet NameCell({ value, id }: { value: string; id: string })}
	{@const platform = data?.data.find((label) => label.id === id)}
	<Tooltip.Provider>
		<Tooltip.Root>
			<Tooltip.Trigger class="flex items-center gap-2">
				<Avatar.Root>
					<Avatar.Image src={platform?.thumbnail_url} alt={value} />
					<Avatar.Fallback>{value.slice(0, 2)}</Avatar.Fallback>
				</Avatar.Root>
				<span class="max-w-50 truncate font-medium">
					{value}
				</span>
			</Tooltip.Trigger>
			<Tooltip.Content>
				<Card.Root class="w-full max-w-sm">
					<Card.Content>
						<div class="flex flex-col items-start gap-6 md:flex-row md:items-center">
							<div class="relative">
								<img
									src={platform?.thumbnail_url || ''}
									alt={platform?.name || 'N/A'}
									class="h-38 w-26 rounded-md object-cover shadow-lg"
									onerror={() => '/default-cover.png'}
								/>
							</div>
							<div class="flex-1 space-y-2">
								<div class="flex flex-col gap-2 md:flex-row md:items-center">
									<h1 class="text-2xl font-bold text-neutral-900 dark:text-white">
										{platform?.name || 'N/A'} ({new Date(
											platform?.created_at || ''
										).toLocaleDateString('en-US', {
											year: 'numeric',
											month: 'long',
											day: 'numeric'
										})})
									</h1>
								</div>
								<div class="mt-1 flex items-center gap-2">
									<Badge
										variant={platform?.is_active ? 'default' : 'destructive'}
										class="px-4 text-xs font-semibold"
									>
										{platform?.is_active ? 'Active' : 'Inactive'}
									</Badge>
								</div>
								<div class="mt-1 flex items-center gap-2">
									<Badge
										variant={platform?.is_premium ? 'default' : 'outline'}
										class="px-4 text-xs font-semibold"
									>
										{platform?.is_premium ? 'Premium' : 'Free'}
									</Badge>
								</div>
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			</Tooltip.Content>
		</Tooltip.Root>
	</Tooltip.Provider>
{/snippet}

{#snippet SlugCell({ value }: { value: string })}
	<div class="flex">
		<span class="max-w-125 truncate font-medium">
			{value}
		</span>
	</div>
{/snippet}

{#snippet TypeCell({ value }: { value: string })}
	<div class="flex max-w-125">
		<Badge variant="default" class="uppercase">{value}</Badge>
	</div>
{/snippet}

{#snippet StatusCell({ value }: { value: string })}
	<div class="flex max-w-125">
		<Badge variant={value === 'Active' ? 'default' : 'destructive'}>{value}</Badge>
	</div>
{/snippet}

{#snippet PremiumCell({ value }: { value: boolean })}
	<div class="flex max-w-125">
		<Badge variant={value ? 'default' : 'destructive'}>{value ? 'Premium' : 'Free'}</Badge>
	</div>
{/snippet}

{#snippet DateCell({ value }: { value: Date | string })}
	<div class="flex">
		<span class="max-w-125 truncate font-medium">
			{formatTimeAgo(value instanceof Date ? value.toISOString() : new Date(value).toISOString())}
		</span>
	</div>
{/snippet}

{#snippet RowActions({ row }: { row: Row<Platform> })}
	{@const platform = row.original}
	<DropdownMenu.Root>
		<DropdownMenu.Trigger>
			{#snippet child({ props })}
				<Button {...props} variant="ghost" class="flex h-8 w-8 p-0 data-[state=open]:bg-muted">
					<EllipsisIcon />
					<span class="sr-only">Open Menu</span>
				</Button>
			{/snippet}
		</DropdownMenu.Trigger>
		<DropdownMenu.Content class="w-40" align="end">
			{#if onView}
				<DropdownMenu.Item onclick={() => onView(platform as Platform)}>View</DropdownMenu.Item>
			{/if}
			{#if onEdit}
				<DropdownMenu.Item onclick={() => onEdit(platform as Platform)}>Edit</DropdownMenu.Item>
			{/if}
		</DropdownMenu.Content>
	</DropdownMenu.Root>
{/snippet}

{#snippet Pagination({ table }: { table: TableType<Platform> })}
	<div class="flex items-center justify-between px-4">
		<div class="hidden flex-1 text-sm text-muted-foreground lg:flex">
			{#if table.getFilteredSelectedRowModel().rows.length > 0}
				<span class="font-medium">{table.getFilteredSelectedRowModel().rows.length}</span>
				of
				<span class="font-medium">{table.getFilteredRowModel().rows.length}</span>
				selected items.
			{/if}
			{#if data?.pagination}
				<span class="ml-2">
					({data.pagination.total_items}
					total items)
				</span>
			{/if}
		</div>
		<div class="flex w-full items-center justify-between gap-8 lg:w-fit">
			<div class="flex items-center gap-2">
				<p class="hidden text-sm font-medium lg:flex">Rows per page</p>
				<Select.Root
					type="single"
					value={String(table.getState().pagination.pageSize)}
					onValueChange={(value) => {
						if (value) {
							table.setPageSize(Number(value));
						}
					}}
				>
					<Select.Trigger class="h-8 w-17.5">
						{String(table.getState().pagination.pageSize)}
					</Select.Trigger>
					<Select.Content side="top">
						{#each [10, 20, 30, 40, 50, 100] as pageSize (pageSize)}
							<Select.Item value={String(pageSize)}>
								{pageSize}
							</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			</div>

			<!-- Page info -->
			<div class="flex w-fit items-center justify-center text-sm font-medium">
				Page
				{table.getState().pagination.pageIndex + 1}
				of
				{table.getPageCount()}
			</div>

			<!-- Navigation buttons -->
			<div class="flex items-center gap-2">
				<Button
					variant="outline"
					class="hidden size-8 p-0 lg:flex"
					onclick={() => table.setPageIndex(0)}
					disabled={!table.getCanPreviousPage()}
				>
					<span class="sr-only">First page</span>
					<ChevronsLeftIcon class="size-4" />
				</Button>
				<Button
					variant="outline"
					class="size-8 p-0"
					onclick={() => table.previousPage()}
					disabled={!table.getCanPreviousPage()}
				>
					<span class="sr-only">Previous page</span>
					<ChevronLeftIcon class="size-4" />
				</Button>
				<Button
					variant="outline"
					class="size-8 p-0"
					onclick={() => table.nextPage()}
					disabled={!table.getCanNextPage()}
				>
					<span class="sr-only">Next page</span>
					<ChevronRightIcon class="size-4" />
				</Button>
				<Button
					variant="outline"
					class="hidden size-8 p-0 lg:flex"
					onclick={() => table.setPageIndex(table.getPageCount() - 1)}
					disabled={!table.getCanNextPage()}
				>
					<span class="sr-only">Last page</span>
					<ChevronsRightIcon class="size-4" />
				</Button>
			</div>
		</div>
	</div>
{/snippet}

{#snippet ColumnHeader({
	column,
	title,
	class: className,
	...restProps
}: { column: Column<Platform>; title: string } & HTMLAttributes<HTMLDivElement>)}
	{#if !column?.getCanSort()}
		<div class={className} {...restProps}>
			{title}
		</div>
	{:else}
		<div class={cn('flex items-center', className)} {...restProps}>
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button
							{...props}
							variant="ghost"
							size="sm"
							class="-ml-3 h-8 data-[state=open]:bg-accent"
						>
							<span>
								{title}
							</span>
							{#if column.getIsSorted() === 'desc'}
								<ArrowDownIcon class="ml-2 size-4" />
							{:else if column.getIsSorted() === 'asc'}
								<ArrowUpIcon class="ml-2 size-4" />
							{:else}
								<ChevronsUpDownIcon class="ml-2 size-4" />
							{/if}
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content align="start">
					<DropdownMenu.Item onclick={() => column.toggleSorting(false)}>
						<ArrowUpIcon class="mr-2 size-3.5 text-muted-foreground/70" />
						Ascending
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => column.toggleSorting(true)}>
						<ArrowDownIcon class="mr-2 size-3.5 text-muted-foreground/70" />
						Descending
					</DropdownMenu.Item>
					<DropdownMenu.Separator />
					<DropdownMenu.Item onclick={() => column.toggleVisibility(false)}>
						<EyeOffIcon class="mr-2 size-3.5 text-muted-foreground/70" />
						Hide column
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		</div>
	{/if}
{/snippet}

<div class={cn('space-y-4', className)}>
	<AdminPlatformTableToolbar {statusOptions} {table} {updateQuery} onreset={() => onreset?.()} />
	<div class="rounded-md border">
		<Table.Root>
			<Table.Header class="sticky top-0 z-10 bg-muted">
				{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
					<Table.Row>
						{#each headerGroup.headers as header (header.id)}
							<Table.Head colspan={header.colSpan}>
								{#if !header.isPlaceholder}
									<FlexRender
										content={header.column.columnDef.header}
										context={header.getContext()}
									/>
								{/if}
							</Table.Head>
						{/each}
					</Table.Row>
				{/each}
			</Table.Header>
			<Table.Body class="**:data-[slot=table-cell]:first:w-8">
				{#if table.getRowModel().rows.length > 0}
					{#each table.getRowModel().rows as row (row.id)}
						<Table.Row data-state={row.getIsSelected() && 'selected'}>
							{#each row.getVisibleCells() as cell (cell.id)}
								<Table.Cell>
									<FlexRender content={cell.column.columnDef.cell} context={cell.getContext()} />
								</Table.Cell>
							{/each}
						</Table.Row>
					{/each}
				{:else}
					<Table.Row>
						<Table.Cell colspan={columns.length} class="h-24 text-center">
							<Empty.Root>
								<Empty.Header>
									<Empty.Media variant="icon">
										<XCircleIcon />
									</Empty.Media>
									<Empty.Title>No results found</Empty.Title>
									<Empty.Description>No results found for the current filters.</Empty.Description>
								</Empty.Header>
								<Empty.Content>
									<div class="flex gap-2">
										<Button variant="outline" size="sm" onclick={onreset}>
											<Icon icon="material-symbols:clear-all" />
											Clear filters
										</Button>
									</div>
								</Empty.Content>
							</Empty.Root>
						</Table.Cell>
					</Table.Row>
				{/if}
			</Table.Body>
		</Table.Root>
	</div>
	{@render Pagination({ table })}
</div>
