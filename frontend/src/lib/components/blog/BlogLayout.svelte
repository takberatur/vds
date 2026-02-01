<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { PostSchema } from '@/utils/schema.js';
	import BlogCallout from '$lib/components/blog/BlogCallout.svelte';
	import BlogNewsLetterCTA from './BlogNewsLetterCTA.svelte';

	let { post, children }: { post?: PostSchema; children?: Snippet } = $props();
</script>

<article class="prose mx-auto">
	{#if children}
		{@render children()}
	{/if}
</article>
