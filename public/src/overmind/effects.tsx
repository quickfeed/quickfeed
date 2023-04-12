import { createConnectTransport } from "@bufbuild/connect-web"
import { QuickFeedService } from "../../proto/qf/quickfeed_connectweb"
import { StreamService } from "../streamService"
import { ResponseClient, createResponseClient } from "../client"

export const client: ResponseClient<typeof QuickFeedService> = (() => {
    return createResponseClient(QuickFeedService, createConnectTransport({
        baseUrl: `https://${window.location.host}`,
    }))
})()


export const streamService = new StreamService()
