import { By } from 'selenium-webdriver'
import { isOverlapping, setupDrivers } from '../TestHelpers'


describe("Front page elements login and logo should not overlap", () => {

    const overlapTests: { width: number, height: number, want: boolean }[] = [
        // Resolutions to test
        { width: 1920, height: 1080, want: false }, // Desktop
        { width: 1366, height: 768, want: false },  // Laptop
        { width: 960, height: 1080, want: false },  // Split screen desktop
        { width: 683, height: 768, want: false }    // Split screen laptop
    ]

    const drivers = setupDrivers()

    drivers.forEach(driver => {
        overlapTests.forEach(test => {
            it(`Should not overlap on res ${test.width}x${test.height}`, async () => {
                await driver.manage().window().setRect({ width: test.width, height: test.height })

                const logo = await driver.findElement(By.className("navbar-brand"))
                const avatar = await driver.findElement(By.id("avatar"))
                const rect = await logo.getRect()
                const rect2 = await avatar.getRect()

                const overlap = isOverlapping(rect, rect2)

                jest.setTimeout(50000)
                expect(overlap).toBe(test.want)
            }, 50000)
        })
    })
})
