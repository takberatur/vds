export const load = async ({ data }) => {
	const { user, settings, lang } = data;

	return {
		user,
		settings,
		lang
	};
};
