<script lang="ts">
	import { onMount } from 'svelte';
	import { MetaTags, type MetaTagsProps } from 'svelte-meta-tags';
	import {
		Accordion,
		AccordionContent,
		AccordionItem,
		AccordionTrigger
	} from '@/components/ui/accordion';
	import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
	import { Separator } from '@/components/ui/separator';
	import { cn } from '@/utils';
	import { BlogTableOfContent } from '@/components/blog/index.js';
	import { format, parseISO } from 'date-fns';
	import { localizeHref } from '@/paraglide/runtime.js';
	import { Home } from '@lucide/svelte';

	let { data } = $props();
	let metaTags = $derived<MetaTagsProps | undefined>(data.pageMetaTags);
	let postData = $derived(data.posts);
	let PostComponent = $state<any | null>(null);

	const modules = import.meta.glob('/src/routes/content/**/*.md');

	onMount(async () => {
		const importer = modules[`/src/routes/content/${data.posts.meta.slug}.md`];

		if (importer) {
			const mod = await importer();
			PostComponent = (mod as any).default;
		}
	});

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
<div class="container mx-auto my-4 px-4 pb-10 md:max-w-5xl">
	<nav aria-label="Breadcrumb">
		<ol
			role="list"
			class="hidden items-center gap-1 text-sm text-muted-foreground md:flex md:flex-row"
		>
			<li>
				<a
					href={localizeHref('/', { locale: data.lang })}
					class="block transition hover:text-muted-foreground/70"
					aria-label="Go to Home"
				>
					<span class="sr-only"> Home </span>
					<Home size={14} />
				</a>
			</li>
			<li class="rtl:rotate-180">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-4 w-4"
					viewBox="0 0 20 20"
					fill="currentColor"
				>
					<path
						fill-rule="evenodd"
						d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
						clip-rule="evenodd"
					/>
				</svg>
			</li>
			<li>
				<a
					href={localizeHref('/blog', { locale: data.lang })}
					class="block transition hover:text-muted-foreground/70"
				>
					Blog
				</a>
			</li>
			<li class="rtl:rotate-180">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					class="h-4 w-4"
					viewBox="0 0 20 20"
					fill="currentColor"
				>
					<path
						fill-rule="evenodd"
						d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
						clip-rule="evenodd"
					/>
				</svg>
			</li>

			<li>
				<!-- svelte-ignore a11y_invalid_attribute -->
				<a href="#" class="block transition hover:text-muted-foreground/70">
					{data.posts.meta.title}
				</a>
			</li>
		</ol>
	</nav>

	<div class="flex flex-col lg:flex-row">
		<div class="lg:hidden">
			<div class="mt-1 mb-4 text-sm leading-snug text-muted-foreground">
				<p class="mb-2">{`${data.posts.readingTime} read`}</p>
				<time
					>Originally published: {format(parseISO(data.posts.meta.publishedDate), 'LLLL d, yyyy')}
				</time>
				{#if data.posts.meta.lastUpdatedDate}
					<br />
					<time
						>Last updated: {format(parseISO(data.posts.meta.lastUpdatedDate), 'LLLL d, yyyy')}
					</time>
				{/if}
			</div>
			<Accordion type="single">
				<AccordionItem value="table-of-contents">
					<AccordionTrigger>Table of Contents</AccordionTrigger>
					<AccordionContent>
						<BlogTableOfContent chapters={data.posts.headings} />
					</AccordionContent>
				</AccordionItem>
			</Accordion>
		</div>
		<article
			class="prose dark:prose-invert hover:prose-a:text-accent-foreground prose-a:prose-headings:mb-3 prose-a:prose-headings:mt-8 prose-a:prose-headings:font-heading prose-a:prose-headings:font-bold prose-a:prose-headings:leading-tight prose-a:prose-headings:no-underline my-4 max-w-7xl lg:mr-auto lg:max-w-2xl"
		>
			<h1 class="font-heading mb-2 text-4xl">{data.posts.meta.title}</h1>
			{#if data.posts.meta.description}
				<p class="mt-0 mb-2 text-base text-neutral-700 dark:text-neutral-200">
					{data.posts.meta.description}
				</p>
			{/if}
			{#if data.posts.meta.thumbnail}
				<figure class="mb-4">
					<img src={data.posts.meta.thumbnail} alt={data.posts.meta.title} class="rounded-md" />
				</figure>
			{/if}
			{#if PostComponent}
				<PostComponent />
			{/if}
		</article>
	</div>
</div>
