import { paraglideVitePlugin } from '@inlang/paraglide-js'
import devtoolsJson from 'vite-plugin-devtools-json';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vitest/config';
import { playwright } from '@vitest/browser-playwright';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
	logLevel: 'info',
	build: {
		minify: true
	},
	server: {
		allowedHosts: [
			'client.giuadiario.info',
			'compositely-sanguinolent-cari.ngrok-free.dev',
			'simontokz.com'
		],
		// hmr: {
		// 	protocol: process.env.NODE_ENV === "development" ? 'wss' : undefined,
		// 	port: process.env.NODE_ENV === "development" ? 5173 : undefined,
		// 	host:
		// 		process.env.NODE_ENV === "development"
		// 			? process.env.ORIGIN
		// 			: undefined,
		// }
	},
	plugins: [
		tailwindcss(),
		sveltekit(),
		paraglideVitePlugin({
			project: './project.inlang',
			outdir: './src/lib/paraglide',
			strategy: ["url", "cookie"],
			disableAsyncLocalStorage: true,
			urlPatterns: [
				{
					pattern: "/",
					localized: [
						["en", "/en"],
						["es", "/es"],
						["id", "/id"]
					],
				},
				{
					pattern: "/:path(.*)?",
					localized: [
						["en", "/en/:path(.*)?"],
						["es", "/es/:path(.*)?"],
						["id", "/id/:path(.*)?"],
					],
				},
			]
		}),
		devtoolsJson()
	],
	test: {
		expect: { requireAssertions: true },
		projects: [
			{
				extends: './vite.config.ts',

				test: {
					name: 'client',

					browser: {
						enabled: true,
						provider: playwright(),
						instances: [{ browser: 'chromium', headless: true }]
					},

					include: ['src/**/*.svelte.{test,spec}.{js,ts}'],
					exclude: ['src/lib/server/**']
				}
			},

			{
				extends: './vite.config.ts',

				test: {
					name: 'server',
					environment: 'node',
					include: ['src/**/*.{test,spec}.{js,ts}'],
					exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	},
	ssr: {
		noExternal: ['svelte-motion']
	},
	optimizeDeps: {
		include: ['svelte', 'svelte/internal']
	},
});
