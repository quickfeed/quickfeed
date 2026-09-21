import { cleanup, fireEvent, render, screen } from "@testing-library/react"
import { afterEach, describe, expect, it } from "vitest"
import LogOutput from "../components/LogOutput"

afterEach(cleanup)

describe("LogOutput", () => {
    it("starts closed, since each test's output is shown with that test", () => {
        render(<LogOutput>the rest of the run</LogOutput>)
        expect(screen.queryByText("the rest of the run")).toBeNull()

        fireEvent.click(screen.getByRole("button", { name: /Show/ }))
        expect(screen.getByText("the rest of the run")).toBeDefined()
    })

    it("starts open when it is all there is to read", () => {
        // A run that failed to build attributed nothing to a test, so the build
        // log holds the compiler's diagnostic and the student needs it at once.
        render(<LogOutput defaultOpen>the submitted code did not compile</LogOutput>)
        expect(screen.getByText("the submitted code did not compile")).toBeDefined()
    })
})
