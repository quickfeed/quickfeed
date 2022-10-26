import { QuickFeedService } from '../gen/qf/quickfeed_connectweb'
import { Submission } from '../gen/qf/types_pb'
import { createConnectTransport, createPromiseClient, PromiseClient } from "@bufbuild/connect-web"


export class StreamService {
    private service: PromiseClient<typeof QuickFeedService>
    private backoff = 1000

    constructor() {
        this.service =  createPromiseClient(QuickFeedService, createConnectTransport({baseUrl: "https://" + window.location.host}))
    }

    // public async notificationStream() {
    //     const stream = this.service.notificationStream({})
    //     try {
    //         window.dispatchEvent(new CustomEvent("startstream"))
    //         for await (const msg of stream) {
    //             window.dispatchEvent(new CustomEvent<Notification>("substream", {detail: msg}))
    //         }
    //     } catch (error) {
    //         if (error instanceof ConnectError) {
    //             console.table({error})
    //             if (error.code === Code.NotFound) {
    //                 //window.dispatchEvent(new CustomEvent("streamdone"))
    //             } else {
    //                 // handle other errors
    //             }
    //         } else {
    //             // handle other errors
    //             // typically this should occur if the server closes the stream
    //             console.log("Error: ", error)
    //         }
    //     } finally {
    //         window.dispatchEvent(new CustomEvent("streamdead"))
    //     }
    // }


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

