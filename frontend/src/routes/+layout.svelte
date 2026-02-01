<script lang="ts">
	import './layout.css';
	import { page } from '$app/state';
	import { onMount, onDestroy, setContext, hydrate } from 'svelte';
	import { browser } from '$app/environment';
	import { afterNavigate, beforeNavigate } from '$app/navigation';
	import { ModeWatcher } from 'mode-watcher';
	import { ProgressBarIndicator, ToastContent, PageLoading } from '@/components';
	import {
		handlePageLoading,
		handlePageReloading,
		handleSubmitLoading,
		handleManualSubmission,
		forceResetProgress,
		smartNavigation,
		waitForPageReady,
		userStore,
		settingStore,
		langStore,
		createWebsocketStore
	} from '$lib/stores';
	import { locales, localizeHref, overwriteGetLocale, type Locale } from '@/paraglide/runtime';

	let { data, children } = $props();

	// svelte-ignore state_referenced_locally
	overwriteGetLocale(() => data.lang as Locale);

	const trackedMethods = ['post', 'patch', 'put', 'delete'];
	let originalFetch: typeof window.fetch;
	let activeIntervals = new Set<number>();
	let cleanupCallbacks = new Set<() => void>();
	let navigationTimeout: ReturnType<typeof setTimeout> | null = null;
	let formSubmissionTracker = new Map<HTMLFormElement, boolean>();
	let autoResetCleanup: (() => void) | null = null;
	let navigationStartTime: number = 0;
	let currentNavigationId: string | null = null;

	// svelte-ignore state_referenced_locally
	const wsStore = createWebsocketStore();

	const waitForDOMReady = (): Promise<void> => {
		return new Promise((resolve) => {
			if (document.readyState === 'complete') {
				resolve();
			} else {
				const handler = () => {
					if (document.readyState === 'complete') {
						document.removeEventListener('readystatechange', handler);
						resolve();
					}
				};
				const safelyIntercept = () => {
					document.addEventListener('readystatechange', handler);
					setTimeout(resolve, 500);
				};
				safelyIntercept();
			}
		});
	};

	const setupInterceptors = () => {
		cleanupAll();
		originalFetch = window.fetch;
		const formCleanup = interceptFormSubmissions();
		const fetchCleanup = interceptFetchRequests();

		cleanupCallbacks.add(() => {
			formCleanup();
			fetchCleanup();
		});
	};

	const cleanupAll = () => {
		if (navigationTimeout) {
			clearTimeout(navigationTimeout);
			navigationTimeout = null;
		}

		activeIntervals.forEach((id) => clearInterval(id));
		cleanupCallbacks.forEach((fn) => fn());
		activeIntervals.clear();
		cleanupCallbacks.clear();
		formSubmissionTracker.clear();
		forceResetProgress();
	};

	function interceptFormSubmissions(): () => void {
		const submitHandler = (e: Event) => {
			const form = e.target as HTMLFormElement;
			if (trackedMethods.includes(form.method.toLowerCase())) {
				if (formSubmissionTracker.get(form)) {
					return;
				}

				formSubmissionTracker.set(form, true);
				handleSubmitLoading(true);

				const timeoutId = setTimeout(() => {
					formSubmissionTracker.delete(form);
					handleSubmitLoading(false);
				}, 10000);

				const completeSubmission = () => {
					clearTimeout(timeoutId);
					formSubmissionTracker.delete(form);
					handleSubmitLoading(false);
				};

				form.addEventListener('formdata', completeSubmission, { once: true });
				form.addEventListener('reset', completeSubmission, { once: true });
			}
		};

		document.addEventListener('submit', submitHandler);
		return () => {
			document.removeEventListener('submit', submitHandler);
			formSubmissionTracker.clear();
		};
	}

	function interceptFetchRequests(): () => void {
		const original = window.fetch;
		const pendingRequests = new Set<string>();

		window.fetch = async (input, init) => {
			const method = init?.method?.toLowerCase();
			let url: string;
			if (typeof input === 'string') {
				url = input;
			} else if (input instanceof URL) {
				url = input.href;
			} else if (input instanceof Request) {
				url = input.url;
			} else {
				url = 'unknown';
			}

			if (method && trackedMethods.includes(method)) {
				const requestKey = `${method}:${url}`;
				if (pendingRequests.has(requestKey)) {
					return original(input, init);
				}

				pendingRequests.add(requestKey);
				handleManualSubmission(true);

				try {
					const response = await original(input, init);
					return response;
				} finally {
					pendingRequests.delete(requestKey);

					setTimeout(() => {
						if (pendingRequests.size === 0) {
							handleManualSubmission(false);
						}
					}, 200);
				}
			}

			return original(input, init);
		};

		return () => {
			window.fetch = original;
			pendingRequests.clear();
		};
	}

	$effect(() => {
		if (data.user) {
			userStore.set(data.user);
		}
		if (data.settings) {
			settingStore.set(data.settings);
		}
		if (data.lang) {
			langStore.set(data.lang);
		}
	});

	onMount(() => {
		wsStore.connect();
		if (!browser) return;

		const initializeApp = async () => {
			try {
				await waitForDOMReady();

				setupInterceptors();
			} catch (error) {
				console.error('❌ App initialization failed:', error);
			}
		};

		initializeApp();

		return () => {
			cleanupAll();
			if (autoResetCleanup) {
				(autoResetCleanup as () => void)();
			}
		};
	});

	onDestroy(() => {
		cleanupAll();

		if (autoResetCleanup) {
			(autoResetCleanup as () => void)();
		}
		smartNavigation.cleanup();

		wsStore.disconnect();
	});
	beforeNavigate(({ from, to, type }) => {
		if (navigationTimeout) {
			clearTimeout(navigationTimeout);
			navigationTimeout = null;
		}

		if (!from || !to) {
			// Full page reload
			handlePageReloading(true);
		} else if (from.url.pathname !== to.url.pathname) {
			// Route navigation
			navigationStartTime = Date.now();
			currentNavigationId = `nav-${navigationStartTime}`;

			smartNavigation.setActiveNavigation(currentNavigationId);
			handlePageLoading(true);

			let timeoutDuration = 8000;

			const targetRoute = to.url.pathname;
			if (targetRoute.includes('/admin')) {
				timeoutDuration = 10000;
			} else if (type === 'popstate') {
				timeoutDuration = 3000;
			} else if (type === 'link') {
				timeoutDuration = 6000;
			}

			navigationTimeout = setTimeout(() => {
				if (currentNavigationId && smartNavigation.isNavigationActive(currentNavigationId)) {
					smartNavigation.clearActiveNavigation(currentNavigationId);
					handlePageLoading(false);
					handlePageReloading(false);
				}
				navigationTimeout = null;
			}, timeoutDuration);
		}
	});
	afterNavigate(({ from, to, type }) => {
		if (navigationTimeout) {
			clearTimeout(navigationTimeout);
			navigationTimeout = null;
		}

		const navId = currentNavigationId;

		if (!navId || !smartNavigation.isNavigationActive(navId)) {
			return;
		}

		waitForPageReady({
			maxWaitTime: smartNavigation.isNavigationSlow(navigationStartTime) ? 1000 : 2000,
			selectors: ['main', '[data-sveltekit-loaded]', '.content', '#app > *']
		})
			.then(() => {
				if (navId && smartNavigation.isNavigationActive(navId)) {
					smartNavigation.clearActiveNavigation(navId);

					setTimeout(() => {
						handlePageLoading(false);
						handlePageReloading(false);
					});
				}
			})
			.catch((error) => {
				if (navId) {
					smartNavigation.clearActiveNavigation(navId);
				}
				handlePageLoading(false);
				handlePageReloading(false);
			});
	});

	setContext('websocket', wsStore);
</script>

<ModeWatcher defaultMode="system" />
<ProgressBarIndicator />
<PageLoading />
<ToastContent />

<main class="font-roboto bg-background text-foreground antialiased">
	{#if children}
		{@render children()}
	{/if}
</main>
<div style="display:none">
	{#each locales as locale}
		<a href={localizeHref(page.url.pathname, { locale })}>{locale}</a>
	{/each}
</div>

<style>
	:global(body) {
		overflow-x: hidden;
	}
</style>
