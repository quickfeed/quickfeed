import { createConnectTransport } from "@connectrpc/connect-web"
import { ConnectError } from "@connectrpc/connect"
import { QuickFeedService } from "../../../../proto/qf/quickfeed_pb"
import { createResponseClient, ResponseClient } from "../../../client"
import { StreamService } from "../../../streamService"


export class ApiClient {
    client: ResponseClient<typeof QuickFeedService>

    /**
     * init initializes a client with the provided error handler.
     * Must be called before accessing the client.
     * @param errorHandler A function that is called when an error occurs.
     */
    public init(errorHandler: (payload?: { method: string; error: ConnectError } | undefined) => void) {
        this.client = createResponseClient(QuickFeedService, createConnectTransport({
            baseUrl: `https://${window.location.host}`
        }), errorHandler)
    }
}

export const api = new ApiClient()

export const streamService = new StreamService()
