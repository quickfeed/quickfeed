import { Builder, By, Capabilities } from 'selenium-webdriver'
import { isOverlapping } from '../testHelpers/testHelpers'

describe("Front page elements login and logo should not overlap", () => {
    //Laptop, desktop, mobile
    const overlapTests: { width: number, height: number, want: boolean }[] = [
        // Insert which resolutions to test here
        { width: 960, height: 1080, want: false },//desktop
        { width: 1366, height: 768, want: false },//Laptop
        { width: 960, height: 1080, want: false },//split screen desktop
        { width: 683, height: 768, want: false } // split screen laptop
    ]
    overlapTests.forEach(test => {
        it(`Should not overlap on res ${test.width}x${test.height}`, async () => {
            //This test requires you to set
            const firefox = require('selenium-webdriver/firefox')
            const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
            const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
                .set("acceptInsecureCerts", true)).build()
            await driver.get("https://127.0.0.1/dev")
            await driver.manage().window().setRect({ width: test.width, height: test.height })

            const logo = await driver.findElement(By.className("navbar-brand"))
            const signIn = await driver.findElement(By.className("signIn"))
            let rect = await logo.getRect()
            let rect2 = await signIn.getRect()

            var overlap = isOverlapping(rect, rect2)

            jest.setTimeout(50000)
            expect(overlap).toBe(test.want)
        }, 50000)
    })

})
