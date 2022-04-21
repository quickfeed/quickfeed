import { Builder, By, Browser, IRectangle, Capabilities } from 'selenium-webdriver'

describe("Front page elements login and logo should not overlap", () => {
    //Laptop, desktop, mobile
    it("Should not overlap on res 1920x1080", async () => {
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev")
        await driver.manage().window().setRect({ width: 1920, height: 1080 })

        let rect: IRectangle
        let rect2: IRectangle

        const logo = await driver.findElement(By.className("navbar-brand"))
        const signIn = await driver.findElement(By.className("signIn"))
        rect = await logo.getRect()
        rect2 = await signIn.getRect()

        var overlap = (rect.x < rect2.x + rect2.width &&
            rect.x + rect.width > rect2.x &&
            rect.y < rect2.y + rect2.height &&
            rect.height + rect.y > rect2.y)

        jest.setTimeout(50000)
        expect(overlap).toBe(false)
    })
    it("Should not overlap on res 1366×768", async () => {
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev")
        await driver.manage().window().setRect({ width: 1366, height: 768 })
        let rect: IRectangle
        let rect2: IRectangle

        const logo = await driver.findElement(By.className("navbar-brand"))
        const signIn = await driver.findElement(By.className("signIn"))
        rect = await logo.getRect()
        rect2 = await signIn.getRect()

        var overlap = (rect.x < rect2.x + rect2.width &&
            rect.x + rect.width > rect2.x &&
            rect.y < rect2.y + rect2.height &&
            rect.height + rect.y > rect2.y)

        jest.setTimeout(50000)
        expect(overlap).toBe(false)
    })
    it("Should not overlap on splitscreen res 1920x1080", async () => {
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev")
        await driver.manage().window().setRect({ width: 960, height: 1080 })

        let rect: IRectangle
        let rect2: IRectangle

        const logo = await driver.findElement(By.className("navbar-brand"))
        const signIn = await driver.findElement(By.className("signIn"))
        rect = await logo.getRect()
        rect2 = await signIn.getRect()

        var overlap = (rect.x < rect2.x + rect2.width &&
            rect.x + rect.width > rect2.x &&
            rect.y < rect2.y + rect2.height &&
            rect.height + rect.y > rect2.y)

        jest.setTimeout(50000)
        expect(overlap).toBe(false)
    })
    it("Should not overlap on splitscreen res 1366×768", async () => {
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev")
        await driver.manage().window().setRect({ width: 683, height: 768 })

        let rect: IRectangle
        let rect2: IRectangle

        const logo = await driver.findElement(By.className("navbar-brand"))
        const signIn = await driver.findElement(By.className("signIn"))
        rect = await logo.getRect()
        rect2 = await signIn.getRect()

        var overlap = (rect.x < rect2.x + rect2.width &&
            rect.x + rect.width > rect2.x &&
            rect.y < rect2.y + rect2.height &&
            rect.height + rect.y > rect2.y)

        jest.setTimeout(50000)
        expect(overlap).toBe(false)
    })
})
