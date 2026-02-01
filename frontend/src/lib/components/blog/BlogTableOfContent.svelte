<script lang="ts" module>
	interface PostHeading {
		depth: number;
		value: string;
		slug: string;
	}
	interface TocProps {
		chapters: PostHeading[];
	}
</script>

<script lang="ts">
	import { cn } from '$lib/utils';
	import { localizeHref } from '@/paraglide/runtime';

	let { chapters }: TocProps = $props();

	let activeSlug = $state<string>('');

	const activeHeading = $derived(chapters.find((heading) => heading.slug === activeSlug));

	$effect(() => {
		if (typeof window === 'undefined') return;

		const observer = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					if (entry?.isIntersecting) {
						activeSlug = entry.target.id;
					}
				});
			},
			{
				rootMargin: '-30% 0px'
			}
		);

		chapters.forEach((chapter) => {
			const element = document.getElementById(chapter.slug);
			if (element) {
				observer.observe(element);
			}
		});

		return () => observer.disconnect();
	});

	function getHeadingClass(heading: PostHeading) {
		const baseClasses = cn(
			'list-none text-sm font-bold transition-colors duration-200 ease-in-out hover:text-accent-foreground',
			activeSlug === heading.slug && 'text-accent-foreground'
		);

		if (heading.depth === 3) {
			return cn(baseClasses, 'ml-6 font-normal');
		} else if (heading.depth === 4) {
			return cn(baseClasses, 'ml-8 font-normal');
		} else if (heading.depth === 5) {
			return cn(baseClasses, 'ml-10 font-normal');
		}

		return baseClasses;
	}

	function scrollToHeading(slug: string, event: Event) {
		event.preventDefault();
		const element = document.getElementById(slug);
		if (element) {
			element.scrollIntoView({ behavior: 'smooth' });
			activeSlug = slug;

			history.pushState(null, '', `#${slug}`);
		}
	}
</script>

<nav class="flex items-center self-start" aria-label="Table of Contents">
	<ol class="list-none space-y-3">
		{#each chapters as heading (heading.slug)}
			<li
				class={getHeadingClass(heading)}
				aria-current={activeSlug === heading.slug ? 'location' : undefined}
			>
				<a
					href={localizeHref(`#${heading.slug}`)}
					onclick={(e) => {
						e.preventDefault();
						scrollToHeading(heading.slug, e);
					}}
					class="cursor-pointer focus:outline-none focus-visible:rounded-sm focus-visible:ring-2 focus-visible:ring-accent-foreground"
				>
					{heading.value}
				</a>
			</li>
		{/each}
	</ol>
</nav>

<style scoped>
	a {
		text-decoration: none;
		color: inherit;
	}

	a:hover {
		text-decoration: underline;
	}

	[aria-current='location'] {
		position: relative;
	}

	[aria-current='location']::before {
		content: '→';
		position: absolute;
		left: -1.5rem;
		top: 50%;
		transform: translateY(-50%);
	}
</style>
