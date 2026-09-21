import { render, screen } from "@testing-library/react"
import AboutPage from "../pages/AboutPage"

describe("AboutPage", () => {
    // The feature blocks used to be screenshots of the light UI, which glared on
    // a dark page. They now render the application's own components, so they
    // follow the active daisyUI theme. A preview that regressed to an <img>
    // would bring the theme mismatch back.
    test("feature previews are rendered markup, not screenshots", () => {
        const { container } = render(<AboutPage />)

        const screenshots = Array.from(container.querySelectorAll("img"))
            .map((img) => img.getAttribute("src") ?? "")

        // What is left is the two flat webp illustrations in the mini feature
        // blocks and the GitHub screenshot, which is not QuickFeed's own UI.
        expect(screenshots).toEqual([
            "/assets/img/overlapping-arrows-no-background.webp",
            "/assets/img/Aplus2-no-background.webp",
            "/assets/img/intro3.png",
        ])
    })

    test("previews use theme tokens rather than baked-in colors", () => {
        const { container } = render(<AboutPage />)

        // The table headers are bg-base-300, which daisyUI resolves to a dark
        // shade under a dark theme and a light one under a light theme.
        const headers = container.querySelectorAll("thead.bg-base-300")
        expect(headers.length).toBeGreaterThan(0)

        for (const element of container.querySelectorAll("[style]")) {
            // Progress bar widths are the only inline styles we expect; a
            // hard-coded color here would not follow the theme.
            expect(element.getAttribute("style")).not.toMatch(/color|background/)
        }
    })

    // The previews illustrate the UI; their buttons and sortable headers are not
    // wired to anything, so they must stay out of the tab order.
    test("previews are inert and described for assistive technology", () => {
        const { container } = render(<AboutPage />)

        const previews = container.querySelectorAll('[role="img"]')
        expect(previews).toHaveLength(3)
        for (const preview of previews) {
            expect(preview.getAttribute("aria-label")).toBeTruthy()
            // inert hides its subtree from assistive technology, so it must sit
            // below the described element, not on it.
            expect(preview.hasAttribute("inert")).toBe(false)
            expect(preview.querySelector("[inert]")).not.toBeNull()
        }
    })

    test("renders the sample lab result through the real score table", () => {
        render(<AboutPage />)

        expect(screen.getByText("TestGitQuestions")).toBeDefined()
        expect(screen.getByText("Total Score")).toBeDefined()
    })
})
