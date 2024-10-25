import { CallOptions, ConnectError, Transport, makeAnyClient } from "@connectrpc/connect";
import { createAsyncIterable } from "@connectrpc/connect/protocol";
import {
    DescMessage,
    DescMethodServerStreaming,
    DescMethodUnary,
    DescService,
    Message,
    MessageInitShape,
    MessageShape,
} from "@bufbuild/protobuf";

export type Response<T extends Message> = {
    error: ConnectError | null
    message: T
}

export type ResponseClient<Desc extends DescService> = {
    [P in keyof Desc["method"]]: 
    Desc["method"][P] extends DescMethodUnary<infer I, infer O> ? (request: MessageInitShape<I>, options?: CallOptions) => Promise<Response<MessageShape<O>>> : 
    Desc["method"][P] extends DescMethodServerStreaming<infer I, infer O> ? (request: MessageInitShape<I>, options?: CallOptions) => AsyncIterable<MessageShape<O>> :
    never
};
/**
 * Create a ResponseClient for the given service, invoking RPCs through the
 * given transport.
 */
export function createResponseClient<T extends DescService>(service: T, transport: Transport, errorHandler: (payload?: { method: string; error: ConnectError; } | undefined) => void): ResponseClient<T> {
    return makeAnyClient(service, (method) => {
        switch (method.methodKind) {
            case "unary":
                return createUnaryFn(transport, method as DescMethodUnary, errorHandler);
            case "server_streaming":
                return createServerStreamingFn(transport, method as DescMethodServerStreaming);
            default:
                return null;
        }
    }) as ResponseClient<T>;
}
/**
 * UnaryFn is the method signature for a unary method of a ResponseClient.
 */
type UnaryFn<I extends DescMessage, O extends DescMessage> = (request: MessageInitShape<I>, options?: CallOptions) => Promise<Response<MessageShape<O>>>;
export function createUnaryFn<I extends DescMessage, O extends DescMessage>(transport: Transport, method: DescMethodUnary<I, O>, errorHandler: (payload?: { method: string; error: ConnectError; } | undefined) => void): UnaryFn<I, O> {
    return async function (input, options) {
        try {
            const response = await transport.unary(
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
            } as Response<MessageShape<O>>;
        }
        catch (error) {
            errorHandler({ method: method.name, error });
            return {
                error,
            } as Response<MessageShape<O>>;
        }
    }
}
/**
 * ServerStreamingFn is the method signature for a server-streaming method of
 * a ResponseClient.
 */
type ServerStreamingFn<I extends DescMessage, O extends DescMessage> = (request: MessageInitShape<I>, options?: CallOptions) => AsyncIterable<MessageShape<O>>;
export function createServerStreamingFn<I extends DescMessage, O extends DescMessage>(transport: Transport, method: DescMethodServerStreaming<I, O>): ServerStreamingFn<I, O> {
    return async function* (input, options): AsyncIterable<MessageShape<O>> {
        const response = await transport.stream<I, O>(
            method,
            options?.signal,
            options?.timeoutMs,
            options?.headers,
            createAsyncIterable([input])
        );
        options?.onHeader?.(response.header);
        yield* response.message;
        options?.onTrailer?.(response.trailer);
    }
}
