<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { ClassValue } from 'svelte/elements';
	import { cn } from '@/utils';

	let {
		type,
		children,
		class: className
	}: {
		type: 'blue' | 'green' | 'red' | 'yellow' | 'default';
		children: Snippet;
		class?: ClassValue;
	} = $props();
</script>

<div
	class={cn(
		'inline-flex rounded-md border border-solid border-neutral-200 px-4 py-2 dark:border-neutral-700',
		className,
		type === 'default'
			? ''
			: `bg-${type}-100 text-${type}-800 dark:bg-${type}-900 dark:text-${type}-200`
	)}
>
	{@render children()}
</div>
