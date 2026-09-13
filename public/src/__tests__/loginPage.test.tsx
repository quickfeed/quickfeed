import { render, screen } from "@testing-library/react"
import { Provider } from "overmind-react"
import { MemoryRouter } from "react-router"
import LoginPage from "../pages/LoginPage"
import { initializeOvermind } from "./TestHelpers"

describe("LoginPage", () => {
    test("explains what access the GitHub app gets before the user signs in", () => {
        render(
            <Provider value={initializeOvermind({})}>
                <MemoryRouter>
                    <LoginPage />
                </MemoryRouter>
            </Provider>
        )

        expect(screen.getByRole("link", { name: "Sign in" })).toBeTruthy()
        expect(screen.getByText(/can only access repositories associated with the courses you have signed up for/)).toBeTruthy()
        expect(screen.getByText(/do not use a GitHub account that you also use for work/)).toBeTruthy()
        const learnMore = screen.getByRole("link", { name: "Learn more" })
        expect(learnMore.getAttribute("href")).toBe("https://docs.github.com/en/apps/using-github-apps/authorizing-github-apps")
    })
})
