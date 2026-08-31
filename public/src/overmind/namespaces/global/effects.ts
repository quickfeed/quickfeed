import type { ConnectError } from "@connectrpc/connect"
import { createConnectTransport } from "@connectrpc/connect-web"
import { QuickFeedService } from "../../../../proto/qf/quickfeed_pb"
import type { ResponseClient } from "../../../client"
import { createResponseClient } from "../../../client"
import { StreamService } from "../../../streamService"


export class ApiClient {
    client!: ResponseClient<typeof QuickFeedService>

    /**
     * init initializes a client with the provided error handler.
     * Must be called before accessing the client.
     * @param errorHandler A function that is called when an error occurs.
     */
    public init(errorHandler: (payload: { method: string; error: ConnectError }) => void) {
        this.client = createResponseClient(QuickFeedService, createConnectTransport({
            baseUrl: `https://${window.location.host}`
        }), errorHandler)
    }
}

export const api = new ApiClient()

export const streamService = new StreamService()
