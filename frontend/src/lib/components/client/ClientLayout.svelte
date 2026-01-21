<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ClientFooter, ClientHeader } from '@/components/client/index.js';
	import { type Locale } from '@/paraglide/runtime';

	let {
		user,
		setting,
		platforms,
		lang,
		children
	}: {
		user?: User | null;
		setting?: SettingsValue | null;
		platforms?: Platform[];
		lang?: string;
		children?: Snippet;
	} = $props();
</script>

<div class="min-h-screen">
	<ClientHeader {user} {setting} lang={lang as Locale} />
	{@render children?.()}
	<ClientFooter {setting} {platforms} lang={lang as Locale} />
</div>
