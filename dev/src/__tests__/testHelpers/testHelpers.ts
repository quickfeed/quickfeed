import { Browser, Builder, IRectangle, ThenableWebDriver } from "selenium-webdriver"

export const isOverlapping = (rect: IRectangle, rect2: IRectangle) => {
    return (rect.x < rect2.x + rect2.width &&
        rect.x + rect.width > rect2.x &&
        rect.y < rect2.y + rect2.height &&
        rect.height + rect.y > rect2.y)
}

// getBuilders returns an array of builders for all the browsers that are supported
// Supported browsers must be defined in the config file: browsers.json
// NOTE: These builders require you to have web driver executables in your PATH
const getBuilders = (): Builder[] => {
    const browsers = require("./browsers.json")
    const builders: Builder[] = []
    Object.entries(Browser).forEach(([key, value]) => {
        if (key in browsers && browsers[key]) {
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
    const builder = new Builder().forBrowser(browser)
    return builder
}

/** setupDrivers returns an array of drivers for all the browsers that are supported
 *  @url (optional) the url to load in the browser
 */
export const setupDrivers = (url?: string): ThenableWebDriver[] => {
    const builders = getBuilders()
    const drivers = builders.map(driver => driver.build())
    beforeAll(async () => {
        // Open the page to be tested in all browsers, before running tests
        await Promise.all(drivers.map(driver => driver.get(url ?? "http://localhost:8082")))
    })

    afterAll(async () => {
        // Close all drivers after the tests are done
        await Promise.all(drivers.map(driver => driver.quit()))
    })

    return drivers
}
