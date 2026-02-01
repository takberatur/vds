<script lang="ts" module>
	type CTAProps = {
		title: string;
		description?: string;
		buttonText: string;
		class?: ClassValue;
	};
</script>

<script lang="ts">
	import BlogNewsLetterSubscribe from './BlogNewsLetterSubscribe.svelte';
	import { cn } from '@/utils';

	let { title, description, buttonText, class: className }: CTAProps = $props();
</script>

<BlogNewsLetterSubscribe
	class={cn(
		'not-prose mx-auto my-6 rounded-3xl bg-[radial-gradient(rgba(0,0,0,0.3)_0.5px,transparent_0.5px)] bg-size-[12px_12px] py-0 text-primary ring-1 ring-primary/20 dark:bg-[radial-gradient(rgba(255,255,255,0.3)_0.5px,transparent_0.5px)]',
		className
	)}
/>
