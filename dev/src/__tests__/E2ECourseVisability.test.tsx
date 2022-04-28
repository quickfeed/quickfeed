import { Builder, By, Capabilities, until } from 'selenium-webdriver'


describe("End to End course visability", () => {
    it("Should not overlap on res 1920x1080", async () => {
        const firefox = require('selenium-webdriver/firefox')
        const service = new firefox.ServiceBuilder('drivers\\geckodriver.exe')
        const driver = new Builder().forBrowser('firefox').setFirefoxService(service).withCapabilities(Capabilities.firefox()
            .set("acceptInsecureCerts", true)).build()
        await driver.get("https://127.0.0.1/dev/")

        //Go to ccoursepage
        const hamburger = await driver.findElement(By.className("hamburger"))
        await driver.wait(until.elementIsVisible(hamburger), 100)
        hamburger.click()
        await driver.sleep(1000)
        const goToCourse = await driver.findElement(By.css(".courseLink"))
        await driver.wait(until.elementIsVisible(goToCourse), 100)
        goToCourse.click()
        await driver.sleep(1000)

        //Find course code from coursecard and hamburger menu
        const card = await driver.findElement(By.css(".card"))
        const cardCourseCode = await card.findElement(By.css(".card-header")).getText()
        const courseCode = cardCourseCode.split("\n")[0]
        await driver.sleep(1000)
        const hamburgerCode = await driver.findElement(By.css("#title")).getText()

        //Unfavorite the course
        await driver.sleep(1000)
        const star = await driver.findElement(By.css(".fa-star"))
        await driver.wait(until.elementIsVisible(star), 100)
        star.click()

        //Get courseCode from myCourses section
        await driver.sleep(1000)
        const myCourse = await driver.findElement(By.css(".myCourses"))
        const myCoursesCard = await myCourse.findElement(By.css(".card-header")).getText()

        const hasMoved = (courseCode === myCoursesCard.split("\n")[0])

        //Find coursecodes in navigator
        await driver.sleep(1000)
        const navigatorCourses = await driver.findElement(By.css(".navigator"))
        await driver.sleep(1000)
        const courses = await navigatorCourses.findElements(By.css("#title"))

        //Check if the course has moved from navigator
        var isInList = true
        for (let i = 0; i < courses.length; i++) {
            if (await courses[i].getText() === hamburgerCode) {
                isInList = false
            }
        }
        const movedFromFavorites = (isInList && hasMoved)

        await driver.close()

        expect(movedFromFavorites).toBe((true))
    }, 50000)
})
