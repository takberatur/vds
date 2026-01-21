import { json } from "@sveltejs/kit";

export const POST = async ({ locals, request }) => {
	const { deps } = locals;
	try {
		const formData = await request.formData();
		const platformID = formData.get('id') as string;
		const file = formData.get('file') as File;
		if (!platformID) {
			return json({
				success: false,
				message: 'No platform ID provided'
			}, { status: 400 });
		}
		if (!file) {
			return json({
				success: false,
				message: 'No file uploaded'
			}, { status: 400 });
		}

		const response = await deps.platformService.UploadThumbnail(platformID, file);
		if (response instanceof Error) {
			throw response;
		}
		return json({
			success: true,
			message: 'Thumbnail updated successfully',
			data: {
				url: response
			}
		}, { status: 200 });


	} catch (error) {
		return json({
			success: false,
			message: error instanceof Error ? error.message : 'Failed to upload thumbnail'
		}, { status: 500 });
	}
}
