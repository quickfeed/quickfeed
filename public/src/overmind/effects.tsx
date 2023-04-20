import { createConnectTransport } from "@bufbuild/connect-web"
import { QuickFeedService } from "../../proto/qf/quickfeed_connectweb"
import { StreamService } from "../streamService"
import { createResponseClient } from "../client"

export const client = createResponseClient(QuickFeedService, createConnectTransport({
    baseUrl: `https://${window.location.host}`
}))

export const streamService = new StreamService()
