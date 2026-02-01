<script lang="ts">
	import type { PostSchema } from '@/utils/schema.js';
	import { Badge } from '@/components/ui/badge';
	import { CardSpotlight } from '@/components';
	import { format, parseISO } from 'date-fns';
	import { CalendarDays, Timer } from '@lucide/svelte';
	import { localizeHref } from '@/paraglide/runtime.js';
	import * as i18n from '@/paraglide/messages.js';

	let {
		post
	}: {
		post?: PostSchema;
	} = $props();
</script>

<CardSpotlight
	variant="neutral"
	shadow="medium"
	spotlightIntensity="medium"
	spotlight
	useBorder
	class="p-0"
>
	<article class="w-full">
		<a
			href={localizeHref(`/blog/${post?.slug}`)}
			class="select-rounded-md group flex w-full items-center gap-2 rounded-md border border-border p-4 leading-none no-underline transition-all outline-none hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
		>
			<div class="flex h-20 max-w-max items-center justify-center">
				<img
					src={post?.thumbnail || ''}
					alt={post?.title || ''}
					class="h-20 w-auto rounded-md object-cover transition-all group-hover:scale-110"
				/>
			</div>
			<div class="w-full space-y-2">
				<h3 class="my-2 line-clamp-2 text-2xl font-bold text-foreground">{post?.title}</h3>
				<div class="flex gap-2 text-sm leading-snug text-muted-foreground">
					<div class="flex items-center gap-1">
						<CalendarDays size={16} />
						<time dateTime={post?.publishedDate}>
							{format(parseISO(post?.publishedDate || ''), 'LLLL d, yyyy')}</time
						>
					</div>
					<span class="opacity-50">|</span>
					<div class="flex items-center gap-1">
						<Timer size={16} />
						<!-- <span>
							{post?.readingTime || ''}
							{post?.words || 0} {i18n.words()}
						</span> -->
					</div>
				</div>
				<ul class="my-4 flex list-none flex-wrap gap-2 p-0">
					{#each post?.tags || [] as tag}
						<li>
							<Badge
								variant="outline"
								class="inline-block rounded-full border border-muted-foreground/50 bg-muted-foreground/10 px-2 py-0.5 text-xs text-muted-foreground"
							>
								{tag}
							</Badge>
						</li>
					{/each}
				</ul>
			</div>
		</a>
	</article>
</CardSpotlight>
