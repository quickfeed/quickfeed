import { Builder, By, Browser, IRectangle, Capabilities } from 'selenium-webdriver'


describe("Check rectangles", () => {
    it("Get `box` rectangle", async () => {
        let rect: IRectangle

        // Start Chrome at `http://localhost:8080`
        // let driver = await new Builder().forBrowser(Browser.FIREFOX).build()
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev")
        await driver.manage().window().maximize()

        jest.setTimeout(5000)
        const el = await driver.findElement(By.className("btn"))
        rect = await el.getRect()
        console.log(rect.x, rect.y)

        expect(rect.width).toBeGreaterThan(0)
    })
})
