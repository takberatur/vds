import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			out: 'build',
			precompress: true,
			dynamic_origin: true
		}),
		csrf: {
			trustedOrigins:
				process.env.NODE_ENV === 'production'
					? ['https://simontokz.com', 'https://www.simontokz.com']
					: ['*']
		},
		alias: {
			'@': './src/lib',
			'@/*': './src/lib/*'
		}
	}
};

export default config;
