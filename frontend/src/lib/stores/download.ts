import { browser } from '$app/environment';
import { PUBLIC_API_URL } from '$env/static/public';
import { derived, writable } from 'svelte/store';

export type DownloadTaskView = {
	id: string;
	status: string;
	progress: number;
	title?: string | null;
	thumbnail_url?: string | null;
	created_at?: string | null;
	file_path?: string | null;
	formats?: DownloadFormat[] | null;
};

export type DownloadState = {
	tasks: Record<string, DownloadTaskView>;
};

export type WebsocketStore = ReturnType<typeof createWebsocketStore>;


export const createWebsocketStore = (userID?: string | null) => {
	let socket: WebSocket | null = null;
	let currentUserId: string | null = null;

	const state = writable<DownloadState>({
		tasks: {}
	});

	const apiWsConfig = (() => {
		try {
			const url = new URL(PUBLIC_API_URL);
			const protocol = url.protocol === 'https:' ? 'wss' : 'ws';
			const basePath = url.pathname.endsWith('/') ? url.pathname.slice(0, -1) : url.pathname;

			return {
				protocol,
				host: url.host,
				basePath
			};
		} catch {
			return null as {
				protocol: string;
				host: string;
				basePath: string;
			} | null;
		}
	})();

	function getWebSocketUrl() {
		if (!browser) return null;

		if (apiWsConfig) {
			const { protocol, host, basePath } = apiWsConfig;

			if (userID) {
				return `${protocol}://${host}${basePath}/downloads/ws/${userID}`;
			}

			return `${protocol}://${host}${basePath}/ws`;
		}

		const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
		const base = `${protocol}://${window.location.host}`;

		if (userID) {
			return `${base}/api/v1/downloads/ws/${userID}`;
		}

		return `${base}/api/v1/ws`;
	}

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

		const targetId = userID ?? null;

		if (
			socket &&
			(socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) &&
			currentUserId === targetId
		) {
			return;
		}

		if (socket) {
			socket.close();
			socket = null;
		}

		const url = getWebSocketUrl();
		if (!url) return;

		currentUserId = targetId;
		socket = new WebSocket(url);

		socket.onopen = () => {
			// console.info('Download websocket connected', { url, userId: currentUserId });
		};

		socket.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				applyEvent(data);
			} catch (error) {
				console.error('Failed to parse download event', error);
			}
		};

		socket.onerror = (event) => {
			console.error('Download websocket error', event);
			if (socket) {
				socket.close();
			}
		};

		socket.onclose = () => {
			const reconnectUserId = currentUserId;
			socket = null;
			if (reconnectUserId === targetId) {
				setTimeout(() => {
					if (!socket && reconnectUserId === currentUserId) {
						connect();
					}
				}, 1000);
			}
		};
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
						created_at: task.created_at ?? existing?.created_at ?? null,
						file_path: task.file_path ?? existing?.file_path ?? null,
						formats: task.formats ?? existing?.formats ?? null
					}
				}
			};
		});
	}

	function disconnect() {
		if (socket) {
			socket.close();
			socket = null;
		}
		state.set({ tasks: {} });
	}

	return {
		state,
		stateSubscribe: state.subscribe,
		connect,
		upsertTaskFromApi,
		disconnect
	};
}
