import { createOvermindMock } from "overmind"
import { config } from "../overmind"
import { State } from "../overmind/state"
import { SubType } from "overmind/lib/internalTypes"
import { ReviewState } from "../overmind/namespaces/review/state"
import { ApiClient } from "../overmind/namespaces/global/effects"
import { create } from "@bufbuild/protobuf"
import { TimestampSchema } from "@bufbuild/protobuf/wkt"

/** initializeOvermind creates a mock Overmind instance with the given state, reviewState, and mockedEffects.
 * @param state the state to initialize the mock with
 * @param mockedEffects the mocked effects to initialize the mock with
 * NOTE: Directly setting derived values in the state is not supported.
*/
export const initializeOvermind = (state: Partial<State & SubType<{ review: Partial<ReviewState> }, object>>, mockedApi?: ApiClient) => {
    const overmind = createOvermindMock(config, {
        global: {
            api: mockedApi
        }
    }, initialState => {
        Object.assign(initialState, state)
    })
    Object.assign(overmind.effects.global.api, mockedApi)
    return overmind
}

/** UnaryApiClient is a type that represents the ApiClient without streaming methods. */
interface UnaryApiClient {
    client: Omit<ApiClient["client"], "submissionStream">
}

/** Methods is a type that represents the methods of the UnaryApiClient */
type Methods = UnaryApiClient["client"]

/** mock is a helper function that takes a method and a mocked function to run in place of the method.
 *  It returns a function that can be used to replace the method in the ApiClient.
 * @param _method the method to mock
 * @param mockFn the function to run in place of the method
 * @example: mock("getSubmission", async (req) => { return { error: null, message: new Submission() } })
*/
export function mock<T extends keyof Methods>(
    _method: T,
    mockFn: (...req: Parameters<Methods[T]>) => ReturnType<Methods[T]>
): Methods[T] {
    return async function (...args: Parameters<Methods[T]>): Promise<ReturnType<Methods[T]>> { // skipcq: JS-0116
        return mockFn(...args) as ReturnType<Methods[T]>
    } as Methods[T]
}

const toTimestamp = (date: Date) => {
    const seconds = BigInt(Math.floor(date.getTime() / 1000))
    const nanos = (date.getTime() % 1000) * 1e6
    return create(TimestampSchema, { seconds, nanos })
}

const dateSet = () => {
    const date = new Date()
    return {
        date,
        year: date.getFullYear(),
        month: date.getMonth(),
        dayOfTheMonth: date.getDate(),
        dayOfTheWeek: date.getDay(),
        hours: date.getHours(),
        minutes: date.getMinutes(),
        seconds: date.getSeconds(),
        milliseconds: date.getMilliseconds(),
    }
}

/*
* timeStamp returns a TimeStamp object with input: years, months, days, and hours.
* Relative to the current date. No input gives the current date.
*/
export const timeStamp = ({ years, months, days, hours }: { years?: number, months?: number, days?: number, hours?: number } = {}) => {
    const set = dateSet()
    const add = (value: number, mentor?: number) => {
        return value + (mentor ?? 0)
    }
    const year = add(set.year, years)
    const month = add(set.month, months)
    const day = add(set.dayOfTheMonth, days)
    const hour = add(set.hours, hours)

    return toTimestamp(new Date(year, month, day, hour))
}
