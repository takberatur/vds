<script lang="ts">
	import { page } from '$app/state';
	import { goto, invalidateAll } from '$app/navigation';
	import { MetaTags } from 'svelte-meta-tags';
	import { handleSubmitLoading } from '@/stores';
	import { AdminSidebarLayout, AdminPlatformTable } from '@/components/admin';
	import { DateRangeInput, CardSpotlight } from '@/components/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import Icon from '@iconify/svelte';
	import { updateUrlParams, createPlatformManager } from '@/stores/query.js';
	import { localizeHref } from '@/paraglide/runtime';

	let { data } = $props();
	let metaTags = $derived(data.pageMetaTags);

	const queryManager = createPlatformManager();
	let query = $state(queryManager.parse(page.url));

	$effect(() => {
		query = queryManager.parse(page.url);
	});

	let dateRange = $derived({
		start: query.date_from,
		end: query.date_to
	});

	async function updateQuery(updates: Partial<typeof query>, resetPage = false) {
		handleSubmitLoading(true);
		await updateUrlParams(goto, page.url, updates, {
			resetPage,
			replaceState: true,
			invalidateAll: true
		});
		handleSubmitLoading(false);
	}

	async function handleDateChange(range: { start: string; end: string } | null) {
		if (range) {
			await updateQuery({ date_from: range.start, date_to: range.end }, true);
		}
	}

	async function resetFilters() {
		const url = new URL(page.url);
		url.search = '';
		await goto(url.toString(), { replaceState: true, invalidateAll: true });
	}

	async function refresh() {
		handleSubmitLoading(true);
		await goto(page.url.toString(), { replaceState: true, invalidateAll: true });
		handleSubmitLoading(false);
		await invalidateAll();
	}
	async function handleClickEdit(data?: Platform) {
		await goto(localizeHref(`/platform/${data?.id}`));
	}
</script>

<MetaTags {...metaTags} />
<AdminSidebarLayout page="Platform" user={data.user} setting={data.settings}>
	<div class="@container/main flex flex-col gap-4 md:gap-6">
		<div class="flex-none px-4 py-4 sm:px-6">
			<div class="space-y-1">
				<h1 class="text-2xl font-bold tracking-tight sm:text-3xl">Platform</h1>
				<p class="text-sm text-muted-foreground">Manage your platform data.</p>
			</div>
		</div>

		<div class="flex-none border-b px-4 py-4 sm:px-6">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
				<div class="w-full lg:w-auto">
					<DateRangeInput
						bind:modelValue={dateRange}
						onchange={handleDateChange}
						class="w-full lg:w-auto"
					/>
				</div>
				<div class="flex flex-wrap items-center justify-center gap-2 lg:justify-end">
					<Button variant="outline" size="sm" onclick={() => refresh()}>
						<Icon icon="material-symbols:refresh" />
						<span class="sr-only lg:not-sr-only"> Refresh data </span>
					</Button>
					<Button variant="destructive" size="sm" onclick={resetFilters}>
						<Icon icon="ic:sharp-clear-all" />
						<span class="sr-only lg:not-sr-only"> Reset filters </span>
					</Button>
				</div>
			</div>
		</div>
		<div class="relative flex flex-col gap-4 overflow-auto p-4">
			<CardSpotlight
				variant="purple"
				shadow="large"
				spotlightIntensity="medium"
				spotlight
				class="p-2"
			>
				<div class="overflow-hidden rounded-lg border bg-white/60 px-1 py-4 dark:bg-black/60">
					<AdminPlatformTable
						data={data.platform}
						{updateQuery}
						class="min-w-full"
						onreset={() => resetFilters()}
						onEdit={handleClickEdit}
					/>
				</div>
			</CardSpotlight>
		</div>
	</div>
</AdminSidebarLayout>
