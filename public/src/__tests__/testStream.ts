/** A pushable async iterable, standing in for a server stream in tests. */
export interface TestStream<T> {
    iterable: AsyncIterable<T>
    /** push delivers one message to whoever is iterating. */
    push: (message: T) => void
    /** close ends the stream the way a bounded request does. */
    close: () => void
    /** fail ends the stream with an error, the way a rejected request does. */
    fail: (error: unknown) => void
}

export const makeStream = <T>(): TestStream<T> => {
    const buffer: T[] = []
    let wake: (() => void) | null = null
    let done = false
    let failure: unknown = null
    const notify = () => {
        const resume = wake
        wake = null
        resume?.()
    }
    return {
        iterable: {
            async *[Symbol.asyncIterator]() {
                for (; ;) {
                    if (buffer.length > 0) {
                        yield buffer.shift() as T
                        continue
                    }
                    if (failure !== null) {
                        const thrown = failure
                        failure = null
                        throw thrown
                    }
                    if (done) {
                        return
                    }
                    await new Promise<void>(resolve => { wake = resolve })
                }
            },
        },
        push: message => { buffer.push(message); notify() },
        close: () => { done = true; notify() },
        fail: error => { failure = error; notify() },
    }
}
