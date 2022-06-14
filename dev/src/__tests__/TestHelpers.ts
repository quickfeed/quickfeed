import { Browser, Builder, IRectangle, ThenableWebDriver } from "selenium-webdriver"

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
        throw new Error("No supported browsers found. Please check the browsers.json file.")
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
