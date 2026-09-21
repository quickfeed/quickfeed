import { render } from "@testing-library/react"
import AboutPage from "../pages/AboutPage"

describe("AboutPage", () => {
    // LoginPage shows the About page to visitors who are not logged in, and the
    // previews render real components. One of them reaching for course state
    // would crash the login page, so this renders without a Provider on purpose.
    test("renders the previews without application state", () => {
        const { container } = render(<AboutPage />)

        // The score and results tables come from the sample data; empty bodies
        // would mean it stopped reaching them.
        expect(container.querySelectorAll("tbody tr").length).toBeGreaterThan(0)
    })

    // The previews are illustrations rather than working UI, so their buttons and
    // sortable headers are inert. Since inert also hides its subtree from
    // assistive technology, the description has to sit above it, not on it.
    test("each preview is described above its inert content", () => {
        const { container } = render(<AboutPage />)

        const previews = container.querySelectorAll('[role="img"]')
        expect(previews.length).toBeGreaterThan(0)
        for (const preview of previews) {
            expect(preview.getAttribute("aria-label")).toBeTruthy()
            expect(preview.hasAttribute("inert")).toBe(false)
            expect(preview.querySelector("[inert]")).not.toBeNull()
        }
    })
})
