import { Builder, By, Capabilities, until } from 'selenium-webdriver'
import { isOverlapping } from '../testHelpers/testHelpers'




describe("Course utility elements should not overlap", () => {
    const overlapTests: { width: number, height: number, want: boolean }[] = [
        // Insert which resolutions to test here
        { width: 960, height: 1080, want: false },//desktop
        { width: 360, height: 740, want: false },//More common phones
        { width: 360, height: 640, want: false },//Older phones
        { width: 412, height: 914, want: false } //Bigger phones


    ]
    overlapTests.forEach(test => {

        it(`Should not overlap on split screen ${test.width}x${test.height}`, async () => {
            const firefox = require('selenium-webdriver/firefox')
            const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
            const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
                .set("acceptInsecureCerts", true)).build()
            await driver.get("https://127.0.0.1/dev/#/course/1")
            await driver.manage().window().setRect({ width: test.width, height: test.height })

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

            expect(overlap).toBe(test.want)
        }, 50000)
    })

})
