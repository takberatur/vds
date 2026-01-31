<script lang="ts">
	import { MetaTags, type MetaTagsProps } from 'svelte-meta-tags';

	let { data } = $props();
	let metaTags = $derived<MetaTagsProps | undefined>(data.pageMetaTags);

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
<h1 class="my-3 text-3xl font-semibold">Blog</h1>

<ul class="flex flex-wrap gap-7">
	{#each data.posts as { meta, path }}
		<li class="bg-base-100 rounded-box w-fit border p-4 shadow-md">
			<h2 class="text-2xl">
				<a href={path} class="link link-primary link-hover">
					{meta.title}
				</a>
			</h2>
			<p class="text-sm font-semibold uppercase">
				{new Date(meta.publishedDate).toDateString()}
			</p>
			<div class="flex flex-wrap gap-x-2 gap-y-3">
				{#each meta.tags as tag}
					<span class="bg-base-200 rounded-box cursor-default px-2 py-1">#{tag}</span>
				{/each}
			</div>
		</li>
	{/each}
</ul>
