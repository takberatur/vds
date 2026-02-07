import type { RequestEvent } from '@sveltejs/kit';
import { BaseService } from './base_service';
import type { ContactSchema, DownloadVideoSchema, WebErrorReportSchema } from '@/utils/schema';

export class WebServiceImpl extends BaseService implements WebService {
	constructor(
		protected readonly event: RequestEvent,
		private readonly api: ApiClient) {
		super(event);
	}

	async Contact(data: ContactSchema): Promise<void | Error> {
		try {
			const response = await this.api.publicRequest('POST', '/web-client/contact', data, true);
			if (!response.success) {
				throw new Error(response.message || 'Failed to contact');
			}
			return void 0;
		} catch (error) {
			const errMsg = error instanceof Error ? error : new Error('Unknown error');
			console.error('Error contacting:', errMsg);
			return errMsg;
		}
	}
	async DownloadVideo(data: DownloadVideoSchema): Promise<ApiResponse<Download>> {
		return await this.api.publicRequest<Download>('POST', '/web-client/download/process/video', data, true);
	}
	async DownloadVideoToMp3(data: DownloadVideoSchema): Promise<ApiResponse<Download>> {
		return await this.api.publicRequest<Download>('POST', '/web-client/download/process/mp3', data, true);
	}
	async ReportError(data: WebErrorReportSchema): Promise<void | Error> {
		try {
			const response = await this.api.publicRequest('POST', '/web-client/report/errors', data, true);
			if (!response.success) {
				throw new Error(response.message || 'Failed to report error');
			}
			return void 0;
		} catch (error) {
			const errMsg = error instanceof Error ? error : new Error('Unknown error');
			console.error('Error reporting:', errMsg);
			return errMsg;
		}
	}
}
