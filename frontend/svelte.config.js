import adapter from 'svelte-adapter-bun';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			out: 'build',
			precompress: true,
			dynamic_origin: true,
			split: false
		}),
		csrf: {
			// checkOrigin: process.env.NODE_ENV === 'production', deprecated
			trustedOrigins:
				process.env.NODE_ENV === 'production'
					? [process.env.ORIGIN ?? 'https://simontokz.com']
					: ['*']
		},
		alias: {
			'@/*': './src/lib/*'
		}
	}
};

export default config;
