import { Builder, By, Capabilities, until } from 'selenium-webdriver'
import { isOverlapping } from './testHelpers'


describe("Course utility elements should not overlap", () => {
    //Laptop, desktop, mobile
    it("Should not overlap on split screen 1920x1080", async () => {
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev/#/course/1")
        await driver.manage().window().setRect({ width: 960, height: 1080 })

        if (await driver.findElement(By.className("closeButton")).isDisplayed()) {
            driver.findElement(By.className("closeButton")).click()
        }

        const switchStudent = driver.findElement(By.className("clickable"))
        await driver.wait(until.elementIsVisible(switchStudent), 100)
        switchStudent.click()

        await driver.sleep(1000)
        const labs = await driver.findElement(By.className("col-md-9"))
        await driver.wait(until.elementIsVisible(labs), 5000)

        const utility = await driver.findElement(By.className("list-group width-resize"))
        await driver.wait(until.elementIsVisible(utility), 5000)
        let rect = await utility.getRect()
        let rect2 = await labs.getRect()

        var overlap = isOverlapping(rect, rect2)

        jest.setTimeout(100000)
        expect(overlap).toBe(false)
    })
    it("Should not overlap on mobile res 360x740", async () => { //More common phones
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev/#/course/1")
        await driver.manage().window().setRect({ width: 360, height: 740 })

        if (await driver.findElement(By.className("closeButton")).isDisplayed()) {
            driver.findElement(By.className("closeButton")).click()
        }

        const switchStudent = driver.findElement(By.className("clickable"))
        await driver.wait(until.elementIsVisible(switchStudent), 100)
        switchStudent.click()

        await driver.sleep(1000)
        const labs = await driver.findElement(By.className("col-md-9"))
        await driver.wait(until.elementIsVisible(labs), 5000)

        const utility = await driver.findElement(By.className("list-group width-resize"))
        await driver.wait(until.elementIsVisible(utility), 5000)
        let rect = await utility.getRect()
        let rect2 = await labs.getRect()

        var overlap = isOverlapping(rect, rect2)

        jest.setTimeout(100000)
        expect(overlap).toBe(false)
    })
    it("Should not overlap on mobile res 360x640", async () => { //Older phones
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev/#/course/1")
        await driver.manage().window().setRect({ width: 360, height: 640 })

        if (await driver.findElement(By.className("closeButton")).isDisplayed()) {
            driver.findElement(By.className("closeButton")).click()
        }

        const switchStudent = driver.findElement(By.className("clickable"))
        await driver.wait(until.elementIsVisible(switchStudent), 100)
        switchStudent.click()

        await driver.sleep(1000)
        const labs = await driver.findElement(By.className("col-md-9"))
        await driver.wait(until.elementIsVisible(labs), 5000)

        const utility = await driver.findElement(By.className("list-group width-resize"))
        await driver.wait(until.elementIsVisible(utility), 5000)
        let rect = await utility.getRect()
        let rect2 = await labs.getRect()

        var overlap = isOverlapping(rect, rect2)

        jest.setTimeout(100000)
        expect(overlap).toBe(false)
    })
    it("Should not overlap on mobile res 412x914", async () => { //Bigger phones
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev/#/course/1")
        await driver.manage().window().setRect({ width: 412, height: 914 })

        if (await driver.findElement(By.className("closeButton")).isDisplayed()) {
            driver.findElement(By.className("closeButton")).click()
        }

        const switchStudent = driver.findElement(By.className("clickable"))
        await driver.wait(until.elementIsVisible(switchStudent), 100)
        switchStudent.click()

        await driver.sleep(1000)
        const labs = await driver.findElement(By.className("col-md-9"))
        await driver.wait(until.elementIsVisible(labs), 5000)

        const utility = await driver.findElement(By.className("list-group width-resize"))
        await driver.wait(until.elementIsVisible(utility), 5000)
        let rect = await utility.getRect()
        let rect2 = await labs.getRect()

        var overlap = isOverlapping(rect, rect2)

        jest.setTimeout(100000)
        expect(overlap).toBe(false)
    })
})
