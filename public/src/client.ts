import { CallOptions, ConnectError, Transport, makeAnyClient } from "@bufbuild/connect";
import { createAsyncIterable } from "@bufbuild/connect/protocol";
import {
    MethodInfo,
    MethodInfoServerStreaming,
    MethodInfoUnary,
    PartialMessage,
    ServiceType,
    Message,
    MethodKind,
    AnyMessage,
} from "@bufbuild/protobuf";

export type Response<T extends AnyMessage> = {
    error: ConnectError | null
    message: T
}

/**
 * ResponseClient is a simple client that supports unary and server-streaming
 * methods. Methods will produce a promise for the response message,
 * or an asynchronous iterable of response messages.
 */
export type ResponseClient<T extends ServiceType> = {
    [P in keyof T["methods"]]:
    T["methods"][P] extends MethodInfoUnary<infer I, infer O> ? (request: PartialMessage<I>, options?: CallOptions) => Promise<Response<O>>
    : T["methods"][P] extends MethodInfoServerStreaming<infer I, infer O> ? (request: PartialMessage<I>, options?: CallOptions) => AsyncIterable<O>
    : never;
};

/**
 * Create a ResponseClient for the given service, invoking RPCs through the
 * given transport.
 */
export function createResponseClient<T extends ServiceType>(
    service: T,
    transport: Transport
) {
    return makeAnyClient(service, (method) => {
        switch (method.kind) {
            case MethodKind.Unary:
                return createUnaryFn(transport, service, method);
            case MethodKind.ServerStreaming:
                return createServerStreamingFn(transport, service, method);
            default:
                return null;
        }
    }) as ResponseClient<T>;
}

/**
  * UnaryFn is the method signature for a unary method of a ResponseClient.
  */
type UnaryFn<I extends Message<I>, O extends Message<O>> = (
    request: PartialMessage<I>,
    options?: CallOptions
) => Promise<Response<O>>;

function createUnaryFn<I extends Message<I>, O extends Message<O>>(
    transport: Transport,
    service: ServiceType,
    method: MethodInfo<I, O>
): UnaryFn<I, O> {
    return async function (input, options) {
        try {
            const response = await transport.unary(
                service,
                method,
                options?.signal,
                options?.timeoutMs,
                options?.headers,
                input
            );
            options?.onHeader?.(response.header);
            options?.onTrailer?.(response.trailer);
            return {
                error: null,
                message: response.message
            } as Response<O>;
        }
        catch (error) {
            return {
                error: error,
            } as Response<O>;
        }
    }
}

/**
 * ServerStreamingFn is the method signature for a server-streaming method of
 * a ResponseClient.
 */
type ServerStreamingFn<I extends Message<I>, O extends Message<O>> = (
    request: PartialMessage<I>,
    options?: CallOptions
) => AsyncIterable<O>;

export function createServerStreamingFn<
    I extends Message<I>,
    O extends Message<O>
>(
    transport: Transport,
    service: ServiceType,
    method: MethodInfo<I, O>
): ServerStreamingFn<I, O> {
    return async function* (input, options): AsyncIterable<O> {
        const inputMessage =
            input instanceof method.I ? input : new method.I(input);
        const response = await transport.stream<I, O>(
            service,
            method,
            options?.signal,
            options?.timeoutMs,
            options?.headers,
            createAsyncIterable([inputMessage])
        );
        options?.onHeader?.(response.header);
        yield* response.message;
        options?.onTrailer?.(response.trailer);
    }
}
