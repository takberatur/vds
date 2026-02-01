<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ClientLayout } from '@/components/client/index.js';

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

<ClientLayout {user} {setting} {platforms} {lang}>
	{#if children}
		{@render children()}
	{/if}
</ClientLayout>
