import { QuickFeedService } from '../proto/qf/quickfeed_pb';
import { Code, createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { ConnStatus } from './Helpers';
export class StreamService {
    service;
    backoff = 1000;
    constructor() {
        this.service = createClient(QuickFeedService, createConnectTransport({ baseUrl: `https://${window.location.host}` }));
    }
    timeout() {
        return new Promise(resolve => setTimeout(resolve, this.backoff));
    }
    async submissionStream(options) {
        const stream = this.service.submissionStream({});
        try {
            options.onStatusChange(ConnStatus.CONNECTED);
            for await (const msg of stream) {
                options.onMessage(msg);
            }
        }
        catch (error) {
            if (error.code === Code.Canceled) {
                options.onError(new Error("Stream was canceled by the server."));
                return;
            }
            if (this.backoff <= 128 * 1000) {
                options.onStatusChange(ConnStatus.RECONNECTING);
                await this.timeout();
                await this.submissionStream(options);
                this.backoff *= 2;
            }
            else {
                this.backoff = 1000;
                options.onError(new Error("An error occurred while connecting to the server"));
            }
        }
    }
}
