import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';
import type { ContactSchema, DownloadVideoSchema } from '@/utils/schema';



export class WebServiceImpl extends BaseService implements WebService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}

	async Contact(data: ContactSchema): Promise<void | Error> {
		try {
			const response = await this.api.publicRequest('POST', '/web-client/contact', data);
			if (!response.success) {
				throw new Error(response.message || 'Failed to contact');
			}
			return void 0;
		} catch (error) {
			return error instanceof Error ? error : new Error('Unknown error');
		}
	}

	async DownloadVideo(data: DownloadVideoSchema): Promise<ApiResponse<Download>> {
		return await this.api.publicRequest<Download>('POST', '/web-client/download/process/video', data, true);
	}
	async DownloadVideoToMp3(data: DownloadVideoSchema): Promise<ApiResponse<Download>> {
		return await this.api.publicRequest<Download>('POST', '/web-client/download/process/mp3', data, true);
	}
}
