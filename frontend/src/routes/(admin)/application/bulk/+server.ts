import { json } from "@sveltejs/kit";

export const DELETE = async ({ request, locals }) => {
	try {
		const { ids } = await request.json();
		if (!ids || !Array.isArray(ids) || ids.length === 0) {
			throw new Error('IDs are required');
		}
		const response = await locals.deps.applicationService.bulkDelete(ids);
		if (response instanceof Error) {
			throw new Error(response.message || 'Failed to delete applications');
		}
		return json({ success: true, message: response }, { status: 200 });
	} catch (error) {
		console.error('Error deleting applications:', error);
		return json({ success: false, message: error instanceof Error ? error.message : 'Unknown error' }, { status: 500 });
	}
}
