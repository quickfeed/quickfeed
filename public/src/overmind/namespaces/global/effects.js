import { createConnectTransport } from "@connectrpc/connect-web";
import { QuickFeedService } from "../../../../proto/qf/quickfeed_pb";
import { createResponseClient } from "../../../client";
import { StreamService } from "../../../streamService";
export class ApiClient {
    client;
    init(errorHandler) {
        this.client = createResponseClient(QuickFeedService, createConnectTransport({
            baseUrl: `https://${window.location.host}`
        }), errorHandler);
    }
}
export const api = new ApiClient();
export const streamService = new StreamService();
