import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import { mdsvex, escapeSvelte } from 'mdsvex';
import { createHighlighter } from 'shiki';
import remarkMath from 'remark-math';
import rehypeKatex from 'rehype-katex';
import rehypeSlug from 'rehype-slug';
import remarkToc from 'remark-toc';

const theme = 'github-dark';
const highlighter = await createHighlighter({
	themes: [theme],
	langs: ['javascript', 'typescript', 'ts', 'js', 'html', 'css', 'bash']
});

/** @type {import('mdsvex').MdsvexOptions} */
const mdsvexOptions = {
	extensions: ['.svx', '.md'],
	layout: {
		blog: './src/lib/components/blog/BlogLayout.svelte'
	},
	remarkPlugins: [remarkMath, rehypeSlug, [remarkToc, { heading: 'toc' }]],
	rehypePlugins: [rehypeKatex],
	highlight: {
		highlighter: async (code, lang = 'text') => {
			const shiki = await getShiki();
			const html = escapeSvelte(shiki.codeToHtml(code, { lang, theme }));
			return html;
		}
	}
};

async function getShiki() {
	return highlighter;
}

/** @type {import('@sveltejs/kit').Config} */
const config = {
	extensions: ['.svelte', '.svx', '.md'],
	preprocess: [mdsvex(mdsvexOptions), vitePreprocess()],
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
