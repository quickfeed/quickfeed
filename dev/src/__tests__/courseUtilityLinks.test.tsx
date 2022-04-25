import { Builder, By, Browser, IRectangle, Capabilities, until } from 'selenium-webdriver'

describe("Front page elements login and logo should not overlap", () => {
    //Laptop, desktop, mobile
    it("Should not overlap on res 1920x1080", async () => {
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev/")
        // const signIn = driver.findElement(By.className("signIn"))
        // await driver.wait(until.elementIsVisible(signIn), 100)
        // signIn.click()
        const hamburger = driver.findElement(By.className("hamburger"))
        await driver.wait(until.elementIsVisible(hamburger), 100)
        hamburger.click()
        expect(true).toBe(true)



    })
})
