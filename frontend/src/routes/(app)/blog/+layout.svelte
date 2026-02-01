<script lang="ts">
	import { ClientFooter, ClientHeader } from '@/components/client/index.js';
	import { type Locale } from '@/paraglide/runtime';

	let { data, children } = $props();
</script>

<svelte:head>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0.16.9/dist/katex.min.css" />
</svelte:head>
<div class="min-h-screen">
	<ClientHeader user={data.user} setting={data.settings} lang={data.lang as Locale} />
	{@render children()}
	<ClientFooter platforms={data.platforms} setting={data.settings} lang={data.lang as Locale} />
</div>
