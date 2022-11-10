import { QuickFeedService } from '../proto/qf/quickfeed_connectweb'
import { Submission } from '../proto/qf/types_pb'
import { Code, createConnectTransport, createPromiseClient, PromiseClient } from "@bufbuild/connect-web"
import { ConnStatus } from './Helpers'


export class StreamService {
    private service: PromiseClient<typeof QuickFeedService>
    private backoff = 1000

    constructor() {
        this.service = createPromiseClient(QuickFeedService, createConnectTransport({ baseUrl: "https://" + window.location.host }))
    }

    // timeout returns a promise that resolves after the current backoff has elapsed
    private async timeout() {
        return new Promise(resolve => setTimeout(resolve, this.backoff))
    }

    public async submissionStream(options: {
        onMessage: (payload?: Submission | undefined) => void,
        onError: (error: Error) => void
        onStatusChange: (status: ConnStatus) => void
    }) {
        const stream = this.service.submissionStream({})
        try {
            options.onStatusChange(ConnStatus.CONNECTED)
            for await (const msg of stream) {
                options.onMessage(msg)
            }
        } catch (error) {
            if (error.code === Code.Canceled) {
                // The stream was canceled, so we don't need to reconnect.
                // This happens when the stream is closed by the server
                // which happens only if the user opens a new stream, i.e., opens the frontend in a new tab.
                options.onError(new Error("Stream was canceled by the server."))
                return
            }

            // Attempt to reconnect up to log2(128) + 1 times, increasing delay between attempts by 2x each time
            // This is a total of 8 attempts with a maximum delay of 255 seconds
            if (this.backoff <= 128 * 1000) {
                // Attempt to reconnect after a backoff
                options.onStatusChange(ConnStatus.RECONNECTING)
                await this.timeout()
                this.submissionStream(options)
                this.backoff *= 2
            } else {
                this.backoff = 1000
                options.onError(new Error("An error occurred while connecting to the server"))
            }
        }
    }
}
