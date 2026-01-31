<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { PostSchema } from '@/utils/schema.js';

	let { post, children }: { post?: PostSchema; children?: Snippet } = $props();
</script>

<article class="prose mx-auto">
	<h1>{post?.title}</h1>
	<p class="text-gray-500">{post?.description}</p>

	{@render children?.()}
</article>
