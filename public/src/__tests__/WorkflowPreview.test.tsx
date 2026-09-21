import { act, render } from "@testing-library/react"
import { afterEach, beforeEach, vi } from "vitest"
import WorkflowPreview from "../components/about/WorkflowPreview"

const mockReducedMotion = (matches: boolean) => {
    vi.stubGlobal("matchMedia", (query: string) => ({
        matches,
        media: query,
        addEventListener: () => { /* the preference does not change during a test */ },
        removeEventListener: () => { /* nothing to remove */ },
    }))
}

// The connectors between the steps are decorative list items; the steps
// themselves are the ones left in the accessibility tree.
const markers = (container: HTMLElement): Element[] =>
    Array.from(container.querySelectorAll('ol > li:not([aria-hidden="true"]) > span.rounded-full'))

const marker = (container: HTMLElement, step: number): string => markers(container)[step].className

// Every marker is a filled bubble except the feedback step, which is a dial.
const iconMarkers = (container: HTMLElement): string[] =>
    markers(container).filter(element => !element.querySelector("svg")).map(element => element.className)

// The running total sits on the dial that replaces the feedback step's icon.
const dialScore = (container: HTMLElement): string | null =>
    container.querySelector(".tabular-nums")?.textContent ?? null

// The second circle of the dial is the arc; the first is its track.
const dialArc = (container: HTMLElement): string =>
    container.querySelector("svg circle + circle")?.getAttribute("class") ?? ""

const floatingGain = (container: HTMLElement): string | null =>
    container.querySelector(".animate-float-up")?.textContent ?? null

describe("WorkflowPreview", () => {
    beforeEach(() => {
        vi.useFakeTimers()
    })

    afterEach(() => {
        vi.useRealTimers()
        vi.unstubAllGlobals()
    })

    // The student pushes three times; only the last push clears the score limit,
    // and the teacher grades once it does.
    test("replays pushes until the score passes, then grades", () => {
        mockReducedMotion(false)
        const { container } = render(<WorkflowPreview />)

        // First push: only the first step is lit and nothing has been scored.
        expect(marker(container, 0)).toContain("scale-110")
        expect(marker(container, 1)).toContain("bg-base-300")
        expect(dialScore(container)).toBe("0%")
        expect(floatingGain(container)).toBeNull()

        // Building: the gears step takes over, and there is still nothing to show.
        advance()
        expect(marker(container, 1)).toContain("scale-110")
        expect(floatingGain(container)).toBeNull()

        // Feedback: the score the push earned floats up, and the dial follows it.
        advance()
        expect(floatingGain(container)).toBe("+33")
        expect(dialScore(container)).toBe("33%")
        expect(dialArc(container)).toContain("stroke-primary")
        expect(marker(container, 3)).toContain("bg-base-300")

        // The second push returns to the first step, carrying the old score and
        // clearing the floating number.
        advance()
        expect(marker(container, 0)).toContain("scale-110")
        expect(marker(container, 1)).not.toContain("scale-110")
        expect(dialScore(container)).toBe("33%")
        expect(floatingGain(container)).toBeNull()

        // Run out the rest of the second attempt and all of the third.
        advance()
        advance()
        expect(floatingGain(container)).toBe("+34")
        advance()
        advance()
        advance()
        expect(floatingGain(container)).toBe("+25")
        expect(dialScore(container)).toBe("92%")

        // The passing score turns the dial green and brings in the teacher, whose
        // approval mark lands.
        expect(dialArc(container)).toContain("stroke-success")
        advance()
        for (const className of iconMarkers(container)) {
            expect(className).toContain("bg-success")
        }
        expect(container.querySelector(".animate-pop-in")).not.toBeNull()

        // The whole sequence loops back to the first push.
        advance()
        expect(dialScore(container)).toBe("0%")
        expect(container.querySelector(".animate-pop-in")).toBeNull()
    })

    // A viewer who has asked for reduced motion gets no timers and no keyframes,
    // and sees the outcome rather than a workflow frozen on its first step.
    test("rests on the graded result when motion is reduced", () => {
        mockReducedMotion(true)
        const { container } = render(<WorkflowPreview />)

        expect(dialScore(container)).toBe("92%")
        expect(dialArc(container)).toContain("stroke-success")
        for (const className of iconMarkers(container)) {
            expect(className).toContain("bg-success")
        }
        expect(vi.getTimerCount()).toBe(0)

        // The approval mark is there, but nothing animates in.
        expect(container.querySelector(".fa-check")).not.toBeNull()
        expect(container.querySelector(".animate-pop-in")).toBeNull()
        expect(container.querySelector(".animate-float-up")).toBeNull()
        expect(container.querySelector(".fa-spin")).toBeNull()

        advance()
        expect(dialScore(container)).toBe("92%")
    })
})

const advance = () => act(() => {
    vi.advanceTimersByTime(4000)
})
