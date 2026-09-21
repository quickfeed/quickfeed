import { render, screen } from "@testing-library/react"
import { AssignmentsRepo, InfoRepo, TestsRepo, studentRepoName } from "../Helpers"
import AboutPage from "../pages/AboutPage"

describe("AboutPage", () => {
    // The page used to carry screenshots of the light UI and flat clipart in
    // fixed colors, both of which ignored the theme. Everything is markup now,
    // so an <img> reappearing here would bring the theme mismatch back.
    test("the page renders no images", () => {
        const { container } = render(<AboutPage />)

        expect(container.querySelectorAll("img")).toHaveLength(0)
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
        expect(previews).toHaveLength(4)
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

    test("summarizes the workflow as a labelled pipeline", () => {
        const { container } = render(<AboutPage />)

        // The connectors between the steps are decorative; the labelled steps
        // are what carries the workflow for a reader. The label is the last child
        // of a step, after its marker and any number floating above it.
        const labels = Array.from(container.querySelectorAll('ol > li:not([aria-hidden="true"]) > span:last-child'))
        expect(labels.map(label => label.textContent)).toEqual([
            "Student pushes code",
            "QuickFeed builds and tests",
            "Student sees feedback",
            "Teacher grades",
        ])
    })

    // The About page used to show a 2022 screenshot of a push to a repository
    // named "labs" in the "autograde-test" organization. QuickFeed has named
    // neither that way for years, so the repository layout is now derived from
    // the same constants the rest of the frontend uses.
    test("names repositories the way QuickFeed creates them", () => {
        render(<AboutPage />)

        expect(screen.getByText(InfoRepo)).toBeDefined()
        expect(screen.getByText(AssignmentsRepo)).toBeDefined()
        expect(screen.getByText(TestsRepo)).toBeDefined()
        expect(screen.getByText(studentRepoName("hfurubotten"))).toBeDefined()
    })
})
