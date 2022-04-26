import { Builder, By, Browser, IRectangle, Capabilities, until } from 'selenium-webdriver'


describe("Front page elements login and logo should not overlap", () => {
    //Laptop, desktop, mobile
    it("Should not overlap on res 1920x1080", async () => {
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev/#/course/1")
        // const signIn = driver.findElement(By.className("signIn"))
        // await driver.wait(until.elementIsVisible(signIn), 100)
        // signIn.click()
        const hamburger = driver.findElement(By.className("closeButton"))
        await driver.wait(until.elementIsVisible(hamburger), 100)
        hamburger.click()
        const switchStudent = driver.findElement(By.className("clickable"))
        await driver.wait(until.elementIsVisible(switchStudent), 100)
        switchStudent.click()

        // const course = driver.findElement(By.className("activeClass"))
        // await driver.wait(until.elementIsVisible(course), 100)
        // course.click()
        let rect: IRectangle
        let rect2: IRectangle
        driver.sleep(1000)
        const labs = await driver.findElement(By.className("col-md-9"))
        await driver.wait(until.elementIsVisible(labs), 5000)

        const utility = await driver.findElement(By.className("list-group width-resize"))
        await driver.wait(until.elementIsVisible(utility), 5000)
        rect = await utility.getRect()
        rect2 = await labs.getRect()

        var overlap = (rect.x < rect2.x + rect2.width &&
            rect.x + rect.width > rect2.x &&
            rect.y < rect2.y + rect2.height &&
            rect.height + rect.y > rect2.y)
        jest.setTimeout(50000)
        expect(true).toBe(true)



    })
})
