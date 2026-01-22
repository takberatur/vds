<script lang="ts">
	import { MetaTags, type MetaTagsProps } from 'svelte-meta-tags';
	import { ClientHomeLayout, ClientHeroSection } from '@/components/client/index.js';

	let { data } = $props();
	let metaTags = $derived<MetaTagsProps | undefined>(data.pageMetaTags);

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
