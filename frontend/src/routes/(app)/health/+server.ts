import { json } from "@sveltejs/kit";
import { PUBLIC_API_URL } from "$env/static/public";

export const GET = async ({ fetch }) => {
	try {
		const response = await fetch(`${PUBLIC_API_URL}/metrics`);
		if (!response.ok) {
			throw new Error(`Failed to fetch metrics: ${response.statusText}`);
		}
		const data = await response.json();
		return json({
			success: true,
			message: 'Health check successful',
			data: {
				metrics: data
			}
		}, { status: 200 });
	} catch (error) {
		return json({
			success: false,
			message: error instanceof Error ? error.message : 'Failed to check health'
		}, { status: 500 });
	}
}
