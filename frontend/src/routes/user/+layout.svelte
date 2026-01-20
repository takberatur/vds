<script lang="ts">
	import { ClientUserLayout } from '$lib/components/client/index.js';

	let { data, children } = $props();
</script>

<ClientUserLayout setting={data.settings}>
	{@render children()}
</ClientUserLayout>
