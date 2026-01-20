import { json } from '@sveltejs/kit';

export const DELETE = async ({ locals, request }) => {
	try {
		const { ids } = await request.json();
		if (!ids || ids.length === 0) {
			return json({
				success: false,
				message: 'Download ID is required'
			}, { status: 400 });
		}
		await locals.deps.downloadService.BulkDelete(ids);
		return json({
			success: true,
			message: `${ids.length} download(s) deleted successfully`
		}, { status: 200 });
	} catch (error) {
		return json({
			success: false,
			message: error instanceof Error ? error.message : 'Failed to delete download'
		}, { status: 500 });
	}
}
