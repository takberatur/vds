<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { format, parseISO } from 'date-fns';
	import { defaultAuthor } from './data.js';

	let {
		params,
		alt = `Article by ${defaultAuthor.name}`
	}: {
		params: { slug: string };
		alt: string;
	} = $props();

	let post = $state<any | null>(null);
	let imageUrl = $state<string | null>(null);
	let error = $state<string | null>(null);

	let date = $derived(post ? post.lastUpdatedDate || post.publishedDate : null);
	let formattedDate = $derived(date ? format(parseISO(date), 'LLLL d, yyyy') : '');
	let readTime = $derived(post ? `${post.readTimeMinutes} min read` : '');

	const size = {
		width: 1200,
		height: 630
	};

	const contentType = 'image/png';

	onMount(async () => {
		try {
			await fetchPostData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load post';
			console.error('Error loading post:', err);
		}
	});

	$effect(() => {
		if (post && browser) {
			generateImage();
		}
	});

	async function fetchPostData() {
		// Implementasi fetching data sesuai dengan struktur data Anda
		// Contoh dengan contentlayer (perlu disesuaikan):
		// import { allPosts } from 'src/lib/contentlayer';
		// post = allPosts.find((p: any) => p.slug === params.slug);

		// Contoh dengan API route:
		// const response = await fetch(`/api/posts/${params.slug}`);
		// if (response.ok) {
		//   post = await response.json();
		// } else {
		//   throw new Error('Post not found');
		// }

		// Simulasi data:
		post = {
			slug: params.slug,
			title: 'Sample Post Title',
			publishedDate: new Date().toISOString(),
			lastUpdatedDate: null,
			readTimeMinutes: 5
		};
	}

	function generateImage() {
		if (!browser || !post) return;

		// Membuat canvas untuk generate image
		const canvas = document.createElement('canvas');
		canvas.width = size.width;
		canvas.height = size.height;
		const ctx = canvas.getContext('2d');

		if (!ctx) return;

		// Background gradient
		const gradient = ctx.createLinearGradient(0, 0, size.width, size.height);
		gradient.addColorStop(0, 'rgba(59, 178, 93, 0.20)');
		gradient.addColorStop(1, 'rgba(59, 121, 178, 0.20)');

		ctx.fillStyle = gradient;
		ctx.fillRect(0, 0, size.width, size.height);

		// Text styling
		ctx.fillStyle = '#222';
		ctx.font = '400 24px system-ui';
		ctx.fillText(defaultAuthor.handle, 48, 64);

		ctx.font = 'bold 48px system-ui';
		ctx.fillText(post.title, 48, 200, size.width - 96);

		ctx.font = '20px system-ui';
		ctx.fillText(`${formattedDate} • ${readTime}`, 48, 300);

		// Convert canvas to data URL
		imageUrl = canvas.toDataURL('image/png');
	}
</script>

<!-- Image Display Component -->
<div class="image-container">
	{#if error}
		<div class="error">
			{error}
		</div>
	{:else if imageUrl}
		<img
			src={imageUrl}
			{alt}
			width={size.width}
			height={size.height}
			style="max-width: 100%; height: auto;"
		/>
	{:else if post}
		<!-- Fallback visual representation -->
		<div
			class="image-preview"
			style="
        width: {size.width}px;
        height: {size.height}px;
        max-width: 100%;
        background: linear-gradient(45deg, rgba(59, 178, 93, 0.20) 0%, rgba(59, 121, 178, 0.20) 100%);
        display: flex;
        align-items: flex-start;
        flex-direction: column;
        justify-content: space-between;
        letter-spacing: -0.02em;
        padding: 64px 48px;
        color: #222;
      "
		>
			<div style="display: flex;">
				<span style="font-size: 24px; font-weight: 400;">
					{defaultAuthor.handle}
				</span>
			</div>
			<div
				style="
          display: flex;
          flex-direction: column;
          align-items: flex-start;
          width: auto;
          max-width: 70%;
        "
			>
				<p
					style="
            font-weight: bold;
            font-size: 48px;
            line-height: 1.1;
            margin: 0;
          "
				>
					{post.title}
				</p>
				<p style="font-size: 20px; margin: 20px 0 0 0;">
					{formattedDate} &middot; {readTime}
				</p>
			</div>
		</div>
	{:else}
		<div class="loading">Loading image...</div>
	{/if}
</div>

<style>
	.image-container {
		display: flex;
		justify-content: center;
		align-items: center;
		margin: 0 auto;
	}

	.image-preview {
		border-radius: 8px;
		overflow: hidden;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
	}

	.error {
		color: #dc2626;
		padding: 20px;
		border: 1px solid #fecaca;
		border-radius: 4px;
		background-color: #fef2f2;
	}

	.loading {
		padding: 40px;
		text-align: center;
		color: #666;
	}

	@media (max-width: 1200px) {
		.image-preview {
			transform: scale(0.8);
			transform-origin: top center;
		}
	}
</style>
