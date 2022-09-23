import { QuickFeedService } from '../gen/qf/quickfeed_connectweb'
import { Submission } from '../gen/qf/types_pb'
import { Code, ConnectError, createConnectTransport, createPromiseClient, PromiseClient } from "@bufbuild/connect-web"


export class StreamService {
    private service: PromiseClient<typeof QuickFeedService>

    constructor() {
        this.service =  createPromiseClient(QuickFeedService, createConnectTransport({baseUrl: "https://" + window.location.host}))
    }


    public async submissionStream() {
        const stream = this.service.submissionStream({})   
        try {
            window.dispatchEvent(new CustomEvent("submission"))
            for await (const msg of stream) {
                window.dispatchEvent(new CustomEvent<Submission>("substream", {detail: msg}))
            }
        } catch (error) {
            if (error instanceof ConnectError) {
                console.table({error})
                if (error.code === Code.NotFound) {
                    //window.dispatchEvent(new CustomEvent("streamdone"))
                } else {
                    // handle other errors
                }
            } else {
                // handle other errors
                // typically this should occur if the server closes the stream
                console.log("Error: ", error)
            }
        }
    }
}

