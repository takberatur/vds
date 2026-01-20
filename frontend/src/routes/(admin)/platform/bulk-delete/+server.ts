import { json } from "@sveltejs/kit";

export const DELETE = async ({ request, locals, params }) => {
	try {
		const { ids } = await request.json();
		if (!ids || ids.length === 0) {
			return json({ success: false, message: 'Platform ids are required' }, { status: 400 });
		}

		const response = await locals.deps.platformService.BulkDeletePlatforms(ids);
		if (response instanceof Error) {
			throw response;
		}
		return json({ success: true, message: 'Platform deleted successfully' }, { status: 200 });
	} catch (error) {
		return json({ success: false, message: error instanceof Error ? error.message : 'Failed to delete platform' }, { status: 500 });
	}
}
