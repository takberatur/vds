<script lang="ts">
	import type { Snippet } from 'svelte';
	import {
		ClientLayout,
		ClientFeatureSection,
		ClientHowToUseSection,
		ClientSupportFormatSection,
		ClientCTASection
	} from '@/components/client/index.js';

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
	{@render children?.()}
	<ClientFeatureSection />

	<!-- How To Section -->
	<ClientHowToUseSection />

	<!-- Supported Formats -->
	<ClientSupportFormatSection />

	<!-- CTA Section -->
	<ClientCTASection />
</ClientLayout>
