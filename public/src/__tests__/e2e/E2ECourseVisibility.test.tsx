import { By, until } from 'selenium-webdriver'
import { setupDrivers } from '../TestHelpers'


describe("End to End course visibility", () => {

    const drivers = setupDrivers()

    drivers.forEach(driver => {
        it("Should not overlap on res 1920x1080", async () => {
            // Go to course page
            const hamburger = await driver.findElement(By.className("hamburger"))
            await driver.wait(until.elementIsVisible(hamburger), 100)
            await hamburger.click()
            const goToCourse = await driver.findElement(By.css(".courseLink"))
            await driver.wait(until.elementIsVisible(goToCourse), 100)
            await goToCourse.click()

            // Find course code from course card and hamburger menu
            const card = await driver.findElement(By.css(".card"))
            const cardCourseCode = await card.findElement(By.css(".card-header")).getText()
            const courseCode = cardCourseCode.split("\n")[0]
            const hamburgerCode = await driver.findElement(By.css("#title")).getText()

            // Unfavorite the course
            const star = await driver.findElement(By.css(".fa-star"))
            await driver.wait(until.elementIsVisible(star), 100)
            await star.click()

            // Get courseCode from myCourses section
            const myCourse = await driver.findElement(By.css(".myCourses"))
            const myCoursesCard = await myCourse.findElement(By.css(".card-header")).getText()

            const hasMoved = (courseCode === myCoursesCard.split("\n")[0])

            // Find course codes in navigator
            const navigatorCourses = await driver.findElement(By.css(".navigator"))
            const courses = await navigatorCourses.findElements(By.css("#title"))

            // Check if the course has moved from navigator
            let isInList = true
            for (let i = 0; i < courses.length; i++) {
                if ((await courses[i].getText()) === hamburgerCode) {
                    isInList = false
                }
            }
            const movedFromFavorites = (isInList && hasMoved)

            await driver.close()
            expect(movedFromFavorites).toBe((true))
        }, 50000)
    })
})
