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
			// disableAsyncLocalStorage: true,
			urlPatterns: [
				{
					pattern: "/",
					localized: [
						["en", "/en"],
						["es", "/es"],
						["id", "/id"],
						["ru", "/ru"],
						["pt", "/pt"],
						["fr", "/fr"],
						["de", "/de"],
						["zh", "/zh"],
						["hi", "/hi"],
						["ar", "/ar"],
						["ja", "/ja"],
						["tr", "/tr"],
						["vi", "/vi"],
						["th", "/th"],
						["el", "/el"],
						["it", "/it"],
					],
				},
				{
					pattern: "/:path(.*)?",
					localized: [
						["en", "/en/:path(.*)?"],
						["es", "/es/:path(.*)?"],
						["id", "/id/:path(.*)?"],
						["ru", "/ru/:path(.*)?"],
						["pt", "/pt/:path(.*)?"],
						["fr", "/fr/:path(.*)?"],
						["de", "/de/:path(.*)?"],
						["zh", "/zh/:path(.*)?"],
						["hi", "/hi/:path(.*)?"],
						["ar", "/ar/:path(.*)?"],
						["ja", "/ja/:path(.*)?"],
						["tr", "/tr/:path(.*)?"],
						["vi", "/vi/:path(.*)?"],
						["th", "/th/:path(.*)?"],
						["el", "/el/:path(.*)?"],
						["it", "/it/:path(.*)?"],
					],
				},
				{
					pattern: '/dashboard',
					localized: [
						["en", "/dashboard"],
						["es", "/dashboard"],
						["id", "/dashboard"],
						["ru", "/dashboard"],
						["pt", "/dashboard"],
						["fr", "/dashboard"],
						["de", "/dashboard"],
						["zh", "/dashboard"],
						["hi", "/dashboard"],
						["ar", "/dashboard"],
						["ja", "/dashboard"],
						["tr", "/dashboard"],
						["vi", "/dashboard"],
						["th", "/dashboard"],
						["el", "/dashboard"],
						["it", "/dashboard"],
					],
				},
				{
					pattern: '/cookies',
					localized: [
						["en", "/cookies"],
						["es", "/cookies"],
						["id", "/cookies"],
						["ru", "/cookies"],
						["pt", "/cookies"],
						["fr", "/cookies"],
						["de", "/cookies"],
						["zh", "/cookies"],
						["hi", "/cookies"],
						["ar", "/cookies"],
						["ja", "/cookies"],
						["tr", "/cookies"],
						["vi", "/cookies"],
						["th", "/cookies"],
						["el", "/cookies"],
						["it", "/cookies"],
					],
				},
				{
					pattern: '/download/:path(.*)?',
					localized: [
						["en", "/download/:path(.*)?"],
						["es", "/download/:path(.*)?"],
						["id", "/download/:path(.*)?"],
						["ru", "/download/:path(.*)?"],
						["pt", "/download/:path(.*)?"],
						["fr", "/download/:path(.*)?"],
						["de", "/download/:path(.*)?"],
						["zh", "/download/:path(.*)?"],
						["hi", "/download/:path(.*)?"],
						["ar", "/download/:path(.*)?"],
						["ja", "/download/:path(.*)?"],
						["tr", "/download/:path(.*)?"],
						["vi", "/download/:path(.*)?"],
						["th", "/download/:path(.*)?"],
						["el", "/download/:path(.*)?"],
						["it", "/download/:path(.*)?"],
					],
				},
				{
					pattern: '/settings/:path(.*)?',
					localized: [
						["en", "/settings/:path(.*)?"],
						["es", "/settings/:path(.*)?"],
						["id", "/settings/:path(.*)?"],
						["ru", "/settings/:path(.*)?"],
						["pt", "/settings/:path(.*)?"],
						["fr", "/settings/:path(.*)?"],
						["de", "/settings/:path(.*)?"],
						["zh", "/settings/:path(.*)?"],
						["hi", "/settings/:path(.*)?"],
						["ar", "/settings/:path(.*)?"],
						["ja", "/settings/:path(.*)?"],
						["tr", "/settings/:path(.*)?"],
						["vi", "/settings/:path(.*)?"],
						["th", "/settings/:path(.*)?"],
						["el", "/settings/:path(.*)?"],
						["it", "/settings/:path(.*)?"],
					],
				},
				{
					pattern: '/accounts/:path(.*)?',
					localized: [
						["en", "/accounts/:path(.*)?"],
						["es", "/accounts/:path(.*)?"],
						["id", "/accounts/:path(.*)?"],
						["ru", "/accounts/:path(.*)?"],
						["pt", "/accounts/:path(.*)?"],
						["fr", "/accounts/:path(.*)?"],
						["de", "/accounts/:path(.*)?"],
						["zh", "/accounts/:path(.*)?"],
						["hi", "/accounts/:path(.*)?"],
						["ar", "/accounts/:path(.*)?"],
						["ja", "/accounts/:path(.*)?"],
						["tr", "/accounts/:path(.*)?"],
						["vi", "/accounts/:path(.*)?"],
						["th", "/accounts/:path(.*)?"],
						["el", "/accounts/:path(.*)?"],
						["it", "/accounts/:path(.*)?"],
					],
				},
				{
					pattern: '/application/:path(.*)?',
					localized: [
						["en", "/application/:path(.*)?"],
						["es", "/application/:path(.*)?"],
						["id", "/application/:path(.*)?"],
						["ru", "/application/:path(.*)?"],
						["pt", "/application/:path(.*)?"],
						["fr", "/application/:path(.*)?"],
						["de", "/application/:path(.*)?"],
						["zh", "/application/:path(.*)?"],
						["hi", "/application/:path(.*)?"],
						["ar", "/application/:path(.*)?"],
						["ja", "/application/:path(.*)?"],
						["tr", "/application/:path(.*)?"],
						["vi", "/application/:path(.*)?"],
						["th", "/application/:path(.*)?"],
						["el", "/application/:path(.*)?"],
						["it", "/application/:path(.*)?"],
					],
				},
				{
					pattern: '/users/:path(.*)?',
					localized: [
						["en", "/users/:path(.*)?"],
						["es", "/users/:path(.*)?"],
						["id", "/users/:path(.*)?"],
						["ru", "/users/:path(.*)?"],
						["pt", "/users/:path(.*)?"],
						["fr", "/users/:path(.*)?"],
						["de", "/users/:path(.*)?"],
						["zh", "/users/:path(.*)?"],
						["hi", "/users/:path(.*)?"],
						["ar", "/users/:path(.*)?"],
						["ja", "/users/:path(.*)?"],
						["tr", "/users/:path(.*)?"],
						["vi", "/users/:path(.*)?"],
						["th", "/users/:path(.*)?"],
						["el", "/users/:path(.*)?"],
						["it", "/users/:path(.*)?"],
					],
				},
				{
					pattern: '/platform/:path(.*)?',
					localized: [
						["en", "/platform/:path(.*)?"],
						["es", "/platform/:path(.*)?"],
						["id", "/platform/:path(.*)?"],
						["ru", "/platform/:path(.*)?"],
						["pt", "/platform/:path(.*)?"],
						["fr", "/platform/:path(.*)?"],
						["de", "/platform/:path(.*)?"],
						["zh", "/platform/:path(.*)?"],
						["hi", "/platform/:path(.*)?"],
						["ar", "/platform/:path(.*)?"],
						["ja", "/platform/:path(.*)?"],
						["tr", "/platform/:path(.*)?"],
						["vi", "/platform/:path(.*)?"],
						["th", "/platform/:path(.*)?"],
						["el", "/platform/:path(.*)?"],
						["it", "/platform/:path(.*)?"],
					],
				},
				{
					pattern: '/subscription/:path(.*)?',
					localized: [
						["en", "/subscription/:path(.*)?"],
						["es", "/subscription/:path(.*)?"],
						["id", "/subscription/:path(.*)?"],
						["ru", "/subscription/:path(.*)?"],
						["pt", "/subscription/:path(.*)?"],
						["fr", "/subscription/:path(.*)?"],
						["de", "/subscription/:path(.*)?"],
						["zh", "/subscription/:path(.*)?"],
						["hi", "/subscription/:path(.*)?"],
						["ar", "/subscription/:path(.*)?"],
						["ja", "/subscription/:path(.*)?"],
						["tr", "/subscription/:path(.*)?"],
						["vi", "/subscription/:path(.*)?"],
						["th", "/subscription/:path(.*)?"],
						["el", "/subscription/:path(.*)?"],
						["it", "/subscription/:path(.*)?"],
					],
				},
				{
					pattern: '/transaction/:path(.*)?',
					localized: [
						["en", "/transaction/:path(.*)?"],
						["es", "/transaction/:path(.*)?"],
						["id", "/transaction/:path(.*)?"],
						["ru", "/transaction/:path(.*)?"],
						["pt", "/transaction/:path(.*)?"],
						["fr", "/transaction/:path(.*)?"],
						["de", "/transaction/:path(.*)?"],
						["zh", "/transaction/:path(.*)?"],
						["hi", "/transaction/:path(.*)?"],
						["ar", "/transaction/:path(.*)?"],
						["ja", "/transaction/:path(.*)?"],
						["tr", "/transaction/:path(.*)?"],
						["vi", "/transaction/:path(.*)?"],
						["th", "/transaction/:path(.*)?"],
						["el", "/transaction/:path(.*)?"],
						["it", "/transaction/:path(.*)?"],
					],
				},
				{
					pattern: '/server-status/:path(.*)?',
					localized: [
						["en", "/server-status/:path(.*)?"],
						["es", "/server-status/:path(.*)?"],
						["id", "/server-status/:path(.*)?"],
						["ru", "/server-status/:path(.*)?"],
						["pt", "/server-status/:path(.*)?"],
						["fr", "/server-status/:path(.*)?"],
						["de", "/server-status/:path(.*)?"],
						["zh", "/server-status/:path(.*)?"],
						["hi", "/server-status/:path(.*)?"],
						["ar", "/server-status/:path(.*)?"],
						["ja", "/server-status/:path(.*)?"],
						["tr", "/server-status/:path(.*)?"],
						["vi", "/server-status/:path(.*)?"],
						["th", "/server-status/:path(.*)?"],
						["el", "/server-status/:path(.*)?"],
						["it", "/server-status/:path(.*)?"],
					],
				}
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
