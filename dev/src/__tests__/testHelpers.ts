import { IRectangle } from "selenium-webdriver"

export const isOverlapping = (rect: IRectangle, rect2: IRectangle) => {
    var overlap = (rect.x < rect2.x + rect2.width &&
        rect.x + rect.width > rect2.x &&
        rect.y < rect2.y + rect2.height &&
        rect.height + rect.y > rect2.y)
    if (overlap) {
        return true //Elements are overlapping
    }
    return false //Elements are not overlapping
}
