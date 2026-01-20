<script lang="ts">
	import { MetaTags } from 'svelte-meta-tags';
	import { ClientHomeLayout, ClientHeroSection } from '@/components/client/index.js';

	let { data } = $props();
	let metaTags = $derived(data.pageMetaTags);
</script>

<MetaTags {...metaTags} />

<ClientHomeLayout
	user={data.user}
	setting={data.settings}
	platforms={data.platforms}
	lang={data.lang}
>
	<ClientHeroSection
		id="hero"
		user={data.user}
		platforms={data.platforms}
		setting={data.settings}
		form={data.form}
	/>
</ClientHomeLayout>
