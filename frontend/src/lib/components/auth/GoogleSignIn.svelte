<script lang="ts">
	import { onMount } from 'svelte';
	import { env } from '$env/dynamic/public';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';

	let {
		disabled = false,
		onSuccess,
		onError
	}: {
		disabled: boolean;
		onSuccess: (response: google.accounts.id.CredentialResponse) => void;
		onError: () => void;
	} = $props();

	let container = $state<HTMLElement | null>(null);

	onMount(() => {
		if (typeof window === 'undefined') return;

		const initializeGoogle = () => {
			if (!container) return;

			if (!window.google) return;

			window.google.accounts.id.initialize({
				client_id: env.PUBLIC_GOOGLE_CLIENT_ID,
				callback: onSuccess,
				auto_select: false,
				cancel_on_tap_outside: true
			});

			window.google.accounts.id.renderButton(container, {
				type: 'standard',
				theme: 'outline',
				size: 'large',
				text: 'continue_with',
				shape: 'rectangular',
				logo_alignment: 'left',
				width: 250
			});
		};

		// Check if script is already loaded
		if (document.getElementById('google-client-script')) {
			initializeGoogle();
			return;
		}

		// Load Google Script
		const script = document.createElement('script');
		script.src = 'https://accounts.google.com/gsi/client';
		script.id = 'google-client-script';
		script.async = true;
		script.defer = true;
		script.onload = initializeGoogle;
		script.onerror = onError;
		document.body.appendChild(script);
	});
</script>

<button bind:this={container} type="button" class="flex items-center justify-center" disabled={disabled}>
	<!-- <div bind:this={container} class="w-full rounded-md"></div> -->
	 Continue with Google
</button>
