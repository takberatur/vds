import { json } from "@sveltejs/kit";

export const DELETE = async ({ params, locals }) => {
	try {
		const { id } = params;
		if (!id) {
			throw new Error('Application ID is required');
		}
		const response = await locals.deps.applicationService.delete(id);
		if (response instanceof Error) {
			throw new Error(response.message || 'Failed to delete application');
		}
		return json({ success: true, message: response }, { status: 200 });
	} catch (error) {
		console.error('Error deleting application:', error);
		return json({ success: false, message: error instanceof Error ? error.message : 'Unknown error' }, { status: 500 });
	}
}
