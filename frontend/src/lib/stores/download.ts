import { browser } from '$app/environment';
import { PUBLIC_CENTRIFUGE_URL } from '$env/static/public';
import { writable } from 'svelte/store';
import { Centrifuge, Subscription } from 'centrifuge';

const CENTRIFUGO_URL = PUBLIC_CENTRIFUGE_URL;

export type DownloadFormat = {
	format_id?: string;
	url?: string;
	ext?: string;
	resolution?: string;
	filesize?: number | null;
	height?: number | null;
	vcodec?: string;
	acodec?: string;
	tbr?: number | null;
};

export type DownloadTaskView = {
	id: string;
	status: string;
	progress: number;
	title?: string | null;
	thumbnail_url?: string | null;
	type: string;
	created_at?: string | null;
	file_path?: string | null;
	formats?: DownloadFormat[] | null;
};

export type DownloadState = {
	tasks: Record<string, DownloadTaskView>;
};

export type WebsocketStore = ReturnType<typeof createWebsocketStore>;


export const createWebsocketStore = (userID?: string | null) => {
	let centrifuge: Centrifuge | null = null;
	let subscriptions: Record<string, Subscription> = {};

	const state = writable<DownloadState>({
		tasks: {}
	});

	function applyEvent(data: any) {
		if (!data || !data.task_id) return;

		const id = String(data.task_id);
		const progress =
			typeof data.progress === 'number'
				? data.progress
				: data.status === 'completed'
					? 100
					: data.status === 'failed'
						? 0
						: 0;

		state.update((current) => {
			const existing = current.tasks[id];
			const payload = data.payload ?? {};

			return {
				tasks: {
					...current.tasks,
					[id]: {
						...existing,
						id,
						status: data.status ?? existing?.status ?? 'processing',
						progress,
						created_at: data.created_at ?? existing?.created_at ?? null,
						file_path: payload.file_path ?? existing?.file_path ?? null,
						formats: payload.formats ?? existing?.formats ?? null
					}
				}
			};
		});
	}

	function connect() {
		if (!browser) return;

		if (centrifuge) return;

		// Use environment variable if available (need to be added to .env) or fallback
		// For this environment, user specified ws://infrastructure-centrifugo:8000
		// But browser needs localhost or public URL.
		const url = CENTRIFUGO_URL;

		centrifuge = new Centrifuge(url, {
			// Add any auth token if needed, but for public downloads it might be anonymous
		});

		centrifuge.on('connected', (ctx) => {
			console.log('Centrifugo connected', ctx);
		});

		centrifuge.on('disconnected', (ctx) => {
			console.log('Centrifugo disconnected', ctx);
		});

		centrifuge.connect();
	}

	function subscribe(taskId: string) {
		if (!browser || !centrifuge) return;
		if (subscriptions[taskId]) return;

		const channel = `download:progress:${taskId}`;
		console.log(`Subscribing to channel: ${channel}`);

		const sub = centrifuge.newSubscription(channel);

		sub.on('publication', (ctx) => {
			// ctx.data is the payload published from backend
			applyEvent(ctx.data);
		});

		sub.on('subscribing', (ctx) => {
			console.log(`Subscribing to ${channel}`, ctx);
		});

		sub.on('subscribed', (ctx) => {
			console.log(`Subscribed to ${channel}`, ctx);
		});

		sub.on('error', (ctx) => {
			console.error(`Subscription error for ${channel}`, ctx);
		});

		sub.subscribe();
		subscriptions[taskId] = sub;
	}

	function upsertTaskFromApi(task: any) {
		if (!task || !task.id) return;

		const id = String(task.id);

		state.update((current) => {
			const existing = current.tasks[id];

			return {
				tasks: {
					...current.tasks,
					[id]: {
						id,
						status: task.status ?? existing?.status ?? 'queued',
						progress: existing?.progress ?? 0,
						title: task.title ?? existing?.title ?? null,
						thumbnail_url: task.thumbnail_url ?? existing?.thumbnail_url ?? null,
						type: task.type ?? existing?.type ?? null,
						created_at: task.created_at ?? existing?.created_at ?? null,
						file_path: task.file_path ?? existing?.file_path ?? null,
						formats: task.formats ?? existing?.formats ?? null
					}
				}
			};
		});
	}

	function disconnect() {
		if (centrifuge) {
			centrifuge.disconnect();
			centrifuge = null;
			subscriptions = {};
		}
		state.set({ tasks: {} });
	}

	return {
		state,
		stateSubscribe: state.subscribe,
		connect,
		upsertTaskFromApi,
		disconnect,
		subscribe // Expose this
	};
};
