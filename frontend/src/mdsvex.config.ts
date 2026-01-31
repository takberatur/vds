import { getShiki } from './lib/content/shiki';

export default {
	extensions: ['.svx', '.md'],
	layout: {
		_: './routes/blog/layout.svelte'
	},
	highlight: {
		//@ts-expect-error
		highlighter: async (code, lang = 'text') => {
			const shiki = await getShiki();
			return shiki.codeToHtml(code, {
				lang,
				theme: 'github-dark'
			});
		}
	}
};
