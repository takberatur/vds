import { json } from "@sveltejs/kit";

export const DELETE = async ({ request, locals, params }) => {
	try {
		const id = params.id;
		if (!id) {
			return json({ success: false, message: 'Platform id is required' }, { status: 400 });
		}

		const response = await locals.deps.platformService.DeletePlatform(id);
		if (response instanceof Error) {
			throw response;
		}
		return json({ success: true, message: 'Platform deleted successfully' }, { status: 200 });
	} catch (error) {
		return json({ success: false, message: error instanceof Error ? error.message : 'Failed to delete platform' }, { status: 500 });
	}
}
