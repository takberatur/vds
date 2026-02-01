export const load = async ({ data }) => {
	const { user, settings, lang, posts, pageMetaTags } = data

	return {
		posts,
		user,
		settings,
		lang,
		pageMetaTags
	};
}
