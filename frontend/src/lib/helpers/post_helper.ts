import type { SvelteComponent } from 'svelte'
import type { RequestEvent } from '@sveltejs/kit';
import { BaseHelper } from './base_helper';
import { PostSchema } from '@/utils/schema.js';


export class PostHelper extends BaseHelper implements PostHelper {
	constructor(event: RequestEvent) {
		super(event);
	}

	async getAllPost(): Promise<{ meta: PostSchema; path: string }[]> {
		const allPostFiles = import.meta.glob('/src/routes/content/*.md');
		const postFilesArr = Object.entries(allPostFiles);

		const posts = await Promise.all(
			postFilesArr.map(async ([path, resolve]) => {
				const post = (await resolve()) as PostSchema;

				return {
					meta: post,
					path: path.slice(11, -3)
				};
			})
		);

		posts.sort((a, b) => new Date(b.meta.publishedDate).getTime() - new Date(a.meta.publishedDate).getTime());
		return posts.filter((post) => post.meta.status === 'published');
	}

	async getPostBySlug(slug: string): Promise<{ component: SvelteComponent<Record<string, any>, any, any>; meta: PostSchema } | Error> {
		try {
			const post = await import(`../../routes/content/${slug}.md`);
			const parsed = PostSchema.safeParse(post.metadata);
			if (!parsed.success) {
				console.error(parsed.error.format());
				throw new Error('Invalid post metadata');
			}
			return {
				component: post.default,
				meta: parsed.data,
			};
		} catch (error) {
			console.error(error);
			throw new Error(error instanceof Error ? error.message : 'Unknown error');
		}
	}
}
