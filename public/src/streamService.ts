import { QuickFeedService } from '../gen/qf/quickfeed_connectweb'
import { Submission } from '../gen/qf/types_pb'
import { createConnectTransport, createPromiseClient, PromiseClient } from "@bufbuild/connect-web"


export class StreamService {
    private service: PromiseClient<typeof QuickFeedService>
    private backoff = 1000

    constructor() {
        this.service =  createPromiseClient(QuickFeedService, createConnectTransport({baseUrl: "https://" + window.location.host}))
    }

    // timeout returns a promise that resolves after the current backoff has elapsed
    private async timeout() {
        return new Promise(resolve => setTimeout(resolve, this.backoff))
    }

    public async submissionStream(options: {onMessage: (payload?: Submission | undefined) => void, onError: (error: Error) => void}) {
        const stream = this.service.submissionStream({})   
        try {
            for await (const msg of stream) {
                options.onMessage(msg)
            }
        } catch (error) {
            /* TODO(jostein): 
             * Our streams currently time out after 2 minutes.
             * Once (if) https://github.com/golang/go/issues/54136 is accepted
             * we should be able to set longer timeouts for streaming requests.
             */
            // TODO: Figure out we should wait even longer
            // TODO: onError could prompt the user to manually reconnect
            
            // Attempt to reconnect up to log2(128) + 1 times, increasing delay between attempts by 2x each time
            // This is a total of 8 attempts with a maximum delay of 255 seconds
            if (this.backoff <= 128 * 1000) {
                // Attempt to reconnect after a backoff
                await this.timeout()
                this.submissionStream(options)
                this.backoff *= 2
            } else {
                this.backoff = 1000
                options.onError(error)
            }
        }
    }
}

