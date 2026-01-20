import { json } from "@sveltejs/kit";

export const POST = async ({ locals }) => {
	await locals.deps.userService.logout();

	locals.deps.authHelper.clearAuthCookies();
	locals.user = null;

	return json({ success: true, message: 'Logged out successfully' });
};
