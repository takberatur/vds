import { json } from '@sveltejs/kit';

export const DELETE = async ({ locals, params }) => {
	try {
		const { id } = params;
		if (!id) {
			return json({
				success: false,
				message: 'Download ID is required'
			}, { status: 400 });
		}
		await locals.deps.downloadService.DeleteDownload(id);
		return json({
			success: true,
			message: 'Download deleted successfully'
		}, { status: 200 });
	} catch (error) {
		return json({
			success: false,
			message: error instanceof Error ? error.message : 'Failed to delete download'
		}, { status: 400 });
	}
}
