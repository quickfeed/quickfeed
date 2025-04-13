import { QuickFeedService } from '../proto/qf/quickfeed_pb'
import { Code, createClient, Client } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { ConnStatus } from './Helpers'
import { Submission, Notification } from '../proto/qf/types_pb'

type options<entity> = {
    onMessage: (payload?: entity | undefined) => void,
    onError: (error: Error) => void
    onStatusChange: (status: ConnStatus) => void
}

type error = {
    code: Code
}

export class StreamService {
    private readonly service: Client<typeof QuickFeedService>
    private backoff = 1000

    constructor() {
        this.service = createClient(QuickFeedService, createConnectTransport({ baseUrl: `https://${window.location.host}` }))
    }

    // timeout returns a promise that resolves after the current backoff has elapsed
    private async timeout() {
        return new Promise(resolve => setTimeout(resolve, this.backoff))
    }

    private async handleStreamError<T>(func: (options: options<T>) => Promise<void>, error: error, options: options<T>, streamType: string) {
        if (error.code === Code.Canceled) {
            // The stream was canceled, so we don't need to reconnect.
            // This happens when the stream is closed by the server
            // which happens only if the user opens a new stream, i.e., opens the frontend in a new tab.
            options.onError(new Error(`Stream ${streamType} was canceled by the server.`))
            return
        }

        // Attempt to reconnect up to log2(128) + 1 times, increasing delay between attempts by 2x each time
        // This is a total of 8 attempts with a maximum delay of 255 seconds
        if (this.backoff <= 128 * 1000) {
            // Attempt to reconnect after a backoff
            options.onStatusChange(ConnStatus.RECONNECTING)
            await this.timeout()
            func(options)
            this.backoff *= 2
        } else {
            this.backoff = 1000
            options.onError(new Error("An error occurred while connecting to the server"))
        }
    }

    public async submissionStream(options: options<Submission>) {
        const stream = this.service.submissionStream({})
        try {
            options.onStatusChange(ConnStatus.CONNECTED)
            for await (const msg of stream) {
                options.onMessage(msg)
            }
        } catch (error) {
            this.handleStreamError(this.submissionStream, error, options, "Submission")
        }
    }

    public async notificationStream(options: options<Notification>) {
        const stream = this.service.notificationStream({})
        try {
            options.onStatusChange(ConnStatus.CONNECTED)
            for await (const msg of stream) {
                options.onMessage(msg)
            }
        } catch (error) {
            this.handleStreamError(this.notificationStream, error, options, "Notification")
        }
    }
}
