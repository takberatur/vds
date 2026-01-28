<script lang="ts">
	import type { Snippet } from 'svelte';
	import {
		ClientLayout,
		ClientMp3FeatureSection,
		ClientMp3HowToUseSection,
		ClientMp3SupportFormatSection,
		ClientMp3CTASection
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
	<ClientMp3FeatureSection />
	<ClientMp3HowToUseSection {lang} />
	<ClientMp3SupportFormatSection />
	<ClientMp3CTASection />
</ClientLayout>
