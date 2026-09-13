import { render } from "@testing-library/react"
import AboutPage from "../pages/AboutPage"

describe("AboutPage", () => {
    // The intro*.png screenshots were captured from the light UI. The
    // .about-screenshot class is what tailwind.css hooks the dark-theme
    // filter onto, so losing it would leave a white glare on a dark page.
    test("screenshots carry the dark-theme treatment class", () => {
        const { container } = render(<AboutPage />)

        const screenshots = Array.from(container.querySelectorAll("img"))
            .filter((img) => img.getAttribute("src")?.endsWith(".png"))

        expect(screenshots).toHaveLength(4)
        for (const img of screenshots) {
            expect(img.className).toContain("about-screenshot")
            // Defines the screenshot edge against both grounds.
            expect(img.className).toContain("border-base-300")
            expect(img.className).toContain("rounded-lg")
        }
    })

    // The two webp illustrations are flat, saturated artwork on a transparent
    // ground and stay legible on a dark page, so they are deliberately left
    // untouched; inverting them would turn red into cyan.
    test("illustrations are not filtered", () => {
        const { container } = render(<AboutPage />)

        const illustrations = Array.from(container.querySelectorAll("img"))
            .filter((img) => img.getAttribute("src")?.endsWith(".webp"))

        expect(illustrations).toHaveLength(2)
        for (const img of illustrations) {
            expect(img.className).not.toContain("about-screenshot")
        }
    })
})
