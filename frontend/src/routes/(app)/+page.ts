export const load = async ({ data }) => {
	const { user, settings, pageMetaTags, platforms, lang, form } = data;

	return {
		user,
		settings,
		pageMetaTags,
		platforms,
		lang,
		form
	};
};
