import { makeAnyClient } from "@connectrpc/connect";
import { createAsyncIterable } from "@connectrpc/connect/protocol";
export function createResponseClient(service, transport, errorHandler) {
    return makeAnyClient(service, (method) => {
        switch (method.methodKind) {
            case "unary":
                return createUnaryFn(transport, method, errorHandler);
            case "server_streaming":
                return createServerStreamingFn(transport, method);
            default:
                return null;
        }
    });
}
export function createUnaryFn(transport, method, errorHandler) {
    return async function (input, options) {
        try {
            const response = await transport.unary(method, options?.signal, options?.timeoutMs, options?.headers, input);
            options?.onHeader?.(response.header);
            options?.onTrailer?.(response.trailer);
            return {
                error: null,
                message: response.message
            };
        }
        catch (error) {
            errorHandler({ method: method.name, error });
            return {
                error,
            };
        }
    };
}
export function createServerStreamingFn(transport, method) {
    return async function* (input, options) {
        const response = await transport.stream(method, options?.signal, options?.timeoutMs, options?.headers, createAsyncIterable([input]));
        options?.onHeader?.(response.header);
        yield* response.message;
        options?.onTrailer?.(response.trailer);
    };
}
