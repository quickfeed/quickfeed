import { act, render } from "@testing-library/react"
import { afterEach, beforeEach, vi } from "vitest"
import WorkflowPreview from "../components/about/WorkflowPreview"
import { previewAttempts, previewScoreLimit } from "../components/about/previewData"

const mockReducedMotion = (matches: boolean) => {
    vi.stubGlobal("matchMedia", (query: string) => ({
        matches,
        media: query,
        addEventListener: () => { /* the preference does not change during a test */ },
        removeEventListener: () => { /* nothing to remove */ },
    }))
}

// The running total sits on the dial that replaces the feedback step's icon.
const score = (container: HTMLElement): number =>
    Number(container.querySelector(".tabular-nums")?.textContent?.replace("%", ""))

const gain = (container: HTMLElement): number | null => {
    const floating = container.querySelector(".animate-float-up")?.textContent
    return floating ? Number(floating.replace("+", "")) : null
}

const isApproved = (container: HTMLElement): boolean =>
    container.querySelector(".fa-check") !== null

// Each push walks three steps, and the grading step closes the cycle.
const stagesPerCycle = previewAttempts.length * 3 + 1

describe("WorkflowPreview", () => {
    beforeEach(() => {
        vi.useFakeTimers()
    })

    afterEach(() => {
        vi.useRealTimers()
        vi.unstubAllGlobals()
    })

    test("walks every push and loops back to the start", () => {
        mockReducedMotion(false)
        const { container } = render(<WorkflowPreview />)

        const scores: number[] = []
        const gains: number[] = []
        let approvedAt = -1
        for (let stage = 0; stage < stagesPerCycle; stage++) {
            scores.push(score(container))
            const floating = gain(container)
            if (floating !== null) {
                gains.push(floating)
            }
            if (isApproved(container)) {
                approvedAt = stage
            }
            advance()
        }

        // Each number that floats up is the step the dial then takes, so the
        // gains and the totals cannot drift apart.
        expect(gains).toEqual(previewAttempts.map((total, i) => total - (previewAttempts[i - 1] ?? 0)))
        expect([...new Set(scores)]).toEqual([0, ...previewAttempts])

        // The teacher arrives last, and only once a push clears the limit.
        expect(approvedAt).toBe(stagesPerCycle - 1)
        expect(scores[approvedAt]).toBeGreaterThanOrEqual(previewScoreLimit)

        // The cycle starts over rather than running off the end of the stages.
        expect(score(container)).toBe(0)
    })

    // Reduced motion is a stated preference, and the page is the first thing a
    // visitor sees. Resting on the outcome beats freezing on the first push.
    test("schedules nothing when motion is reduced", () => {
        mockReducedMotion(true)
        const { container } = render(<WorkflowPreview />)

        expect(vi.getTimerCount()).toBe(0)
        expect(score(container)).toBe(previewAttempts[previewAttempts.length - 1])
        expect(isApproved(container)).toBe(true)
        expect(container.querySelector(".animate-pop-in")).toBeNull()
        expect(container.querySelector(".animate-float-up")).toBeNull()
    })
})

const advance = () => act(() => {
    vi.advanceTimersByTime(4000)
})
