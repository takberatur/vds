<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ClientFooter, ClientHeader } from '@/components/client/index.js';

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
	<ClientHeader {user} {setting} />
	{@render children?.()}
	<ClientFooter {setting} {platforms} {lang} />
</div>
