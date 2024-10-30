import { createOvermindMock } from "overmind"
import { Browser, Builder, IRectangle, ThenableWebDriver } from "selenium-webdriver"
import { config } from "../overmind"
import { State } from "../overmind/state"
import { SubType } from "overmind/lib/internalTypes"
import { ReviewState } from "../overmind/namespaces/review/state"
import { ApiClient } from "../overmind/effects"

export const isOverlapping = (rect: IRectangle, rect2: IRectangle) => {
    return (rect.x < rect2.x + rect2.width &&
        rect.x + rect.width > rect2.x &&
        rect.y < rect2.y + rect2.height &&
        rect.height + rect.y > rect2.y)
}

// getBuilders returns an array of builders for all the browsers that are supported.
// Supported browsers must be defined in the config.json file.
// These builders require you to have web driver executables in your PATH.
const getBuilders = (): Builder[] => {
    const config = require("./config.json")
    const builders: Builder[] = []
    Object.entries(Browser).forEach(([key, value]) => {
        if (key in config.BROWSERS && config.BROWSERS[key]) {
            const builder = createBuilder(value)
            builders.push(builder)
        }
    })
    if (builders.length === 0) {
        throw new Error("No supported browsers found. Please check the config.json file.")
    }
    return builders
}

const createBuilder = (browser: string) => {
    return new Builder().forBrowser(browser)
}

// getBaseUrl returns the base url to be used for the tests.
export const getBaseUrl = (): string => {
    const config = require("./config.json")
    let url = config.BASE_URL as string
    if (url.endsWith("/")) {
        url = url.slice(0, -1)
    }
    return url
}

/**
 *  setupDrivers returns an array of drivers for all the browsers that are supported.
 *  @path (optional) the path to load in the browser.
 *  @example: setupDrivers("/path/to/page")
 */
export const setupDrivers = (path?: string): ThenableWebDriver[] => {
    const baseUrl = getBaseUrl()
    const builders = getBuilders()
    const drivers = builders.map(driver => driver.build())

    beforeAll(async () => {
        // Open the page to be tested in all browsers, before running tests
        await Promise.all(drivers.map(driver => driver.get(baseUrl + (path ? path : ""))))
    })

    afterAll(async () => {
        // Close all drivers after the tests are done
        await Promise.all(drivers.map(driver => driver.quit()))
    })

    return drivers
}

/** initializeOvermind creates a mock Overmind instance with the given state, reviewState, and mockedEffects.
 * @param state the state to initialize the mock with
 * @param mockedEffects the mocked effects to initialize the mock with
 * NOTE: Directly setting derived values in the state is not supported.
*/
export const initializeOvermind = (state: Partial<State & SubType<{ review: Partial<ReviewState>; }, object>>, mockedApi?: ApiClient) => {
    const overmind = createOvermindMock(config, {
        api: mockedApi
    }, initialState => {
        Object.assign(initialState, state)
    });
    Object.assign(overmind.effects.api, mockedApi)
    return overmind
}

/** UnaryApiClient is a type that represents the ApiClient without streaming methods. */
interface UnaryApiClient {
    client: Omit<ApiClient["client"], "submissionStream">;
}

/** Methods is a type that represents the methods of the UnaryApiClient */
type Methods = UnaryApiClient["client"];

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
    return async function (...args: Parameters<Methods[T]>): Promise<ReturnType<Methods[T]>> {
        return mockFn(...args) as ReturnType<Methods[T]>;
    } as Methods[T];
}
