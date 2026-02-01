<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { MetaTags, type MetaTagsProps } from 'svelte-meta-tags';
	import { Separator } from '@/components/ui/separator';
	import { PostPreview, BlogPagination, BlogFilter } from '@/components/blog';
	import { updateUrlParams, createBlogPostManager } from '@/stores/query.js';
	import { handleSubmitLoading } from '@/stores';
	import * as i18n from '@/paraglide/messages.js';

	let { data } = $props();
	let metaTags = $derived<MetaTagsProps | undefined>(data.pageMetaTags);
	let postData = $derived(data.posts.data?.map((p) => p.meta) || []);

	const queryManager = createBlogPostManager();
	let query = $state(queryManager.parse(page.url));

	$effect(() => {
		query = queryManager.parse(page.url);
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

	async function resetFilters() {
		const url = new URL(page.url);
		url.search = '';
		await goto(url.toString(), { replaceState: true, invalidateAll: true });
	}

	if (typeof window !== 'undefined') {
		// svelte-ignore state_referenced_locally
		const initial = metaTags;

		metaTags = undefined;

		$effect(() => {
			metaTags = initial;
		});
	}
</script>

<MetaTags {...metaTags} />
<div class="container my-4">
	<div
		class="prose dark:prose-invert prose-headings:font-heading prose-headings:font-bold prose-headings:leading-tight hover:prose-a:text-accent-foreground prose-a:prose-headings:no-underline mx-auto max-w-5xl"
	>
		<h1 class="mt-0 text-3xl font-semibold">{i18n.latest_blog()}</h1>
		<Separator class="my-4" />
		<div class="my-4 rounded-md bg-neutral-100 p-4 dark:bg-neutral-800">
			<BlogFilter data={postData} {updateQuery} onreset={resetFilters} />
		</div>
		<div class="grid grid-flow-row gap-2">
			{#each postData as post}
				<PostPreview {post} />
			{/each}
		</div>
		<div class="mt-4">
			<BlogPagination data={data.posts} {updateQuery} />
		</div>
	</div>
</div>
